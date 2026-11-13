package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

type AddrParam uint16

func (a AddrParam) String() string {
	return fmt.Sprintf(":%d", a)
}

type Config struct {
	Build    BuildInfo
	HTTP     HTTPConfig
	GRPC     GRPCConfig
	Database DatabaseConfig
	Admin    AdminConfig
	Push     PushConfig
	RabbitMQ RabbitMQConfig
}

type HTTPConfig struct {
	Addr AddrParam `env:"HTTP_ADDR" env-default:"8080"`
}

type GRPCConfig struct {
	Addr AddrParam `env:"GRPC_ADDR" env-default:"9090"`
}

type DatabaseConfig struct {
	DSN string `env:"DATABASE_DSN"`
}

type AdminConfig struct {
	APIKey string `env:"ADMIN_API_KEY"`
}

type PushConfig struct {
	VAPIDPublicKey  string `env:"VAPID_PUBLIC_KEY"`
	VAPIDPrivateKey string `env:"VAPID_PRIVATE_KEY"`
	VAPIDSubject    string `env:"VAPID_SUBJECT" env-default:"ops@example.com"`
}

type RabbitMQConfig struct {
	DSN string `env:"RABBITMQ_DSN"`
}

func Load() (Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return Config{}, fmt.Errorf("read environment: %w", err)
		}
	}
	cfg.Build = GetBuildInfo()
	return cfg, nil
}
