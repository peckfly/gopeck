package service

import (
	"github.com/gin-gonic/gin"
	"github.com/peckfly/gopeck/internal/mods/admin/biz"
	"github.com/peckfly/gopeck/internal/pkg/common"
)

type LoginService struct {
	uc *biz.LoginUsecase
}

func NewLoginService(uc *biz.LoginUsecase) *LoginService {
	return &LoginService{uc: uc}
}

// GetCaptcha @Tags LoginAPI
// @Summary Get captcha ID
// @Success 200 {object} common.ResponseResult{data=schema.Captcha}
// @Router /api/v1/captcha/id [get]
func (a *LoginService) GetCaptcha(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := a.uc.GetCaptcha(ctx)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, data)
}

// ResponseCaptcha @Tags LoginAPI
// @Summary Response captcha image
// @Param id query string true "Captcha ID"
// @Param reload query number false "Reload captcha image (reload=1)"
// @Produce image/png
// @Success 200 "Captcha image"
// @Failure 404 {object} common.ResponseResult
// @Router /api/v1/captcha/image [get
func (a *LoginService) ResponseCaptcha(c *gin.Context) {
	ctx := c.Request.Context()
	err := a.uc.ResponseCaptcha(ctx, c.Writer, c.Query("id"), c.Query("reload") == "1")
	if err != nil {
		common.ResError(c, err)
	}
}

// Login @Tags LoginAPI
// @Summary Login system with username and password
// @Param body body schema.LoginForm true "Request body"
// @Success 200 {object} common.ResponseResult{data=schema.LoginToken}
// @Failure 400 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/login [post
func (a *LoginService) Login(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(biz.LoginForm)
	if err := common.ParseJSON(c, item); err != nil {
		common.ResError(c, err)
		return
	}
	data, err := a.uc.Login(ctx, item.Trim())
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, data)
}

// Logout @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Logout system
// @Success 200 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/current/logout [post]
func (a *LoginService) Logout(c *gin.Context) {
	ctx := c.Request.Context()
	err := a.uc.Logout(ctx)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}

// RefreshToken @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Refresh current access token
// @Success 200 {object} common.ResponseResult{data=schema.LoginToken}
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/current/refresh-token [post]
func (a *LoginService) RefreshToken(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := a.uc.RefreshToken(ctx)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, data)
}

// GetUserInfo @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Get current user info
// @Success 200 {object} common.ResponseResult{data=schema.User}
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/current/user [get
func (a *LoginService) GetUserInfo(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := a.uc.GetUserInfo(ctx)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, data)
}

// UpdatePassword @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Change current user password
// @Param body body schema.UpdateLoginPassword true "Request body"
// @Success 200 {object} common.ResponseResult
// @Failure 400 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/current/password [put]
func (a *LoginService) UpdatePassword(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(biz.UpdateLoginPassword)
	if err := common.ParseJSON(c, item); err != nil {
		common.ResError(c, err)
		return
	}

	err := a.uc.UpdatePassword(ctx, item)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}

// QueryMenus @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Query current user menus based on the current user role
// @Success 200 {object} common.ResponseResult{data=[]schema.Menu}
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/current/menus [get]
func (a *LoginService) QueryMenus(c *gin.Context) {
	ctx := c.Request.Context()
	data, err := a.uc.QueryMenus(ctx)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResSuccess(c, data)
}

// UpdateUser @Tags LoginAPI
// @Security ApiKeyAuth
// @Summary Update current user info
// @Param body body schema.UpdateCurrentUser true "Request body"
// @Success 200 {object} common.ResponseResult
// @Failure 400 {object} common.ResponseResult
// @Failure 401 {object} common.ResponseResult
// @Failure 500 {object} common.ResponseResult
// @Router /api/v1/current/user [put]
func (a *LoginService) UpdateUser(c *gin.Context) {
	ctx := c.Request.Context()
	item := new(biz.UpdateCurrentUser)
	if err := common.ParseJSON(c, item); err != nil {
		common.ResError(c, err)
		return
	}

	err := a.uc.UpdateUser(ctx, item)
	if err != nil {
		common.ResError(c, err)
		return
	}
	common.ResOK(c)
}

func (a *LoginService) ParseUserID(c *gin.Context) (string, error) {
	id, err := a.uc.ParseUserID(c)
	if err != nil {
		return "", err
	}
	return id, nil
}
