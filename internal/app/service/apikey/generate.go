package apikey

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"

	"github.com/vishenosik/notify/internal/domain/entity"
	"github.com/vishenosik/notify/internal/domain/xerrors"
)

const (
	keyPrefix     = "ntf_"
	keyByteLength = 32
)

type Material struct {
	plaintext string
	hash      []byte
	prefix    string
	last4     string
}

func (m Material) Plaintext() string { return m.plaintext }

func (m Material) Hash() []byte { return m.hash }

func (m Material) Prefix() string { return m.prefix }

func (m Material) Last4() string { return m.last4 }

func GenerateMaterial() (*Material, error) {
	raw := make([]byte, keyByteLength)
	if _, err := rand.Read(raw); err != nil {
		return nil, xerrors.APIKeyGenerationError()
	}

	plaintext := keyPrefix + base64.RawURLEncoding.EncodeToString(raw)
	if len(plaintext) < entity.APIKeyPrefixLen+entity.APIKeyLast4Len {
		return nil, xerrors.APIKeyGenerationError()
	}

	return &Material{
		plaintext: plaintext,
		hash:      HashKey(plaintext),
		prefix:    plaintext[:entity.APIKeyPrefixLen],
		last4:     plaintext[len(plaintext)-entity.APIKeyLast4Len:],
	}, nil
}

func HashKey(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

func generateMaterial() (*Material, error) {
	return GenerateMaterial()
}

func hashKey(raw string) []byte {
	return HashKey(raw)
}
