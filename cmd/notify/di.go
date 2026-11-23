package main

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc"

	grpcapi "github.com/vishenosik/notify/internal/api/grpc"
	"github.com/vishenosik/notify/internal/api/rest"
	interfaces "github.com/vishenosik/notify/internal/app/interface"
	apikeyuc "github.com/vishenosik/notify/internal/app/service/apikey"
	clientuc "github.com/vishenosik/notify/internal/app/service/client"
	"github.com/vishenosik/notify/internal/app/service/health"
	pushuc "github.com/vishenosik/notify/internal/app/service/push"
	"github.com/vishenosik/notify/internal/config"
	"github.com/vishenosik/notify/internal/domain/repository"
	"github.com/vishenosik/notify/internal/infrastructure/db/postgres"
	broker "github.com/vishenosik/notify/internal/infrastructure/integration/rabbitmq"
	"github.com/vishenosik/notify/internal/infrastructure/integration/webpush"
	"github.com/vishenosik/notify/pkg/logs"
	"github.com/vishenosik/notify/pkg/rabbitmq"
)

type diContainer struct {
	cfg    *config.Config
	logger *slog.Logger

	httpHandler http.Handler
	grpcServer  *grpc.Server

	webPushNotifier repository.WebPushNotifier

	clientUseCase interfaces.ClientUseCase
	apiKeyUseCase interfaces.APIKeyUseCase
	pushUseCase   interfaces.PushUseCase
	healthUseCase interfaces.HealthUseCase

	// database

	pool                 *postgres.Pool
	clientRepo           repository.ClientRepository
	apiKeyRepo           repository.APIKeyRepository
	pushSubscriptionRepo repository.PushSubscriptionRepository

	// integration
	rabbitmqClient   *rabbitmq.AmqpClient
	pushSendConsumer sync.Once

	// closers
	closers []io.Closer
}

func newDIContainer() *diContainer {
	return &diContainer{
		pushSendConsumer: sync.Once{},
		closers:          make([]io.Closer, 0),
	}
}

func (d *diContainer) Config() *config.Config {
	if d.cfg == nil {
		cfg, err := config.Load()
		if err != nil {
			slog.Error("failed to load config", logs.Error(err))
			os.Exit(1)
		}
		d.cfg = &cfg
	}
	return d.cfg
}

func (d *diContainer) Logger() *slog.Logger {
	if d.logger == nil {
		cfg := d.Config()
		d.logger = logs.NewLogger(logs.LoggerConfig{
			AppName:     cfg.Build.ServiceName,
			Env:         cfg.Build.Env,
			EnableDebug: config.FlagDebug,
		})
	}
	return d.logger
}

func (d *diContainer) Pool() *postgres.Pool {
	if d.pool == nil {
		pool, err := postgres.NewPool(context.Background(), d.Config().Database.DSN)
		if err != nil {
			slog.Error("failed to connect to database", logs.Error(err))
			os.Exit(1)
		}
		d.pool = pool
		d.closers = append(d.closers, pool)
	}
	return d.pool
}

func (d *diContainer) ClientRepo() repository.ClientRepository {
	if d.clientRepo == nil {
		d.clientRepo = d.Pool().ClientRepository()
	}
	return d.clientRepo
}

func (d *diContainer) APIKeyRepo() repository.APIKeyRepository {
	if d.apiKeyRepo == nil {
		d.apiKeyRepo = d.Pool().APIKeyRepository()
	}
	return d.apiKeyRepo
}

func (d *diContainer) PushSubscriptionRepo() repository.PushSubscriptionRepository {
	if d.pushSubscriptionRepo == nil {
		d.pushSubscriptionRepo = d.Pool().PushSubscriptionRepository()
	}
	return d.pushSubscriptionRepo
}

func (d *diContainer) RabbitmqClient() *rabbitmq.AmqpClient {
	if d.rabbitmqClient == nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

		client, err := rabbitmq.NewRabbitMQClient(
			ctx,
			&rabbitmq.Config{
				DSN: d.Config().RabbitMQ.DSN,
			},
			rabbitmq.WithLogger(d.Logger()),
		)
		if err != nil {
			slog.Error("failed to create web push sender", logs.Error(err))
			cancel()
			os.Exit(1)
		}
		d.rabbitmqClient = client
		d.closers = append(d.closers, client)
		cancel()
	}
	return d.rabbitmqClient
}

func (d *diContainer) PushSendBroker() {
	d.pushSendConsumer.Do(func() {
		_, err := broker.NewPushSendBroker(
			d.RabbitmqClient(),
			d.PushUseCase(),
		)
		if err != nil {
			slog.Error("failed to create push send broker", logs.Error(err))
			os.Exit(1)
		}
	})
}

func (d *diContainer) WebPushNotifier() repository.WebPushNotifier {
	if d.webPushNotifier == nil {
		cfg := d.Config().Push
		if strings.TrimSpace(cfg.VAPIDPublicKey) == "" || strings.TrimSpace(cfg.VAPIDPrivateKey) == "" {
			d.webPushNotifier = webpush.NewNoopSender()
			return d.webPushNotifier
		}
		sender, err := webpush.NewSender(&webpush.Config{
			VAPIDPublicKey:  cfg.VAPIDPublicKey,
			VAPIDPrivateKey: cfg.VAPIDPrivateKey,
			VAPIDSubject:    cfg.VAPIDSubject,
		})
		if err != nil {
			slog.Error("failed to create web push sender", logs.Error(err))
			os.Exit(1)
		}
		d.webPushNotifier = sender
	}
	return d.webPushNotifier
}

func (d *diContainer) APIKeyUseCase() interfaces.APIKeyUseCase {
	if d.apiKeyUseCase == nil {
		d.apiKeyUseCase = apikeyuc.New(d.ClientRepo(), d.APIKeyRepo(), d.Pool())
	}
	return d.apiKeyUseCase
}

func (d *diContainer) ClientUseCase() interfaces.ClientUseCase {
	if d.clientUseCase == nil {
		d.clientUseCase = clientuc.New(
			d.ClientRepo(),
			d.APIKeyRepo(),
			d.PushSubscriptionRepo(),
			d.Pool(),
		)
	}
	return d.clientUseCase
}

func (d *diContainer) PushUseCase() interfaces.PushUseCase {
	if d.pushUseCase == nil {
		cfg := d.Config().Push
		d.pushUseCase = pushuc.New(
			d.PushSubscriptionRepo(),
			d.APIKeyUseCase(),
			d.WebPushNotifier(),
			cfg.VAPIDPublicKey,
			cfg.VAPIDPrivateKey,
		)
	}
	return d.pushUseCase
}

func (d *diContainer) HealthUseCase() interfaces.HealthUseCase {
	if d.healthUseCase == nil {
		d.healthUseCase = health.New(d.Pool(), d.RabbitmqClient())
	}
	return d.healthUseCase
}

func (d *diContainer) HTTPHandler() http.Handler {
	if d.httpHandler == nil {
		d.httpHandler = rest.Routes(d.HealthUseCase())
	}
	return d.httpHandler
}

func (d *diContainer) GRPCServer() *grpc.Server {
	if d.grpcServer == nil {
		d.grpcServer = grpcapi.NewServer(
			d.Config().Admin.APIKey,
			d.PushUseCase(),
			d.ClientUseCase(),
			d.APIKeyUseCase(),
		)
	}
	return d.grpcServer
}
