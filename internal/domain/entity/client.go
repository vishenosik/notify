package entity

import (
	"strings"

	"github.com/hashicorp/go-multierror"

	"github.com/vishenosik/notify/internal/domain/xerrors"
)

type Client struct {
	base
	name        string
	description string
}

func NewClient(variants ClientVariants, opts ...func(*Client)) *Client {
	c := &Client{
		base:        newBase(),
		name:        strings.TrimSpace(variants.Name),
		description: strings.TrimSpace(variants.Description),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func CreatedClient(variants Base) func(*Client) {
	return func(c *Client) {
		c.base = newBase(created(variants))
	}
}

func (c *Client) Name() string { return c.name }

func (c *Client) Description() string { return c.description }

type ValidatedClient = *Validated[*Client]

type ClientVariants struct {
	Name        string
	Description string
}

func (c Client) validate() error {
	var err *multierror.Error

	if c.name == "" {
		err = multierror.Append(err, xerrors.InvalidNameError("client", c.name))
	}

	if er := c.base.validate(); er != nil {
		err = multierror.Append(err, er)
	}

	return err.ErrorOrNil()
}
