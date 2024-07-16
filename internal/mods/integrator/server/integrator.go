package server

import (
	"context"
	v1 "github.com/peckfly/gopeck/api/integrator/v1"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/mods/integrator/service"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"github.com/peckfly/gopeck/pkg/proc"
	"github.com/peckfly/gopeck/pkg/registry"
	"go.uber.org/zap"
)

type IntegratorServer struct {
	conf             *conf.ServerConf
	integrateService *service.IntegrateService
	registrar        registry.Registrar

	grpcServer *registry.GrpcServer
}

// NewIntegratorServer create integrator server
func NewIntegratorServer(
	conf *conf.ServerConf,
	integrateService *service.IntegrateService,
	registrar registry.Registrar,
) *IntegratorServer {
	return &IntegratorServer{
		conf:             conf,
		integrateService: integrateService,
		registrar:        registrar,
	}
}

func (s *IntegratorServer) Init() error {
	grpcServer := registry.NewGrpcServer(s.conf.Name,
		s.conf.Server.Grpc.Addr,
		registry.WithServerRegistrar(s.registrar))
	v1.RegisterIntegrateServiceServer(grpcServer, s.integrateService)
	s.grpcServer = grpcServer
	return nil
}

func (s *IntegratorServer) Run(ctx context.Context, cleanUp func()) error {
	logc.Info(ctx, "start integrator server")
	return proc.GracefulRun(ctx, func(ctx context.Context) (func(), error) {
		go func() {
			err := s.grpcServer.Start()
			log.Must(err)
		}()
		return func() {
			s.gracefulStop()
			if cleanUp != nil {
				cleanUp()
			}
		}, nil
	})
}

func (s *IntegratorServer) gracefulStop() {
	// todo
	// 1、unregister from etcd
	// 2、wait for all receive summary task finish
	// 3、close grpc server
	// 4、shutdown
	err := s.Stop()
	if err != nil {
		log.Error("stop server error", zap.Error(err))
	}
}

func (s *IntegratorServer) Stop() error {
	return s.grpcServer.Close()
}
