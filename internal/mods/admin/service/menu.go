package service

import (
	"github.com/gin-gonic/gin"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/pkg/common"
)

type MenuService struct {
	uc *biz.MenuUsecase
}

func NewMenuService(uc *biz.MenuUsecase) *MenuService {
	return &MenuService{uc: uc}
}

// Query @Tags MenuAPI
// @Security ApiKeyAuth
// @Summary Query menu tree data
// @Param code query string false "Code path of menu (like xxx.xxx.xxx)"
// @Param name query string false "Name of menu"
// @Param includeResources query bool false "Whether to include menu resources"
// @Success 200 {object} common.ResponseResult{data=[]schema.Menu}
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/menus [get]
func (a *MenuService) Query(c *gin.Context) {
	ctx := c.Request.Context()
	var params biz.MenuQueryParam
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

// Get @Tags MenuAPI
// @Security ApiKeyAuth
// @Summary Get menu record by ID
// @Param id path string true "unique id"
// @Success 200 {object} common.ResponseResult{data=schema.Menu}
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/menus/{id} [get]
func (a *MenuService) Get(c *gin.Context) {
	ctx := c.Request.Context()
	item, err := a.uc.Get(ctx, c.Param("id"))
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, item)
}

// Create @Tags MenuAPI
// @Security ApiKeyAuth
// @Summary Create menu record
// @Param body body schema.MenuForm true "Request body"
// @Success 200 {object} common.ResponseResult{data=schema.Menu}
// @Failure 400 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/menus [post]
func (a *MenuService) Create(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(biz.MenuForm)
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

// Update @Tags MenuAPI
// @Security ApiKeyAuth
// @Summary Update menu record by ID
// @Param id path string true "unique id"
// @Param body body schema.MenuForm true "Request body"
// @Success 200 {object} common.ResponseResult
// @Failure 400 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/menus/{id} [put]
func (a *MenuService) Update(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(biz.MenuForm)
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

// Delete @Tags MenuAPI
// @Security ApiKeyAuth
// @Summary Delete menu record by ID
// @Param id path string true "unique id"
// @Success 200 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/menus/{id} [delete]
func (a *MenuService) Delete(c *gin.Context) {
	ctx := c.Request.Context()
	err := a.uc.Delete(ctx, c.Param("id"))
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}
