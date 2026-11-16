package health

import (
	"context"
	"sync"
	"time"

	interfaces "github.com/vishenosik/notify/internal/app/interface"
	"github.com/vishenosik/notify/internal/domain/entity"
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

func (s *service) Check(ctx context.Context) (*entity.HealthCheck, error) {
	var (
		wg sync.WaitGroup
		mu sync.Mutex
	)

	subsystems := make([]entity.SubsystemHealth, 0, len(s.pingers))

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
				subsystems = append(subsystems, entity.SubsystemHealth{
					Name:    pinger.Name(),
					Healthy: false,
				})
				return
			}

			subsystems = append(subsystems, entity.SubsystemHealth{
				Name:    pinger.Name(),
				Healthy: true,
			})
		})
	}

	wg.Wait()

	return &entity.HealthCheck{
		Subsystems: subsystems,
	}, nil
}
