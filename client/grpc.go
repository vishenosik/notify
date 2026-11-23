package client

import (
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

type grpcClient struct {
	dialOpts []grpc.DialOption
	conn     *grpc.ClientConn
	cfg      *HttpConnectionConfig
}

type grpcClientOption func(*grpcClient)

// NewGrpcClient constructs a Direct client and always dials gRPC.
func NewGrpcClient(cfg *HttpConnectionConfig, opts ...grpcClientOption) (*grpcClient, error) {
	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("validate config: %w", err)
	}

	client := &grpcClient{
		cfg: cfg,
	}

	for _, opt := range opts {
		opt(client)
	}

	if err := client.dialGRPC(); err != nil {
		return nil, err
	}

	return client, nil
}

func (c *grpcClient) dialGRPC() error {
	var creds credentials.TransportCredentials
	if c.cfg.Insecure {
		creds = insecure.NewCredentials()
	} else {
		creds = credentials.NewClientTLSFromCert(nil, "")
	}

	dialOpts := make([]grpc.DialOption, 0, len(c.dialOpts)+1)
	dialOpts = append(dialOpts,
		grpc.WithTransportCredentials(creds),
		grpc.WithChainUnaryInterceptor(unaryAuthInterceptor(c.cfg.ApiKey)),
	)
	dialOpts = append(dialOpts, c.dialOpts...)

	conn, err := grpc.NewClient(c.cfg.Addr, dialOpts...)
	if err != nil {
		return fmt.Errorf("dial grpc %s: %w", c.cfg.Addr, err)
	}

	c.conn = conn
	return nil
}

func (c *grpcClient) Close() error {
	return c.conn.Close()
}
