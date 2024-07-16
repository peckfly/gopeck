package biz

import (
	"context"
	"github.com/panjf2000/ants/v2"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"go.uber.org/zap"
	"math"
)

const KST int = 3

func (b *RequesterUsecase) report(ctx context.Context, r *Requester) error {
	var err error
	r.poolFunc, err = ants.NewPoolWithFunc(b.conf.ReportGoroutineNum, func(data interface{}) {
		result, ok := data.(*repo.Aggregate)
		if !ok {
			return
		}
		err = b.queRepository.RatePush(ctx, r.TaskId, result)
		if err != nil {
			log.Error("failed to push result", zap.Error(err))
		}
	})
	if err != nil {
		log.Error("failed to create pool", zap.Error(err))
		return err
	}
	intervalsLen := len(r.Nums)
	stat := func() {
		ags := make([]*repo.Aggregate, intervalsLen)
		for i := range ags {
			ags[i] = repo.NewAggeRate()
		}
		tars := make(map[int64]*repo.Aggregate, KST)
		stop := false
		for result := range r.results {
			timestamp := result.TimeStamp
			costSecond := timestamp - r.StartTime
			interval := min(int(intervalsLen)-1, int(costSecond)/int(r.StepIntervalTime))

			var car *repo.Aggregate
			var ok bool
			if car, ok = tars[timestamp]; !ok {
				car = repo.NewAggeRate()
				car.Timestamp = timestamp
				car.Interval = interval
				tars[timestamp] = car
			}
			car.TotalNum++
			car.TotalResponseContentLength += result.ResponseContentLength
			car.DurationMap[int32(result.Duration.Milliseconds())]++
			car.StatusMap[int32(result.StatusCode)]++
			car.ErrorMap[result.Err]++
			car.BodyCheckResultMap[result.BodyCheckResult]++
			if result.Stop {
				car.Stop = true
				stop = true
			}
			if len(tars) > KST {
				minTimestamp := int64(math.MaxInt64)
				var bar *repo.Aggregate
				for ts, ar := range tars {
					if ts < minTimestamp {
						minTimestamp = ts
						bar = ar
					}
				}
				err = r.poolFunc.Invoke(bar)
				if err != nil {
					logc.Error(ctx, "failed to push result", zap.Error(err))
				}
				delete(tars, minTimestamp)
			}

			ags[interval].Interval = interval
			ags[interval].TotalNum++
			ags[interval].TotalResponseContentLength += result.ResponseContentLength
			ags[interval].DurationMap[int32(result.Duration.Milliseconds())]++
			ags[interval].StatusMap[int32(result.StatusCode)]++
			ags[interval].ErrorMap[result.Err]++
			ags[interval].BodyCheckResultMap[result.BodyCheckResult]++
		}
		for _, bar := range tars {
			err = r.poolFunc.Invoke(bar)
			if err != nil {
				logc.Error(ctx, "failed to push result", zap.Error(err))
			}
		}
		logc.Info(ctx, "result channel is closed, done report")
		if err != nil {
			logc.Error(ctx, "failed to push stop result", zap.Error(err))
		}
		for i, agr := range ags {
			err = b.queRepository.AggregatePush(ctx, r.TaskId, &repo.Aggregate{
				PlanId:                     r.PlanId,
				TaskId:                     r.TaskId,
				Interval:                   agr.Interval,
				TotalNum:                   agr.TotalNum,
				TotalResponseContentLength: agr.TotalResponseContentLength,
				DurationMap:                agr.DurationMap,
				StatusMap:                  agr.StatusMap,
				ErrorMap:                   agr.ErrorMap,
				BodyCheckResultMap:         agr.BodyCheckResultMap,
				Stop:                       i == len(ags)-1 && stop,
			})
		}
		if err != nil {
			logc.Error(ctx, "failed to push aggregate result", zap.Error(err))
		}
		r.done <- true
	}
	go stat()
	return nil
}
