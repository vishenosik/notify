package rabbitmq

import (
	"fmt"
	"strings"

	"github.com/hashicorp/go-multierror"
)

const (
	defaultExchange   = "notify"
	defaultQueue      = "notify.push.send"
	defaultRoutingKey = "push.send"
	defaultPrefetch   = 1
)

// Config holds the AMQP URL and topology for a single RabbitMQ client.
type Config struct {
	URL        string
	Exchange   string
	Queue      string
	RoutingKey string
	Prefetch   int
}

// Validate checks that URL is present and applies topology defaults.
func (c *Config) Validate() error {
	errs := &multierror.Error{}

	c.URL = strings.TrimSpace(c.URL)
	if c.URL == "" {
		errs = multierror.Append(errs, fmt.Errorf("URL is required"))
	}

	if err := errs.ErrorOrNil(); err != nil {
		return err
	}

	c.Exchange = strings.TrimSpace(c.Exchange)
	if c.Exchange == "" {
		c.Exchange = defaultExchange
	}

	c.Queue = strings.TrimSpace(c.Queue)
	if c.Queue == "" {
		c.Queue = defaultQueue
	}

	c.RoutingKey = strings.TrimSpace(c.RoutingKey)
	if c.RoutingKey == "" {
		c.RoutingKey = defaultRoutingKey
	}

	if c.Prefetch <= 0 {
		c.Prefetch = defaultPrefetch
	}

	return errs.ErrorOrNil()
}
