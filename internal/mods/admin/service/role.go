package service

import (
	"github.com/gin-gonic/gin"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/pkg/common"
)

type RoleService struct {
	uc *biz.RoleUsecase
}

func NewRoleService(uc *biz.RoleUsecase) *RoleService {
	return &RoleService{uc: uc}
}

// Query @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary Query role list
// @Param current query int true "pagination index" default(1)
// @Param pageSize query int true "pagination size" default(10)
// @Param name query string false "Display name of role"
// @Param status query string false "Status of role (disabled, enabled)"
// @Success 200 {object} common.ResponseResult{data=[]biz.Role}
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/roles [get]
func (a *RoleService) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params biz.RoleQueryParam
	if err := common.ParseQuery(c, &params); err != nil {
		common.ResError(c, err)
		return
	}

	result, err := a.uc.Query(ctx, params)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResPage(c, result.Data, result.PageResult)
}

// Get @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary Get role record by ID
// @Param id path string true "unique id"
// @Success 200 {object} common.ResponseResult{data=biz.Role}
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/roles/{id} [get]
func (a *RoleService) Get(c *gin.Context) {
	ctx := c.Request.Context()
	item, err := a.uc.Get(ctx, c.Param("id"))
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, item)
}

// Create @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary Create role record
// @Param body body biz.RoleForm true "Request body"
// @Success 200 {object} common.ResponseResult{data=biz.Role}
// @Failure 400 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/roles [post]
func (a *RoleService) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(biz.RoleForm)
	if err := common.ParseJSON(c, item); err != nil {
		common.ResError(c, err)
		return
	} else if err := item.Validate(); err != nil {
		common.ResError(c, err)
		return
	}

	result, err := a.uc.Create(ctx, item)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, result)
}

// Update @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary Update role record by ID
// @Param id path string true "unique id"
// @Param body body biz.RoleForm true "Request body"
// @Success 200 {object} common.ResponseResult
// @Failure 400 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/roles/{id} [put]
func (a *RoleService) Update(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(biz.RoleForm)
	if err := common.ParseJSON(c, item); err != nil {
		common.ResError(c, err)
		return
	} else if err := item.Validate(); err != nil {
		common.ResError(c, err)
		return
	}

	err := a.uc.Update(ctx, c.Param("id"), item)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}

// Delete @Tags RoleAPI
// @Security ApiKeyAuth
// @Summary Delete role record by ID
// @Param id path string true "unique id"
// @Success 200 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/roles/{id} [delete]
func (a *RoleService) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := a.uc.Delete(ctx, c.Param("id"))
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}
