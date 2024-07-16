package registry

import (
	"context"
	"fmt"
)

// Registrar is service registrar.
type Registrar interface {
	// Register the registration.
	Register(ctx context.Context, service *ServiceInstance) error
	// Unregister the registration.
	Unregister(ctx context.Context, service *ServiceInstance) error
}

// Discovery is service discovery.
type Discovery interface {
	// GetService return the service instances in memory according to the service name.
	GetService(ctx context.Context, serviceName string) ([]*ServiceInstance, error)
	// Watch creates a watcher according to the service name.
	Watch(ctx context.Context, serviceName string) (Watcher, error)
}

// Watcher is service watcher.
type Watcher interface {
	// Next returns services in the following two cases:
	// 1.the first time to watch and the service instance list is not empty.
	// 2.any service instance changes found.
	// if the above two conditions are not met, it will block until context deadline exceeded or canceled
	Next() ([]*ServiceInstance, error)
	// Stop close the watcher.
	Stop() error
}

// ServiceInstance is an instance of a service in a discovery system.
type ServiceInstance struct {
	// Name is the service name as registered.
	Name string `json:"name"`
	Addr string `json:"address"`
	// Metadata is the kv pair metadata associated with the service instance.
	Metadata map[string]string `json:"metadata"`
}

func (i *ServiceInstance) String() string {
	return fmt.Sprintf("%s-%s", i.Name, i.Addr)
}
