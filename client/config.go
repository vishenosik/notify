package client

import (
	"fmt"
	"strings"
)

type HttpConnectionConfig struct {
	Addr     string
	Insecure bool
	ApiKey   string
}

func (c *HttpConnectionConfig) validate() error {
	c.Addr = strings.TrimSpace(c.Addr)
	if c.Addr == "" {
		return fmt.Errorf("addr is required")
	}

	c.ApiKey = strings.TrimSpace(c.ApiKey)
	if c.ApiKey == "" {
		return fmt.Errorf("api key is required")
	}

	return nil
}

type BrokerConnectionConfig struct {
	DSN    string
	ApiKey string
}

func (c *BrokerConnectionConfig) validate() error {
	c.DSN = strings.TrimSpace(c.DSN)
	if c.DSN == "" {
		return fmt.Errorf("DSN is required")
	}

	c.ApiKey = strings.TrimSpace(c.ApiKey)
	if c.ApiKey == "" {
		return fmt.Errorf("api key is required")
	}

	return nil
}
