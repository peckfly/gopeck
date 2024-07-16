package registry

import (
	"context"
	"crypto/tls"
	"github.com/peckfly/gopeck/pkg/log"
	"github.com/peckfly/gopeck/pkg/registry/endpoint"
	"github.com/peckfly/gopeck/pkg/registry/host"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"net"
	"net/url"
	"time"
)

type (
	GrpcServer struct {
		name    string
		addr    string
		timeout time.Duration

		metadata        map[string]string
		listener        net.Listener
		serviceInstance *ServiceInstance
		registrar       Registrar
		*grpc.Server
		endpoint *url.URL
		tlsConf  *tls.Config
		err      error
	}
	ServerOption func(server *GrpcServer)
)

// NewGrpcServer create grpc server
func NewGrpcServer(name string, addr string, opts ...ServerOption) *GrpcServer {
	server := &GrpcServer{
		name:    name,
		addr:    addr,
		timeout: time.Second * 2,
		Server:  grpc.NewServer(),
	}
	for _, opt := range opts {
		opt(server)
	}
	return server
}

// WithServerRegistrar set registrar
func WithServerRegistrar(registrar Registrar) ServerOption {
	return func(server *GrpcServer) {
		server.registrar = registrar
	}
}

// WithServerTimeout set timeout
func WithServerTimeout(timeout time.Duration) ServerOption {
	return func(server *GrpcServer) {
		server.timeout = timeout
	}
}

func WithServerMetadata(metadata map[string]string) ServerOption {
	return func(server *GrpcServer) {
		server.metadata = metadata
	}
}

// TLSConfig with TLS config.
func TLSConfig(c *tls.Config) ServerOption {
	return func(s *GrpcServer) {
		s.tlsConf = c
	}
}

// Endpoint with server address.
func Endpoint(endpoint *url.URL) ServerOption {
	return func(s *GrpcServer) {
		s.endpoint = endpoint
	}
}

// Start start grpc server with registrar option
func (s *GrpcServer) Start() error {
	listener, err := net.Listen("tcp", ":"+s.addr)
	if err != nil {
		return err
	}
	s.listener = listener
	if s.registrar != nil {
		if s.endpoint == nil {
			addr, err := host.Extract(s.addr, s.listener)
			if err != nil {
				s.err = err
				return err
			}
			s.endpoint = endpoint.NewEndpoint(endpoint.Scheme("grpc", s.tlsConf != nil), addr)
		}

		s.serviceInstance = &ServiceInstance{
			Name:     s.name,
			Addr:     s.endpoint.String(),
			Metadata: s.metadata,
		}
		ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
		err = s.registrar.Register(ctx, s.serviceInstance)
		cancel()
		if err != nil {
			log.Info("register service instance error", zap.String("serviceName", s.name), zap.Error(err))
			return nil
		}
	}
	return s.Server.Serve(listener)
}

// Close close grpc server
func (s *GrpcServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()
	if s.registrar != nil {
		if err := s.registrar.Unregister(ctx, s.serviceInstance); err != nil {
			return err
		}
	}
	s.GracefulStop()
	return nil
}

func (s *GrpcServer) Endpoint() (*url.URL, error) {
	return s.endpoint, nil
}
