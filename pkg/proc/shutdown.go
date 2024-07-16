package proc

import (
	"context"
	"github.com/peckfly/gopeck/pkg/log"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// GracefulRun The Run function sets up a signal handler and executes a handler function until a termination signal
// is received.
func GracefulRun(ctx context.Context, handler func(ctx context.Context) (func(), error)) error {
	state := 1
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGUSR1, syscall.SIGUSR2)
	cleanFn, err := handler(ctx)
	if err != nil {
		return err
	}
EXIT:
	for {
		sig := <-sc
		log.Context(ctx).Info("Received signal", zap.String("signal", sig.String()))
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			state = 0
			break EXIT
		case syscall.SIGHUP:
		default:
			break EXIT
		}
	}
	cleanFn()
	log.Context(ctx).Info("Server exit, bye...")
	time.Sleep(time.Millisecond * 100)
	os.Exit(state)
	return nil
}
