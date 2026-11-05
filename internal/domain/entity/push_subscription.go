package entity

import (
	"strings"

	"github.com/hashicorp/go-multierror"

	xerrors "github.com/vishenosik/notify/internal/domain/xerrors"
)

type PushSubscription struct {
	base
	endpoint  string
	p256dh    string
	auth      string
	userAgent string
}

func NewPushSubscription(variants PushSubscriptionVariants, opts ...func(*PushSubscription)) *PushSubscription {
	s := &PushSubscription{
		base:      newBase(),
		endpoint:  variants.Endpoint,
		p256dh:    variants.P256dh,
		auth:      variants.Auth,
		userAgent: variants.UserAgent,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func CreatedPushSubscription(variants Base) func(*PushSubscription) {
	return func(s *PushSubscription) {
		s.base = newBase(created(variants))
	}
}

func (s *PushSubscription) Endpoint() string { return s.endpoint }

func (s *PushSubscription) P256dh() string { return s.p256dh }

func (s *PushSubscription) Auth() string { return s.auth }

func (s *PushSubscription) UserAgent() string { return s.userAgent }

type ValidatedPushSubscription = *Validated[*PushSubscription]

type PushSubscriptionVariants struct {
	Endpoint  string
	P256dh    string
	Auth      string
	UserAgent string
}

func (s PushSubscription) validate() error {
	var err *multierror.Error

	if strings.TrimSpace(s.endpoint) == "" {
		err = multierror.Append(err, xerrors.PushEndpointRequiredError())
	}
	if strings.TrimSpace(s.p256dh) == "" {
		err = multierror.Append(err, xerrors.PushP256dhRequiredError())
	}
	if strings.TrimSpace(s.auth) == "" {
		err = multierror.Append(err, xerrors.PushAuthRequiredError())
	}

	if er := s.base.validate(); er != nil {
		err = multierror.Append(err, er)
	}

	return err.ErrorOrNil()
}
