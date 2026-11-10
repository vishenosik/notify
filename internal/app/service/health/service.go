package health

import (
	"context"
	"sync"
	"time"

	"github.com/vishenosik/notify/internal/app/common"
	interfaces "github.com/vishenosik/notify/internal/app/interface"
	"github.com/vishenosik/notify/internal/domain/repository"
)

const healthCheckPingTimeout = 2 * time.Second

type service struct {
	pingers []repository.Pinger
}

func New(pingers ...repository.Pinger) interfaces.HealthUseCase {
	return &service{
		pingers: pingers,
	}
}

func (s *service) Check(ctx context.Context) (*common.HealthCheckResponse, error) {
	results := make(map[string]bool)

	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		failed bool
	)

	for _, pinger := range s.pingers {
		if pinger == nil {
			continue
		}

		wg.Go(func() {
			ctx, cancel := context.WithTimeout(ctx, healthCheckPingTimeout)
			defer cancel()

			mu.Lock()
			defer mu.Unlock()

			if err := pinger.Ping(ctx); err != nil {
				failed = true
				results[pinger.Name()] = true
				return
			}

			results[pinger.Name()] = false
		})
	}

	wg.Wait()

	return &common.HealthCheckResponse{
		Healthy:    !failed,
		Subsystems: results,
	}, nil
}
