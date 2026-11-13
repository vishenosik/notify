package rabbitmq

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
)

// Config holds the AMQP URL and topology for a single RabbitMQ client.
type Config struct {
	DSN string
}

// Validate checks that URL is present and applies topology defaults.
func (c *Config) Validate() error {
	errs := &multierror.Error{}

	c.DSN = strings.TrimSpace(c.DSN)
	if c.DSN == "" {
		errs = multierror.Append(errs, fmt.Errorf("DSN is required"))
	}

	return errs.ErrorOrNil()
}

const (
	defaultPrefetch = 1
)

// Config holds the AMQP URL and topology for a single RabbitMQ client.
type ChannelConfig struct {
	Exchange   string
	Queue      string
	RoutingKey string
	Prefetch   uint8
}

// Validate checks that URL is present and applies topology defaults.
func (c *ChannelConfig) Validate() error {
	errs := &multierror.Error{}

	if err := errs.ErrorOrNil(); err != nil {
		return err
	}

	c.Exchange = strings.TrimSpace(c.Exchange)
	if c.Exchange == "" {
		errs = multierror.Append(errs, fmt.Errorf("exchange is required"))
	}

	c.Queue = strings.TrimSpace(c.Queue)
	if c.Queue == "" {
		errs = multierror.Append(errs, fmt.Errorf("queue is required"))
	}

	c.RoutingKey = strings.TrimSpace(c.RoutingKey)
	if c.RoutingKey == "" {
		errs = multierror.Append(errs, fmt.Errorf("routing key is required"))
	}

	if c.Prefetch == 0 {
		c.Prefetch = defaultPrefetch
	}

	return errs.ErrorOrNil()
}
