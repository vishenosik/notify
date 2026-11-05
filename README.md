# notify

Web Push notification service: gRPC for subscriptions, a caller-service registry with API keys,
and delivery via RabbitMQ and the Web Push Protocol (VAPID).

**Only other services** talk to this backend. The settings UI is a separate application.

Specs:
- [Architecture](docs/specs/architecture.md)
- [Service description](docs/specs/README.md)
- [Database schema](docs/specs/database.dbml)

## Stack

- Go 1.27
- gRPC — `PushService`, `ClientRegistry`
- [chi](https://github.com/go-chi/chi) — HTTP `/healthz`, `/version`
- [pgx](https://github.com/jackc/pgx) + [sqlc](https://sqlc.dev/) — PostgreSQL
- [goose](https://github.com/pressly/goose) — migrations
- RabbitMQ — push send queue
- [Task](https://taskfile.dev/) — build and development

## Quick start

```bash
task setup
task compose:up
task goose:up-all
task run-dev
```

gRPC listens on `:9090`, HTTP health on `:8080`.

## Contract

[Protobuf files](internal/interface/api/grpc/proto)

### gRPC `PushService` (`x-api-key`)

- `GetVapidPublicKey`
- `Subscribe(endpoint, p256dh, auth, user_agent)`
- `Unsubscribe(endpoint | subscription_id)`

### gRPC `ClientRegistry` (`x-admin-key` = `ADMIN_API_KEY`)

- `CreateClient` — creates a client and returns the `ntf_…` key once
- `ListClients` / `GetClient` / `DeleteClient`
- `RotateAPIKey`

### RabbitMQ `notify.push.send`

```json
{
  "subscription_id": "uuid",
  "api_key": "ntf_…",
  "title": "Title",
  "body": "Body",
  "data": { "url": "/somewhere" }
}
```

The message is authenticated with the service API key. The push is delivered only to a
subscription that belongs to that service.
