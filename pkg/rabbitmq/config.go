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
func (c *Config) validate() error {
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

type channelConfig interface {
	validate() error
	id() string
}

// Config holds the AMQP URL and topology for a single RabbitMQ client.
type ChannelConfig struct {
	Exchange   string
	RoutingKey string
}

func (c *ChannelConfig) id() string {
	return c.Exchange
}

// Validate checks that URL is present and applies topology defaults.
func (c *ChannelConfig) validate() error {
	errs := &multierror.Error{}

	c.Exchange = strings.TrimSpace(c.Exchange)
	if c.Exchange == "" {
		errs = multierror.Append(errs, fmt.Errorf("exchange is required"))
	}

	c.RoutingKey = strings.TrimSpace(c.RoutingKey)
	if c.RoutingKey == "" {
		errs = multierror.Append(errs, fmt.Errorf("routing key is required"))
	}

	return errs.ErrorOrNil()
}

// Config holds the AMQP URL and topology for a single RabbitMQ client.
type ConsumerConfig struct {
	ChannelConfig
	Queue    string
	Prefetch uint8
}

// Validate checks that URL is present and applies topology defaults.
func (c *ConsumerConfig) validate() error {
	errs := &multierror.Error{}

	if err := c.ChannelConfig.validate(); err != nil {
		errs = multierror.Append(errs, err)
	}

	c.Queue = strings.TrimSpace(c.Queue)
	if c.Queue == "" {
		errs = multierror.Append(errs, fmt.Errorf("queue is required"))
	}

	if c.Prefetch == 0 {
		c.Prefetch = defaultPrefetch
	}

	return errs.ErrorOrNil()
}

// PublisherConfig holds topology for a publish-only AMQP channel.
type PublisherConfig struct {
	ChannelConfig
	Queue string
}

// Validate checks that exchange and routing key are present. Queue is optional:
// when set, the publisher declares and binds it so messages are not dropped
// before a consumer comes up.
func (c *PublisherConfig) validate() error {
	errs := &multierror.Error{}

	if err := c.ChannelConfig.validate(); err != nil {
		errs = multierror.Append(errs, err)
	}

	c.Queue = strings.TrimSpace(c.Queue)

	return errs.ErrorOrNil()
}
