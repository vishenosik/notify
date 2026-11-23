package mapper

import (
	notifyv1 "github.com/vishenosik/notify/internal/api/grpc/gen/notify_v1"
	"github.com/vishenosik/notify/internal/app/command"
	"github.com/vishenosik/notify/internal/app/common"
)

func CreateClientCommand(req *notifyv1.CreateClientRequest) *command.CreateClientCommand {
	if req == nil {
		return &command.CreateClientCommand{}
	}
	return &command.CreateClientCommand{
		Name:        req.GetName(),
		Description: req.GetDescription(),
	}
}

func CreateClientResponse(result *common.CreatedClientResult) *notifyv1.CreateClientResponse {
	if result == nil {
		return &notifyv1.CreateClientResponse{}
	}
	return &notifyv1.CreateClientResponse{
		Client: Client(&result.ClientResult),
		Key:    result.Key,
	}
}

func Client(result *common.ClientResult) *notifyv1.Client {
	if result == nil {
		return nil
	}
	return &notifyv1.Client{
		Id:           result.ID.String(),
		Name:         result.Name,
		Description:  result.Description,
		MaskedApiKey: result.MaskedAPIKey,
		CreatedAt:    result.CreatedAt.UTC().Unix(),
		UpdatedAt:    result.UpdatedAt.UTC().Unix(),
	}
}

func Clients(results []*common.ClientResult) []*notifyv1.Client {
	out := make([]*notifyv1.Client, 0, len(results))
	for _, result := range results {
		out = append(out, Client(result))
	}
	return out
}

func RotateAPIKeyResponse(result *common.CreatedClientResult) *notifyv1.RotateAPIKeyResponse {
	if result == nil {
		return &notifyv1.RotateAPIKeyResponse{}
	}
	return &notifyv1.RotateAPIKeyResponse{
		Service: Client(&result.ClientResult),
		Key:     result.Key,
	}
}
