package biz

import (
	"context"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"github.com/peckfly/gopeck/pkg/numx"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

func (b *RequesterUsecase) runStepRpsRequest(ctx context.Context, client *http.Client, r *Requester) {
	logc.Info(ctx, "start run step rps request", zap.Uint64("task_id", r.TaskId), zap.Int32("stress_time", r.StressTime))
	pacer := ConstantPacer{int(r.Nums[0]), time.Second}
	began, count := time.Now(), uint64(0)
	start := time.Now()
	taskChan := make(chan struct{})
	var wg sync.WaitGroup
	du := time.Duration(r.StressTime) * time.Second
	costGoroutineNums := 1
	wg.Add(1)
	go b.requestFromChan(ctx, taskChan, &wg, client, r)
	intervalsLen := len(r.Nums)
	cm := make([]bool, intervalsLen)
	cm[0] = true
	for {
		elapsed := time.Since(began)
		wait, stop := pacer.Pace(elapsed, count)
		if stop {
			break
		}
		cost := time.Since(start)
		if cost > du {
			break
		}
		seconds := cost.Seconds()
		interval := min(int(intervalsLen)-1, int(seconds)/int(r.StepIntervalTime))
		if !cm[interval] {
			pacer.Freq = int(r.Nums[interval])
			count = 0
			began = time.Now()
			cm[interval] = true
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

func (b *RequesterUsecase) runStepConcurrencyRequest(ctx context.Context, client *http.Client, r *Requester) {
	var wg sync.WaitGroup
	began := time.Now()
	du := time.Duration(r.StressTime) * time.Second
	intervalsLen := numx.CeilDiv(r.StressTime, r.StepIntervalTime)
	run := func(num int) {
		for i := 0; i < num; i++ {
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
	}
	preNum := 0
	for i := 0; i < int(intervalsLen); i++ {
		run(int(r.Nums[i]) - preNum)
		time.Sleep(time.Duration(r.StepIntervalTime) * time.Minute)
		preNum = int(r.Nums[i])
	}
	wg.Wait()
	elapsed := time.Since(began)
	logc.Info(ctx, "run request concurrency mode cost:", zap.Uint64("task_id", r.TaskId), zap.Duration("cost", elapsed))
}
