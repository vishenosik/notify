package config

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/vishenosik/notify/internal/config/tag"
)

func TestAddrParamString(t *testing.T) {
	t.Parallel()

	require.Equal(t, ":8080", AddrParam(8080).String())
	require.Equal(t, ":0", AddrParam(0).String())
	require.Equal(t, ":9090", AddrParam(9090).String())
}

func TestLoadDefaultsFromEnv(t *testing.T) {
	isolateLoad(t)

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, AddrParam(8080), cfg.HTTP.Addr)
	require.Equal(t, ":8080", cfg.HTTP.Addr.String())
	require.Equal(t, AddrParam(9090), cfg.GRPC.Addr)
	require.Equal(t, ":9090", cfg.GRPC.Addr.String())
	require.Empty(t, cfg.Database.DSN)
	require.Empty(t, cfg.Admin.APIKey)
	require.Empty(t, cfg.Push.VAPIDPublicKey)
	require.Empty(t, cfg.Push.VAPIDPrivateKey)
	require.Equal(t, "ops@example.com", cfg.Push.VAPIDSubject)
	require.Empty(t, cfg.RabbitMQ.DSN)
	require.Equal(t, GetBuildInfo(), cfg.Build)
	require.Equal(t, tag.Env, cfg.Build.Env.String())
}

func TestLoadFromEnvironmentVariables(t *testing.T) {
	isolateLoad(t)

	t.Setenv("HTTP_ADDR", "8181")
	t.Setenv("GRPC_ADDR", "9191")
	t.Setenv("DATABASE_DSN", "postgres://localhost/notify")
	t.Setenv("ADMIN_API_KEY", "admin-secret")
	t.Setenv("VAPID_PUBLIC_KEY", "pub")
	t.Setenv("VAPID_PRIVATE_KEY", "priv")
	t.Setenv("VAPID_SUBJECT", "ops@test.example")
	t.Setenv("RABBITMQ_DSN", "amqp://guest:guest@localhost:5672/")

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, AddrParam(8181), cfg.HTTP.Addr)
	require.Equal(t, AddrParam(9191), cfg.GRPC.Addr)
	require.Equal(t, "postgres://localhost/notify", cfg.Database.DSN)
	require.Equal(t, "admin-secret", cfg.Admin.APIKey)
	require.Equal(t, "pub", cfg.Push.VAPIDPublicKey)
	require.Equal(t, "priv", cfg.Push.VAPIDPrivateKey)
	require.Equal(t, "ops@test.example", cfg.Push.VAPIDSubject)
	require.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.RabbitMQ.DSN)
}

func TestLoadFromDotEnvFile(t *testing.T) {
	isolateLoad(t)

	content := "" +
		"HTTP_ADDR=7777\n" +
		"GRPC_ADDR=8888\n" +
		"DATABASE_DSN=postgres://file/db\n" +
		"ADMIN_API_KEY=file-admin\n" +
		"VAPID_PUBLIC_KEY=file-pub\n" +
		"VAPID_PRIVATE_KEY=file-priv\n"
	require.NoError(t, os.WriteFile(filepath.Join(".", ".env"), []byte(content), 0o600))

	cfg, err := Load()
	require.NoError(t, err)
	require.Equal(t, AddrParam(7777), cfg.HTTP.Addr)
	require.Equal(t, AddrParam(8888), cfg.GRPC.Addr)
	require.Equal(t, "postgres://file/db", cfg.Database.DSN)
	require.Equal(t, "file-admin", cfg.Admin.APIKey)
	require.Equal(t, "file-pub", cfg.Push.VAPIDPublicKey)
	require.Equal(t, "file-priv", cfg.Push.VAPIDPrivateKey)
}

func TestLoadInvalidEnv(t *testing.T) {
	tests := []struct {
		name string
		key  string
		val  string
	}{
		{name: "invalid port", key: "HTTP_ADDR", val: "abc"},
		{name: "port overflow", key: "GRPC_ADDR", val: "70000"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isolateLoad(t)
			t.Setenv(tc.key, tc.val)
			_, err := Load()
			require.Error(t, err)
			require.ErrorContains(t, err, "read environment")
		})
	}
}

func TestLogValueRedactsSecrets(t *testing.T) {
	t.Parallel()

	cfg := Config{
		Admin: AdminConfig{APIKey: "super-secret-admin"},
		Push: PushConfig{
			VAPIDPublicKey:  "pub",
			VAPIDPrivateKey: "private-vapid",
			VAPIDSubject:    "ops@example.com",
		},
		Database: DatabaseConfig{DSN: "postgres://user:password@localhost:5432/notify"},
		RabbitMQ: RabbitMQConfig{DSN: "amqp://guest:guest@localhost:5672/"},
	}

	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, nil))
	logger.Info("loaded", slog.Any("config", cfg))
	text := buf.String()
	require.NotContains(t, text, "super-secret-admin")
	require.NotContains(t, text, "private-vapid")
	require.NotContains(t, text, "password")
	require.Contains(t, text, "api_key_configured=true")
	require.Contains(t, text, "vapid_private_key_configured=true")
	require.Contains(t, text, "xxxxx")
}

func isolateLoad(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	for _, key := range []string{
		"HTTP_ADDR",
		"GRPC_ADDR",
		"DATABASE_DSN",
		"ADMIN_API_KEY",
		"VAPID_PUBLIC_KEY",
		"VAPID_PRIVATE_KEY",
		"VAPID_SUBJECT",
		"RABBITMQ_DSN",
		"RABBITMQ_EXCHANGE",
		"RABBITMQ_QUEUE",
		"RABBITMQ_ROUTING_KEY",
	} {
		t.Setenv(key, "")
		require.NoError(t, os.Unsetenv(key))
	}
}
