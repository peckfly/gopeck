package data

import (
	"context"
	"fmt"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/mods/integrator/biz"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"go.uber.org/zap"
)

// refer to internal/mods/integrator/biz/repo.go
const reportColumns = `plan_id,task_id,url,timestamp,total_num,total_response_content_length,duration_map,status_map,error_map,body_check_result_map,latency_map`

type reporterRepository struct {
	client    driver.Conn
	tableName string
}

// NewReporterRepository create reporter repository
func NewReporterRepository(client driver.Conn, conf *conf.ServerConf) biz.ReporterRepository {
	return &reporterRepository{
		client:    client,
		tableName: conf.Data.Clickhouse.StressReporterTableName,
	}
}

// Report report stress record to clickhouse
func (r *reporterRepository) Report(ctx context.Context, details []*biz.Report) error {
	if len(details) == 0 {
		return nil
	}
	batch, err := r.client.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s (%s)", r.tableName, reportColumns))
	if err != nil {
		logc.Error(ctx, "prepare batch error", zap.Error(err))
	}
	for _, row := range details {
		if err := batch.Append(
			row.PlanId,
			row.TaskId,
			row.Url,
			row.Timestamp,
			row.TotalNum,
			row.TotalResponseContentLength,
			row.DurationMap,
			row.StatusMap,
			row.ErrorMap,
			row.BodyCheckResultMap,
			row.LatencyMap,
		); err != nil {
			logc.Error(ctx, "failed to append data: ", zap.Error(err))
		}
	}
	if err := batch.Send(); err != nil {
		logc.Error(ctx, "failed to send batch: ", zap.Error(err))
	}
	return err
}
