package client

import (
	"log/slog"

	"google.golang.org/grpc"
)

// WithGRPCDialOptions appends extra gRPC dial options such as a custom dialer.
func WithGRPCDialOptions(opts ...grpc.DialOption) grpcClientOption {
	return func(c *grpcClient) {
		c.dialOpts = append(c.dialOpts, opts...)
	}
}

// WithBrokerLogger sets the logger used by the RabbitMQ client.
func WithBrokerLogger(logger *slog.Logger) brokerClientOption {
	return func(c *BrokerClient) {
		c.logger = logger
	}
}
