package webpush

import (
	"crypto/ecdh"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
)

const (
	defaultTTL         = 60
	defaultHTTPTimeout = 10 * time.Second
	defaultSubject     = "ops@example.com"
)

// Config holds VAPID credentials and send options for a single Web Push sender.
type Config struct {
	VAPIDPublicKey  string
	VAPIDPrivateKey string
	VAPIDSubject    string
	TTL             uint16
	HTTPClient      HTTPClient
}

// HTTPClient is the transport used to POST an encrypted push. Tests inject a stub.
type HTTPClient interface {
	Do(*http.Request) (*http.Response, error)
}

// Validate checks that both VAPID keys are present and form a matching P-256 pair.
func (c *Config) Validate() error {
	errs := &multierror.Error{}

	if strings.TrimSpace(c.VAPIDPublicKey) == "" {
		errs = multierror.Append(errs, fmt.Errorf("VAPID public key is required"))
	}

	if strings.TrimSpace(c.VAPIDPrivateKey) == "" {
		errs = multierror.Append(errs, fmt.Errorf("VAPID private key is required"))
	}

	if err := errs.ErrorOrNil(); err != nil {
		return err
	}

	if err := validateVAPIDKeyPair(c.VAPIDPublicKey, c.VAPIDPrivateKey); err != nil {
		return fmt.Errorf("VAPID key pair is invalid: %w", err)
	}

	if c.TTL <= 0 {
		c.TTL = defaultTTL
	}

	if c.HTTPClient == nil {
		c.HTTPClient = &http.Client{Timeout: defaultHTTPTimeout}
	}

	return errs.ErrorOrNil()
}

func validateVAPIDKeyPair(publicKey, privateKey string) error {
	publicKeyBytes, err := decodeVapidKey(publicKey)
	if err != nil {
		return fmt.Errorf("decode VAPID public key: %w", err)
	}

	privateKeyBytes, err := decodeVapidKey(privateKey)
	if err != nil {
		return fmt.Errorf("decode VAPID private key: %w", err)
	}

	p256 := ecdh.P256()

	private, err := p256.NewPrivateKey(padP256PrivateKey(privateKeyBytes))
	if err != nil {
		return fmt.Errorf("parse VAPID private key: %w", err)
	}

	public, err := p256.NewPublicKey(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("parse VAPID public key: %w", err)
	}

	if !private.PublicKey().Equal(public) {
		return fmt.Errorf("VAPID public key does not match private key")
	}

	return nil
}

func padP256PrivateKey(key []byte) []byte {
	const p256PrivateKeySize = 32
	if len(key) >= p256PrivateKeySize {
		return key
	}
	padded := make([]byte, p256PrivateKeySize)
	copy(padded[p256PrivateKeySize-len(key):], key)
	return padded
}

func decodeVapidKey(key string) ([]byte, error) {
	bytes, err := base64.URLEncoding.DecodeString(key)
	if err != nil {
		return base64.RawURLEncoding.DecodeString(key)
	}
	return bytes, nil
}
