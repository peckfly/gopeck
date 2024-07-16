package server

import (
	"context"
	v1 "github.com/peckfly/gopeck/api/pecker/v1"
	"github.com/peckfly/gopeck/internal/conf"
	"github.com/peckfly/gopeck/internal/mods/common/repo"
	"github.com/peckfly/gopeck/internal/mods/pecker/service"
	"github.com/peckfly/gopeck/internal/pkg/consts"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/peckfly/gopeck/pkg/log/logc"
	"github.com/peckfly/gopeck/pkg/proc"
	"github.com/peckfly/gopeck/pkg/registry"
	"go.uber.org/zap"
	"strconv"
	"time"
)

type PeckerServer struct {
	conf           *conf.ServerConf
	peckService    *service.PeckService
	registrar      registry.Registrar
	grpcServer     *registry.GrpcServer
	nodeRepository repo.NodeRepository
}

// NewPeckerServer create pecker server
func NewPeckerServer(
	conf *conf.ServerConf,
	peckService *service.PeckService,
	registrar registry.Registrar,
	nodeRepository repo.NodeRepository,
) *PeckerServer {
	return &PeckerServer{
		conf:           conf,
		peckService:    peckService,
		registrar:      registrar,
		nodeRepository: nodeRepository,
	}
}

func (s *PeckerServer) Init() error {
	grpcServer := registry.NewGrpcServer(s.conf.Name,
		s.conf.Server.Grpc.Addr,
		registry.WithServerRegistrar(s.registrar),
		registry.WithServerMetadata(map[string]string{
			consts.MaxConcurrencyNum: strconv.Itoa(s.conf.StressConf.MaxGoroutineNum),
			consts.MaxRpsNum:         strconv.Itoa(s.conf.StressConf.MaxRpsNum),
		}))
	v1.RegisterPeckServiceServer(grpcServer, s.peckService)
	s.grpcServer = grpcServer
	return nil
}

func (s *PeckerServer) Run(ctx context.Context, cleanUp func()) error {
	logc.Info(ctx, "start pecker server")
	return proc.GracefulRun(ctx, func(ctx context.Context) (func(), error) {
		go func() {
			err := s.grpcServer.Start()
			log.Must(err)
		}()
		go s.ReportNodeTask()
		return func() {
			s.gracefulStop()
			if cleanUp != nil {
				cleanUp()
			}
			s.deleteNodeInfo()
		}, nil
	})
}

func (s *PeckerServer) gracefulStop() {
	// todo
	// 1、unregister from etcd
	// 2、wait for all stress task finish
	// 3、close grpc server
	// 4、shutdown
	err := s.Stop()
	if err != nil {
		log.Error("stop server error", zap.Error(err))
	}
}

func (s *PeckerServer) Stop() error {
	return s.grpcServer.Close()
}

// ReportNodeTask generates a function comment for the given function body in a markdown code block with the correct language syntax.
//
// No parameters.
// No return value.
func (s *PeckerServer) ReportNodeTask() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			endpoint, err := s.grpcServer.Endpoint()
			if err != nil {
				log.Error("get endpoint error", zap.Error(err))
				continue
			}
			nodeState := GetNodeState()
			nodeState.Addr = endpoint.String()
			err = s.nodeRepository.ReportNodeInfo(context.Background(), nodeState)
			if err != nil {
				log.Error("report node info error", zap.Error(err))
			}
		}
	}
}

func (s *PeckerServer) deleteNodeInfo() {
	endpoint, err := s.grpcServer.Endpoint()
	if err != nil {
		log.Error("get endpoint error", zap.Error(err))
		return
	}
	err = s.nodeRepository.DeleteNodeInfo(context.Background(), endpoint.String())
	log.Info("delete node info", zap.String("endpoint", endpoint.String()))
	if err != nil {
		log.Error("delete node info error", zap.Error(err))
	}
	err = s.nodeRepository.DeleteNodeStateInfo(context.Background(), endpoint.String())
	if err != nil {
		log.Error("delete node info error", zap.Error(err))
	}
}
