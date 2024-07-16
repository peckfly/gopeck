package service

import (
	"github.com/gin-gonic/gin"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/pkg/common"
)

type StressService struct {
	uc *biz.StressUsecase
}

func NewStressService(uc *biz.StressUsecase) *StressService {
	return &StressService{uc: uc}
}

func (s *StressService) StartStress(c *gin.Context) {
	var plan biz.Plan
	if err := c.ShouldBindJSON(&plan); err != nil {
		common.ResError(c, err)
		return
	}
	plan.UserId = getUserId(c)
	err := s.uc.StartStress(c.Request.Context(), &plan)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}

func (s *StressService) RestartStress(c *gin.Context) {
	var restart biz.Restart
	restart.UserId = getUserId(c)
	if err := c.ShouldBindJSON(&restart); err != nil {
		common.ResError(c, err)
		return
	}
	err := s.uc.RestartStress(c.Request.Context(), &restart)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}

func (s *StressService) StopStress(c *gin.Context) {
	var stop biz.Stop
	stop.UserId = getUserId(c)
	if err := c.ShouldBindJSON(&stop); err != nil {
		common.ResError(c, err)
		return
	}
	err := s.uc.StopStress(c.Request.Context(), &stop)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}

func (s *StressService) QueryPlanRecords(c *gin.Context) {
	userId := getUserId(c)
	var params biz.PlanRecordQuery
	if err := common.ParseQuery(c, &params); err != nil {
		common.ResError(c, err)
		return
	}
	result, err := s.uc.QueryUserRecords(c.Request.Context(), userId, params)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResPage(c, result.Data, result.PageResult)
}

func (s *StressService) QueryTaskRecords(c *gin.Context) {
	userId := getUserId(c)
	var params biz.TaskRecordQuery
	if err := common.ParseQuery(c, &params); err != nil {
		common.ResError(c, err)
		return
	}
	record, err := s.uc.QueryPlanTaskRecords(c.Request.Context(), userId, params)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, record)
}

func (s *StressService) QueryPlanAndTaskRecords(c *gin.Context) {
	userId := getUserId(c)
	var params biz.TaskRecordQuery
	if err := common.ParseQuery(c, &params); err != nil {
		common.ResError(c, err)
		return
	}
	record, err := s.uc.QueryPlanAndTaskRecords(c.Request.Context(), userId, params)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, record)
}

func getUserId(c *gin.Context) string {
	return common.FromUserID(c.Request.Context())
}
