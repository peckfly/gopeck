package data

import (
	"context"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"gorm.io/gorm"
)

type (
	recordRepository struct {
		*gorm.DB
	}
)

func NewRecordRepository(db *gorm.DB) repo.RecordRepository {
	return &recordRepository{db}
}

func GetPlanRecordDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return common.GetDB(ctx, defDB).Model(new(repo.PlanRecord))
}

func GetTaskRecordDB(ctx context.Context, defDB *gorm.DB) *gorm.DB {
	return common.GetDB(ctx, defDB).Model(new(repo.TaskRecord))
}

func (r *recordRepository) CreatePlan(ctx context.Context, in *repo.PlanRecord) error {
	return r.WithContext(ctx).Create(in).Error
}

func (r *recordRepository) BatchCreateTasks(ctx context.Context, tasks []repo.TaskRecord) error {
	tx := r.WithContext(ctx).CreateInBatches(tasks, len(tasks))
	if tx.Error != nil {
		return tx.Error
	}
	if tx.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *recordRepository) UpdateTaskById(ctx context.Context, id uint64, in *repo.TaskRecord) error {
	return r.WithContext(ctx).Model(&repo.TaskRecord{}).Where("task_id = ?", id).Updates(in).Error
}

func (r *recordRepository) UpdatePlanById(ctx context.Context, planId uint64, record *repo.PlanRecord) error {
	return r.WithContext(ctx).Model(&repo.PlanRecord{}).Where("plan_id = ?", planId).Updates(record).Error
}

func (r *recordRepository) FindTaskListByPlanId(ctx context.Context, planId uint64) ([]*repo.TaskRecord, error) {
	var records []*repo.TaskRecord
	err := r.WithContext(ctx).Where("plan_id = ?", planId).Find(&records).Error
	return records, err
}

func (r *recordRepository) QueryUserRecordsByUserId(ctx context.Context, userId string, planId uint64, planName string,
	startTime, endTime int64, pp common.PaginationParam) ([]*repo.PlanRecord, *common.PaginationResult, error) {
	db := GetPlanRecordDB(ctx, r.DB)
	db.Where("user_id = ?", userId)
	if planId > 0 {
		db = db.Where("plan_id = ?", planId)
	}
	if len(planName) > 0 {
		db = db.Where("plan_name like ?", "%"+planName+"%")
	}
	if startTime > 0 {
		db = db.Where("create_time >= ?", startTime)
	}
	if endTime > 0 {
		db = db.Where("create_time <= ?", endTime)
	}
	var list []*repo.PlanRecord
	pageResult, err := common.WrapPageQuery(ctx, db, pp, common.QueryOptions{
		OrderFields: []common.OrderByParam{
			{Field: "create_time", Direction: common.DESC},
		},
	}, &list)
	if err != nil {
		return nil, nil, errors.WithStack(err)
	}
	return list, pageResult, nil
}

func (r *recordRepository) QueryTaskRecordsByPlanId(ctx context.Context, planId uint64) (records []*repo.TaskRecord, err error) {
	var tasks []*repo.TaskRecord
	err = r.WithContext(ctx).Where("plan_id = ?", planId).Order("create_time desc").Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *recordRepository) FindPlanRecordByPlanId(ctx context.Context, id uint64) (*repo.PlanRecord, error) {
	var record repo.PlanRecord
	err := r.WithContext(ctx).Where("plan_id = ?", id).First(&record).Error
	return &record, err
}

func (r *recordRepository) FindTaskListByPlanIdWithSize(ctx context.Context, planId uint64, count int) (records []*repo.TaskRecord, err error) {
	var tasks []*repo.TaskRecord
	err = r.WithContext(ctx).Where("plan_id = ?", planId).Order("create_time desc").Limit(count).Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}
