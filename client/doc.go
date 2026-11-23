// Package client is the public SDK for the notify service.
//
// External projects import github.com/vishenosik/notify/client and load
// NOTIFY_* configuration with Load. Callers then construct the transport they
// need:
//
//   - Direct is a gRPC client for Web Push subscriptions and ClientRegistry.
//   - Broker is a RabbitMQ publish-only client for push delivery.
package client
