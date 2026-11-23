package webpush

import (
	"encoding/base64"
	"net/http"
	"testing"
	"time"

	webpushlib "github.com/SherClockHolmes/webpush-go"
	"github.com/stretchr/testify/require"
)

func generateVAPIDPair(t *testing.T) (privateKey, publicKey string) {
	t.Helper()
	priv, pub, err := webpushlib.GenerateVAPIDKeys()
	require.NoError(t, err)
	return priv, pub
}

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	priv, pub := generateVAPIDPair(t)
	otherPriv, otherPub := generateVAPIDPair(t)

	paddedPub := mustPadURLEncoding(t, pub)
	paddedPriv := mustPadURLEncoding(t, priv)

	tooLongPriv := base64.RawURLEncoding.EncodeToString(make([]byte, 33))
	unparseablePub := base64.RawURLEncoding.EncodeToString(append([]byte{0x04}, bytesOf(0xff, 64)...))

	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "valid pair",
			cfg: Config{
				VAPIDPublicKey:  pub,
				VAPIDPrivateKey: priv,
			},
		},
		{
			name: "valid pair with padded base64",
			cfg: Config{
				VAPIDPublicKey:  paddedPub,
				VAPIDPrivateKey: paddedPriv,
			},
		},
		{
			name: "empty public key",
			cfg: Config{
				VAPIDPrivateKey: priv,
			},
			wantErr: "VAPID public key is required",
		},
		{
			name: "whitespace public key",
			cfg: Config{
				VAPIDPublicKey:  "  ",
				VAPIDPrivateKey: priv,
			},
			wantErr: "VAPID public key is required",
		},
		{
			name: "empty private key",
			cfg: Config{
				VAPIDPublicKey: pub,
			},
			wantErr: "VAPID private key is required",
		},
		{
			name:    "both keys empty",
			cfg:     Config{},
			wantErr: "VAPID public key is required",
		},
		{
			name: "invalid public key encoding",
			cfg: Config{
				VAPIDPublicKey:  "!!!",
				VAPIDPrivateKey: priv,
			},
			wantErr: "decode VAPID public key",
		},
		{
			name: "invalid private key encoding",
			cfg: Config{
				VAPIDPublicKey:  pub,
				VAPIDPrivateKey: "!!!",
			},
			wantErr: "decode VAPID private key",
		},
		{
			name: "unparseable private key",
			cfg: Config{
				VAPIDPublicKey:  pub,
				VAPIDPrivateKey: tooLongPriv,
			},
			wantErr: "parse VAPID private key",
		},
		{
			name: "unparseable public key",
			cfg: Config{
				VAPIDPublicKey:  unparseablePub,
				VAPIDPrivateKey: priv,
			},
			wantErr: "parse VAPID public key",
		},
		{
			name: "mismatched pair",
			cfg: Config{
				VAPIDPublicKey:  otherPub,
				VAPIDPrivateKey: priv,
			},
			wantErr: "VAPID public key does not match private key",
		},
		{
			name: "valid independent pair",
			cfg: Config{
				VAPIDPublicKey:  otherPub,
				VAPIDPrivateKey: otherPriv,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			cfg := tc.cfg
			err := cfg.Validate()
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			require.ErrorContains(t, err, tc.wantErr)
		})
	}

	t.Run("both keys empty includes private key error", func(t *testing.T) {
		t.Parallel()
		err := (&Config{}).Validate()
		require.ErrorContains(t, err, "VAPID private key is required")
	})
}

func TestConfigValidateDefaults(t *testing.T) {
	t.Parallel()

	priv, pub := generateVAPIDPair(t)

	t.Run("zero ttl and nil client", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			VAPIDPublicKey:  pub,
			VAPIDPrivateKey: priv,
		}
		require.NoError(t, cfg.Validate())
		require.Equal(t, uint16(defaultTTL), cfg.TTL)

		client, ok := cfg.HTTPClient.(*http.Client)
		require.True(t, ok)
		require.Equal(t, defaultHTTPTimeout, client.Timeout)
	})

	t.Run("keeps provided ttl and client", func(t *testing.T) {
		t.Parallel()
		provided := &http.Client{Timeout: time.Second}
		cfg := &Config{
			VAPIDPublicKey:  pub,
			VAPIDPrivateKey: priv,
			TTL:             120,
			HTTPClient:      provided,
		}
		require.NoError(t, cfg.Validate())
		require.Equal(t, uint16(120), cfg.TTL)
		require.Equal(t, provided, cfg.HTTPClient)
	})
}

func TestPadP256PrivateKey(t *testing.T) {
	t.Parallel()

	require.Equal(t, make([]byte, 32), padP256PrivateKey(nil))
	require.Equal(t, append(make([]byte, 29), 1, 2, 3), padP256PrivateKey([]byte{1, 2, 3}))

	full := bytesOf(0xab, 32)
	require.Equal(t, full, padP256PrivateKey(full))

	tooLong := bytesOf(0xab, 33)
	require.Equal(t, tooLong, padP256PrivateKey(tooLong))
}

func mustPadURLEncoding(t *testing.T, rawURL string) string {
	t.Helper()
	decoded, err := base64.RawURLEncoding.DecodeString(rawURL)
	require.NoError(t, err)
	return base64.URLEncoding.EncodeToString(decoded)
}

func bytesOf(b byte, n int) []byte {
	out := make([]byte, n)
	for i := range out {
		out[i] = b
	}
	return out
}
