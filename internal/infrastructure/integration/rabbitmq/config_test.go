package rabbitmq

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfigValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "valid",
			cfg: Config{
				URL: "amqp://guest:guest@localhost:5672/",
			},
		},
		{
			name:    "empty url",
			cfg:     Config{},
			wantErr: "URL is required",
		},
		{
			name: "whitespace url",
			cfg: Config{
				URL: "  \t",
			},
			wantErr: "URL is required",
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
}

func TestConfigValidateDefaults(t *testing.T) {
	t.Parallel()

	t.Run("empty topology and zero prefetch", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			URL: " amqp://guest:guest@localhost:5672/ ",
		}
		require.NoError(t, cfg.Validate())
		require.Equal(t, "amqp://guest:guest@localhost:5672/", cfg.URL)
		require.Equal(t, defaultExchange, cfg.Exchange)
		require.Equal(t, defaultQueue, cfg.Queue)
		require.Equal(t, defaultRoutingKey, cfg.RoutingKey)
		require.Equal(t, defaultPrefetch, cfg.Prefetch)
	})

	t.Run("whitespace topology uses defaults", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			URL:        "amqp://guest:guest@localhost:5672/",
			Exchange:   "  ",
			Queue:      "\t",
			RoutingKey: " ",
			Prefetch:   -1,
		}
		require.NoError(t, cfg.Validate())
		require.Equal(t, defaultExchange, cfg.Exchange)
		require.Equal(t, defaultQueue, cfg.Queue)
		require.Equal(t, defaultRoutingKey, cfg.RoutingKey)
		require.Equal(t, defaultPrefetch, cfg.Prefetch)
	})

	t.Run("keeps provided topology", func(t *testing.T) {
		t.Parallel()
		cfg := &Config{
			URL:        "amqp://guest:guest@localhost:5672/",
			Exchange:   "custom",
			Queue:      "custom.queue",
			RoutingKey: "custom.key",
			Prefetch:   8,
		}
		require.NoError(t, cfg.Validate())
		require.Equal(t, "custom", cfg.Exchange)
		require.Equal(t, "custom.queue", cfg.Queue)
		require.Equal(t, "custom.key", cfg.RoutingKey)
		require.Equal(t, 8, cfg.Prefetch)
	})
}
