package entity

import (
	"errors"
	"maps"
	"strings"

	"github.com/hashicorp/go-multierror"
)

var (
	ErrPushEndpointRequired = errors.New("push endpoint is required")
	ErrPushP256dhRequired   = errors.New("push p256dh is required")
	ErrPushAuthRequired     = errors.New("push auth is required")
	ErrPushTitleRequired    = errors.New("push title is required")
)

// PushTarget is the Web Push subscription material required to deliver one message.
type PushTarget struct {
	endpoint string
	p256dh   string
	auth     string
}

func NewPushTarget(endpoint, p256dh, auth string) PushTarget {
	return PushTarget{
		endpoint: strings.TrimSpace(endpoint),
		p256dh:   strings.TrimSpace(p256dh),
		auth:     strings.TrimSpace(auth),
	}
}

func (t PushTarget) Endpoint() string { return t.endpoint }

func (t PushTarget) P256dh() string { return t.p256dh }

func (t PushTarget) Auth() string { return t.auth }

func (t PushTarget) Validate() error {
	var err *multierror.Error

	if t.endpoint == "" {
		err = multierror.Append(err, ErrPushEndpointRequired)
	}
	if t.p256dh == "" {
		err = multierror.Append(err, ErrPushP256dhRequired)
	}
	if t.auth == "" {
		err = multierror.Append(err, ErrPushAuthRequired)
	}

	return err.ErrorOrNil()
}

// PushMessage is the browser-visible payload of a single web push.
type PushMessage struct {
	title string
	body  string
	data  map[string]string
}

func NewPushMessage(title, body string, data map[string]string) PushMessage {
	return PushMessage{
		title: strings.TrimSpace(title),
		body:  strings.TrimSpace(body),
		data:  maps.Clone(data),
	}
}

func (m PushMessage) Title() string { return m.title }

func (m PushMessage) Body() string { return m.body }

func (m PushMessage) Data() map[string]string {
	return maps.Clone(m.data)
}

func (m PushMessage) Validate() error {
	if m.title == "" {
		return ErrPushTitleRequired
	}
	return nil
}

// PushSendResult is the provider outcome of a single send.
// Expired is true when the subscription is gone (HTTP 403, 404, or 410)
// and the caller should drop it.
type PushSendResult struct {
	StatusCode int
	Expired    bool
}
