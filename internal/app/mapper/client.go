package mapper

import (
	"github.com/vishenosik/notify/internal/app/common"
	"github.com/vishenosik/notify/internal/domain/entity"
)

func ClientResult(client *entity.Client, key *entity.APIKey) *common.ClientResult {
	if client == nil {
		return nil
	}
	return &common.ClientResult{
		ID:           client.ID(),
		Name:         client.Name(),
		Description:  client.Description(),
		MaskedAPIKey: MaskedAPIKey(key),
		CreatedAt:    client.CreatedAt(),
		UpdatedAt:    client.UpdatedAt(),
	}
}

func CreatedClientResult(client *entity.Client, key *entity.APIKey, plaintext string) *common.CreatedClientResult {
	result := ClientResult(client, key)
	if result == nil {
		return nil
	}
	return &common.CreatedClientResult{
		ClientResult: *result,
		Key:          plaintext,
	}
}

func MaskedAPIKey(key *entity.APIKey) string {
	if key == nil {
		return ""
	}
	prefix := key.Prefix()
	last4 := key.Last4()
	if prefix == "" && last4 == "" {
		return ""
	}
	return prefix + "..." + last4
}
