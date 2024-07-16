package biz

import (
	"context"
	"crypto/tls"
	"github.com/panjf2000/ants/v2"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/internal/pkg/enums"
	"github.com/peckfly/gopeck/pkg/atomicx"
	"github.com/peckfly/gopeck/pkg/interpreter"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"github.com/peckfly/gopeck/pkg/netx"
	"github.com/peckfly/gopeck/pkg/registry"
	"go.uber.org/zap"
	"golang.org/x/net/http2"
	"io"
	"math/rand"
	"net/http"
	"net/http/httptrace"
	"net/url"
	"strings"
	"sync"
	"time"
)

var (
	// control the stop signal of all request
	stops     = make(map[uint64]*atomicx.AtomicBool)
	startTime = time.Now()
)

type (
	RequesterUsecase struct {
		conf           conf.WorkerStressConf
		queRepository  repo.QueRepository
		nodeRepository repo.NodeRepository
		discovery      registry.Discovery
	}

	Requester struct {
		PlanId              uint64
		TaskId              uint64
		StressType          int32
		StressMode          int32
		Num                 int32
		StepIntervalTime    int32
		Nums                []int32
		MaxConnections      int32
		MaxIdleConnections  int32
		StressTime          int32
		Timeout             int32
		Method              string
		Url                 string
		Headers             map[string]string
		Query               string
		Body                string
		DynamicParams       []*DynamicParam
		ResponseCheckScript string

		DisableKeepAlive bool
		H2               bool
		MaxBodySize      int64

		DisableCompression bool
		DisableRedirects   bool
		Proxy              string

		Addr string

		responseChecker *interpreter.EvalInterpreter
		// Writer is where results will be written. If nil, results are written to stdout.
		Writer  io.Writer
		results chan *repo.Result
		start   time.Duration

		done chan bool

		poolFunc *ants.PoolWithFunc

		StartTime int64
	}
	DynamicParam struct {
		Headers map[string]string
		Query   map[string]string
		Body    string
	}
)

func NewRequesterUsecase(conf *conf.ServerConf, queRepository repo.QueRepository, nodeRepository repo.NodeRepository, discovery registry.Discovery) *RequesterUsecase {
	return &RequesterUsecase{
		conf:           conf.StressConf,
		queRepository:  queRepository,
		nodeRepository: nodeRepository,
		discovery:      discovery,
	}
}

func (b *RequesterUsecase) Request(ctx context.Context, r *Requester) error {
	logc.Info(ctx, "start request", zap.Uint64("task_id", r.TaskId))
	err := r.init(&b.conf)
	if err != nil {
		return err
	}
	go b.request(r)
	return nil
}

func (b *Requester) init(conf *conf.WorkerStressConf) error {
	maxNum := b.Num
	if b.StressMode == int32(enums.Step) {
		maxNum = max(b.Nums[len(b.Nums)-1], maxNum)
	}
	if enums.StressType(b.StressType) == enums.Rps {
		b.results = make(chan *repo.Result, min(maxNum*conf.RpsResultChanBlowup, conf.MaxResultChanSize))
	} else {
		b.results = make(chan *repo.Result, min(maxNum, conf.MaxResultChanSize))
	}
	b.done = make(chan bool, 1)
	stops[b.TaskId] = atomicx.ForAtomicBool(false)
	if len(b.ResponseCheckScript) > 0 {
		evalInterpreter, err := interpreter.NewEvalInterpreter(b.ResponseCheckScript)
		if err != nil {
			return err
		}
		b.responseChecker = evalInterpreter
	}
	return nil
}

func (b *RequesterUsecase) request(r *Requester) {
	ctx := context.Background()
	r.start = now()
	err := b.report(ctx, r)
	if err != nil {
		return
	}
	b.run(ctx, r)
	b.done(ctx, r)
}

func (b *RequesterUsecase) run(ctx context.Context, r *Requester) {
	if r.MaxConnections == 0 {
		r.MaxConnections = int32(b.conf.DefaultMaxConnections)
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
		MaxConnsPerHost:     int(r.MaxConnections),
		MaxIdleConnsPerHost: int(r.MaxIdleConnections),
		DisableKeepAlives:   r.DisableKeepAlive,
		DisableCompression:  r.DisableCompression,
	}
	if len(r.Proxy) > 0 {
		if proxyUrl, err := url.Parse(r.Proxy); err == nil {
			tr.Proxy = http.ProxyURL(proxyUrl)
		}
	}
	if r.H2 {
		http2.ConfigureTransport(tr)
	} else {
		tr.ForceAttemptHTTP2 = false
		tr.TLSNextProto = make(map[string]func(string, *tls.Conn) http.RoundTripper)
	}
	client := &http.Client{Transport: tr, Timeout: time.Duration(r.Timeout) * time.Second}
	if r.DisableRedirects {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	r.StartTime = time.Now().Unix()
	if r.StressType == int32(enums.Rps) {
		if r.StressMode == int32(enums.Step) {
			b.runStepRpsRequest(ctx, client, r)
		} else {
			b.runRpsRequest(ctx, client, r)
		}
	} else if r.StressType == int32(enums.Concurrency) {
		if r.StressMode == int32(enums.Step) {
			b.runStepConcurrencyRequest(ctx, client, r)
		} else {
			b.runConcurrencyRequest(ctx, client, r)
		}
	}
}

func (b *RequesterUsecase) runRpsRequest(ctx context.Context, client *http.Client, r *Requester) {
	pacer := ConstantPacer{int(r.Num), time.Second}
	began, count := time.Now(), uint64(0)
	taskChan := make(chan struct{})
	var wg sync.WaitGroup
	du := time.Duration(r.StressTime) * time.Second
	costGoroutineNums := 1
	wg.Add(1)
	go b.requestFromChan(ctx, taskChan, &wg, client, r)
	for {
		elapsed := time.Since(began)
		if elapsed > du {
			break
		}
		wait, stop := pacer.Pace(elapsed, count)
		if stop {
			break
		}
		time.Sleep(wait)
		if stops[r.TaskId].True() {
			logc.Info(ctx, "stop request got signal", zap.Uint64("task_id", r.TaskId))
			break
		}
		select {
		case taskChan <- struct{}{}:
			count++
			continue
		default:
			wg.Add(1)
			costGoroutineNums++
			go b.requestFromChan(ctx, taskChan, &wg, client, r)
		}
		select {
		case taskChan <- struct{}{}:
			count++
		}
	}
	close(taskChan)
	wg.Wait()
	elapsed := time.Since(began)
	logc.Info(ctx, "run request rps mode cost:", zap.Uint64("task_id", r.TaskId), zap.Duration("cost", elapsed), zap.Int("cost_goroutine_num", costGoroutineNums))
}

func (b *RequesterUsecase) runConcurrencyRequest(ctx context.Context, client *http.Client, r *Requester) {
	var wg sync.WaitGroup
	began := time.Now()
	du := time.Duration(r.StressTime) * time.Second
	for i := 0; i < int(r.Num); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				elapsed := time.Since(began)
				if elapsed > du {
					break
				}
				if stops[r.TaskId].True() {
					logc.Info(ctx, "stop request got signal", zap.Uint64("task_id", r.TaskId))
					break
				}
				b.goRequest(ctx, client, r)
			}
		}()
	}
	wg.Wait()
	elapsed := time.Since(began)
	logc.Info(ctx, "run request concurrency mode cost:", zap.Uint64("task_id", r.TaskId), zap.Duration("cost", elapsed))
}

func (b *RequesterUsecase) requestFromChan(ctx context.Context, taskChan <-chan struct{}, wg *sync.WaitGroup, client *http.Client, r *Requester) {
	defer wg.Done()
	for range taskChan {
		b.goRequest(ctx, client, r)
	}
}

func (b *RequesterUsecase) goRequest(ctx context.Context, client *http.Client, r *Requester) {
	s := now()
	var responseContentLength int64
	var code int
	var dnsStart, connStart, resStart, reqStart, delayStart time.Duration
	var dnsDuration, connDuration, resDuration, reqDuration, delayDuration time.Duration
	var req *http.Request
	var rndIndex int
	isDynamic := r.DynamicParams != nil && len(r.DynamicParams) > 0
	var err error
	if isDynamic {
		rndIndex = rand.Intn(len(r.DynamicParams))
		req, err = constructRequest(r.Method, r.Url, r.DynamicParams[rndIndex].Headers, r.DynamicParams[rndIndex].Query, r.DynamicParams[rndIndex].Body)
	} else {
		req, err = constructRequest(r.Method, r.Url, r.Headers, netx.ParseQuery(r.Query), r.Body)
	}
	if err != nil {
		logc.Error(ctx, "failed to construct request", zap.Error(err))
		return
	}
	trace := &httptrace.ClientTrace{
		DNSStart: func(info httptrace.DNSStartInfo) {
			dnsStart = now()
		},
		DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
			dnsDuration = now() - dnsStart
		},
		GetConn: func(h string) {
			connStart = now()
		},
		GotConn: func(connInfo httptrace.GotConnInfo) {
			if !connInfo.Reused {
				connDuration = now() - connStart
			}
			reqStart = now()
		},
		WroteRequest: func(w httptrace.WroteRequestInfo) {
			reqDuration = now() - reqStart
			delayStart = now()
		},
		GotFirstResponseByte: func() {
			delayDuration = now() - delayStart
			resStart = now()
		},
	}
	req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
	requestTime := time.Now().Unix()
	resp, err := client.Do(req)
	var errorStr string
	var respBody []byte
	if err == nil {
		responseContentLength = resp.ContentLength
		code = resp.StatusCode
		body := io.Reader(resp.Body)
		// TODO read body cost time and lose performance
		if r.MaxBodySize > 0 {
			body = io.LimitReader(body, r.MaxBodySize)
		}
		respBody, err = io.ReadAll(body)
		if err != nil {
			errorStr = b.getErrorWithCutLength(err)
		}
		resp.Body.Close()
	} else {
		errorStr = b.getErrorWithCutLength(err)
	}
	var bodyResult string
	if r.responseChecker != nil {
		err = r.responseChecker.ExecuteScript(func(executor any) {
			bodyResult = executor.(func(string) string)(string(respBody))
		})
		if err != nil {
			logc.Error(ctx, "failed to execute script", zap.Error(err))
		}
	}
	t := now()
	resDuration = t - resStart
	finish := t - s
	if finish > time.Duration(b.conf.MaxTimeoutSecond)*time.Second {
		finish = time.Duration(b.conf.MaxTimeoutSecond) * time.Second
	}
	_, _, _, _, _ = dnsDuration, connDuration, reqDuration, delayDuration, resDuration
	r.results <- &repo.Result{
		Err:                   errorStr,
		StatusCode:            code,
		Duration:              finish,
		ResponseContentLength: responseContentLength,
		TimeStamp:             requestTime,
		BodyCheckResult:       bodyResult,
		Stop:                  stops[r.TaskId].True(),
	}
}

func (b *RequesterUsecase) getErrorWithCutLength(err error) string {
	if err == nil {
		return ""
	}
	errorStr := err.Error()
	if len(errorStr) > b.conf.ErrorCutLength {
		errorStr = errorStr[len(errorStr)-b.conf.ErrorCutLength:]
	}
	return errorStr
}

func constructRequest(method string, urlStr string, headers map[string]string, query map[string]string, body string) (*http.Request, error) {
	reqURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	if len(query) > 0 {
		q := reqURL.Query()
		for k, v := range query {
			q.Add(k, v)
		}
		reqURL.RawQuery = q.Encode()
	}

	requestBody := strings.NewReader(body)

	req, err := http.NewRequest(method, reqURL.String(), requestBody)
	if err != nil {
		return nil, err
	}

	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return req, nil
}

func (b *RequesterUsecase) done(ctx context.Context, r *Requester) {
	close(r.results)
	total := now() - r.start
	<-r.done
	r.poolFunc.Release()
	delete(stops, r.TaskId)
	costNum := int(r.Num)
	if r.StressMode == int32(enums.Step) && len(r.Nums) > 0 {
		costNum = int(r.Nums[len(r.Nums)-1])
	}
	err := b.nodeRepository.UpdateNodeCostNum(ctx, r.Addr, int(r.StressType), costNum)
	if err != nil {
		logc.Error(ctx, "update node cost num failed", zap.Error(err))
	}
	logc.Info(ctx, "done total cost ", zap.Duration("total", total))
}

func (b *RequesterUsecase) Stop(ctx context.Context, planId uint64, taskId uint64) error {
	logc.Info(ctx, "stop request info", zap.Uint64("plan_id", planId), zap.Uint64("task_id", taskId))
	stops[taskId].CompareAndSwap(false, true)
	return nil
}

func now() time.Duration { return time.Since(startTime) }
