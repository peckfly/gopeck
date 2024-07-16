package service

import (
	"github.com/gin-gonic/gin"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/pkg/common"
)

type UserService struct {
	uc *biz.UserUsecase
}

func NewUserService(uc *biz.UserUsecase) *UserService {
	return &UserService{uc: uc}
}

// Query @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Query user list
// @Param current query int true "pagination index" default(1)
// @Param pageSize query int true "pagination size" default(10)
// @Param username query string false "Username for login"
// @Param name query string false "Name of user"
// @Param status query string false "Status of user (activated, freezed)"
// @Success 200 {object} common.ResponseResult{data=[]biz.User}
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/users [get]
func (a *UserService) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params biz.UserQueryParam
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

// Get @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Get user record by ID
// @Param id path string true "unique id"
// @Success 200 {object} common.ResponseResult{data=biz.User}
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/users/{id} [get]
func (a *UserService) Get(c *gin.Context) {
	ctx := c.Request.Context()
	item, err := a.uc.Get(ctx, c.Param("id"))
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, item)
}

// Create @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Create user record
// @Param body body biz.UserForm true "Request body"
// @Success 200 {object} common.ResponseResult{data=biz.User}
// @Failure 400 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/users [post]
func (a *UserService) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(biz.UserForm)
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

// Update @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Update user record by ID
// @Param id path string true "unique id"
// @Param body body biz.UserForm true "Request body"
// @Success 200 {object} common.ResponseResult
// @Failure 400 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/users/{id} [put]
func (a *UserService) Update(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(biz.UserForm)
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

// Delete @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Delete user record by ID
// @Param id path string true "unique id"
// @Success 200 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/users/{id} [delete]
func (a *UserService) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := a.uc.Delete(ctx, c.Param("id"))
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}

// ResetPassword @Tags UserAPI
// @Security ApiKeyAuth
// @Summary Reset user password by ID
// @Param id path string true "unique id"
// @Success 200 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/users/{id}/reset-pwd [patch]
func (a *UserService) ResetPassword(c *gin.Context) {
	ctx := c.Request.Context()
	err := a.uc.ResetPassword(ctx, c.Param("id"))
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}
