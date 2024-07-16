package server

import (
	"context"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/mods/admin/service"
	"github.com/peckfly/gopeck/internal/pkg/common"
	"github.com/peckfly/gopeck/internal/pkg/errors"
	"github.com/peckfly/gopeck/internal/pkg/middleware"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"github.com/peckfly/gopeck/pkg/proc"
	"go.uber.org/zap"
	"strconv"
)

const (
	apiPrefix            = "/api/"
	maxCopyContentLength = 33554432
)

var allowedPrefixes = []string{apiPrefix}

type AdminServer struct {
	conf          *conf.ServerConf
	stressService *service.StressService
	userService   *service.UserService
	loginService  *service.LoginService
	casbinService *service.CasbinService
	menuService   *service.MenuService
	roleService   *service.RoleService
	nodeService   *service.NodeService

	engine *gin.Engine
	port   int
}

func NewAdminServer(
	conf *conf.ServerConf,
	stressService *service.StressService,
	userService *service.UserService,
	loginService *service.LoginService,
	casbinService *service.CasbinService,
	menuService *service.MenuService,
	roleService *service.RoleService,
	nodeService *service.NodeService,
) *AdminServer {
	return &AdminServer{
		conf:          conf,
		stressService: stressService,
		userService:   userService,
		loginService:  loginService,
		casbinService: casbinService,
		menuService:   menuService,
		roleService:   roleService,
		nodeService:   nodeService,
	}
}

func (s *AdminServer) Init(ctx context.Context) error {
	engine := gin.New()
	s.engine = engine
	gin.DefaultWriter = log.NewZapInfoWriter()
	gin.DefaultErrorWriter = log.NewZapErrorWriter()
	engine.NoMethod(func(c *gin.Context) {
		common.ResError(c, errors.MethodNotAllowed("", "Method Not Allowed"))
	})
	engine.NoRoute(func(c *gin.Context) {
		common.ResError(c, errors.NotFound("", "Not Found"))
	})
	s.useHttpMiddlewares()
	s.port = s.conf.Server.Http.Port
	s.registerRoutersV1()
	return s.casbinService.Load(ctx)
}

func (s *AdminServer) registerRoutersV1() {
	v1 := s.engine.Group(apiPrefix + "v1")
	stressGroup := v1.Group("stress")
	{
		stressGroup.POST("start", s.stressService.StartStress)
		stressGroup.POST("stop", s.stressService.StopStress)
		stressGroup.POST("restart", s.stressService.RestartStress)
		stressGroup.GET("record_plan", s.stressService.QueryPlanRecords)
		stressGroup.GET("record_task", s.stressService.QueryTaskRecords)
		stressGroup.GET("plan_query", s.stressService.QueryPlanAndTaskRecords)
	}
	nodeGroup := v1.Group("nodes")
	{
		nodeGroup.GET("list", s.nodeService.QueryAllNodes)
		nodeGroup.GET("detail", s.nodeService.QueryNodes)
		nodeGroup.POST("update_quota", s.nodeService.UpdateNodeQuota)
	}
	captcha := v1.Group("captcha")
	{
		captcha.GET("id", s.loginService.GetCaptcha)
		captcha.GET("image", s.loginService.ResponseCaptcha)
	}
	v1.POST("login", s.loginService.Login)

	current := v1.Group("current")
	{
		current.POST("refresh-token", s.loginService.RefreshToken)
		current.GET("user", s.loginService.GetUserInfo)
		current.GET("menus", s.loginService.QueryMenus)
		current.PUT("password", s.loginService.UpdatePassword)
		current.PUT("user", s.loginService.UpdateUser)
		current.POST("logout", s.loginService.Logout)
	}
	menu := v1.Group("menus")
	{
		menu.GET("", s.menuService.Query)
		menu.GET(":id", s.menuService.Get)
		menu.POST("", s.menuService.Create)
		menu.PUT(":id", s.menuService.Update)
		menu.DELETE(":id", s.menuService.Delete)
	}
	role := v1.Group("roles")
	{
		role.GET("", s.roleService.Query)
		role.GET(":id", s.roleService.Get)
		role.POST("", s.roleService.Create)
		role.PUT(":id", s.roleService.Update)
		role.DELETE(":id", s.roleService.Delete)
	}
	user := v1.Group("users")
	{
		user.GET("", s.userService.Query)
		user.GET(":id", s.userService.Get)
		user.POST("", s.userService.Create)
		user.PUT(":id", s.userService.Update)
		user.DELETE(":id", s.userService.Delete)
		user.PATCH(":id/reset-pwd", s.userService.ResetPassword)
	}
}

func (s *AdminServer) useHttpMiddlewares() {
	s.engine.Use(gin.Recovery())
	s.engine.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		AllowedPathPrefixes:      allowedPrefixes,
		MaxOutputRequestBodyLen:  1024 * 1024,
		MaxOutputResponseBodyLen: 1024 * 1024,
	}))
	s.engine.Use(middleware.Cors())
	s.engine.Use(middleware.TraceWithConfig(middleware.TraceConfig{
		AllowedPathPrefixes: allowedPrefixes,
		RequestHeaderKey:    "X-Request-Id",
		ResponseTraceKey:    "X-Trace-Id",
	}))
	s.engine.Use(middleware.CopyBodyWithConfig(middleware.CopyBodyConfig{
		AllowedPathPrefixes: allowedPrefixes,
		MaxContentLen:       maxCopyContentLength,
	}))
	s.engine.Use(middleware.AuthWithConfig(middleware.AuthConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: s.conf.Auth.SkippedPathPrefixes,
		ParseUserID:         s.loginService.ParseUserID,
		RootID:              s.conf.Rbac.RootId, //  fix Auth.Disable  = true user not exist problem
	}))
	s.engine.Use(middleware.CasbinWithConfig(middleware.CasbinConfig{
		AllowedPathPrefixes: allowedPrefixes,
		SkippedPathPrefixes: s.conf.Casbin.SkippedPathPrefixes,
		Skipper: func(c *gin.Context) bool {
			if s.conf.Casbin.Disable ||
				common.FromIsRootUser(c.Request.Context()) {
				return true
			}
			return false
		},
		GetEnforcer: func(c *gin.Context) *casbin.Enforcer {
			return s.casbinService.GetEnforcer()
		},
		GetSubjects: func(c *gin.Context) []string {
			return common.FromUserCache(c.Request.Context()).RoleIDs
		},
	}))
}

func (s *AdminServer) Run(ctx context.Context, cleanUp ...func()) error {
	logc.Info(ctx, "start admin server")
	return proc.GracefulRun(ctx, func(ctx context.Context) (func(), error) {
		go func() {
			err := s.engine.Run(":" + strconv.Itoa(s.port))
			log.Must(err)
		}()
		return func() {
			err := s.casbinService.Release(ctx)
			logc.Error(ctx, "casbin release error", zap.Error(err))
			for _, cleanFunc := range cleanUp {
				cleanFunc()
			}
		}, nil
	})
}
