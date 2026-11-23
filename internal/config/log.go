package config

import (
	"log/slog"
	"net/url"
	"strings"
)

func (c Config) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("Build", c.Build),
		slog.Any("HTTP", c.HTTP),
		slog.Any("GRPC", c.GRPC),
		slog.Attr{Key: "Database", Value: c.Database.LogValue()},
		slog.Attr{Key: "Admin", Value: c.Admin.LogValue()},
		slog.Attr{Key: "Push", Value: c.Push.LogValue()},
		slog.Attr{Key: "RabbitMQ", Value: c.RabbitMQ.LogValue()},
	)
}

func (c AdminConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Bool("api_key_configured", strings.TrimSpace(c.APIKey) != ""),
	)
}

func (c PushConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Bool("vapid_public_key_configured", strings.TrimSpace(c.VAPIDPublicKey) != ""),
		slog.Bool("vapid_private_key_configured", strings.TrimSpace(c.VAPIDPrivateKey) != ""),
		slog.String("vapid_subject", c.VAPIDSubject),
	)
}

func (c DatabaseConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("dsn", redactURL(c.DSN)),
	)
}

func (c RabbitMQConfig) LogValue() slog.Value {
	return slog.GroupValue(
		slog.String("dsn", redactURL(c.DSN)),
	)
}

func redactURL(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.User == nil {
		return raw
	}
	username := parsed.User.Username()
	if _, hasPassword := parsed.User.Password(); hasPassword {
		parsed.User = url.UserPassword(username, "xxxxx")
	}
	return parsed.Redacted()
}
