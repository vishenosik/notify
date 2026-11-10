package mapper

import (
	notifyv1 "github.com/vishenosik/notify/internal/api/grpc/gen/notify_v1"
	"github.com/vishenosik/notify/internal/app/command"
	"github.com/vishenosik/notify/internal/app/common"
)

func SubscribePushCommand(req *notifyv1.SubscribeRequest) *command.SubscribePushCommand {
	if req == nil {
		return &command.SubscribePushCommand{}
	}
	return &command.SubscribePushCommand{
		Endpoint:  req.GetEndpoint(),
		P256dh:    req.GetP256Dh(),
		Auth:      req.GetAuth(),
		UserAgent: req.GetUserAgent(),
	}
}

func SubscribeResponse(result *common.PushSubscriptionResult) *notifyv1.SubscribeResponse {
	if result == nil {
		return &notifyv1.SubscribeResponse{}
	}
	return &notifyv1.SubscribeResponse{Id: result.ID.String()}
}

func VAPIDPublicKeyResponse(result *common.VAPIDPublicKeyResult) *notifyv1.GetVapidPublicKeyResponse {
	if result == nil {
		return &notifyv1.GetVapidPublicKeyResponse{}
	}
	return &notifyv1.GetVapidPublicKeyResponse{PublicKey: result.PublicKey}
}
