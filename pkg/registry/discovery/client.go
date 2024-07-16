package discovery

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/peckfly/gopeck/pkg/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	grpcinsecure "google.golang.org/grpc/credentials/insecure"
)

const (
	defaultBalancerName = "round_robin"
)

type (
	clientOptions struct {
		endpoint  string
		discovery registry.Discovery
		insecure  bool
		tlsConf   *tls.Config
		grpcOpts  []grpc.DialOption

		balancerName string
	}
	ClientOption func(o *clientOptions)
)

// WithDiscovery with client discovery.
func WithDiscovery(d registry.Discovery) ClientOption {
	return func(o *clientOptions) {
		o.discovery = d
	}
}

// WithTLSConfig with TLS config.
func WithTLSConfig(c *tls.Config) ClientOption {
	return func(o *clientOptions) {
		o.tlsConf = c
	}
}

// WithOptions with gRPC options.
func WithOptions(opts ...grpc.DialOption) ClientOption {
	return func(o *clientOptions) {
		o.grpcOpts = opts
	}
}

func WithBalancer(name string) ClientOption {
	return func(o *clientOptions) {
		o.balancerName = name
	}
}

// Dial dials grpc endpoint, return client connection.
func Dial(ctx context.Context, endpoint string, opts ...ClientOption) (*grpc.ClientConn, error) {
	options := clientOptions{
		endpoint:     endpoint,
		insecure:     true,
		balancerName: defaultBalancerName,
	}
	for _, opt := range opts {
		opt(&options)
	}
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"loadBalancingPolicy": "%s"}`, options.balancerName)),
	}
	if options.discovery != nil {
		grpcOpts = append(grpcOpts,
			grpc.WithResolvers(
				NewBuilder(
					options.discovery,
					WithInsecure(options.insecure),
				)))
	}
	if options.insecure {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcinsecure.NewCredentials()))
	}
	if options.tlsConf != nil {
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(credentials.NewTLS(options.tlsConf)))
	}
	if len(options.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.grpcOpts...)
	}
	return grpc.DialContext(ctx, options.endpoint, grpcOpts...)
}
