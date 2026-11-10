package mapper

import (
	"github.com/vishenosik/notify/internal/app/common"
	"github.com/vishenosik/notify/internal/domain/entity"
)

func PushSubscriptionResult(sub *entity.PushSubscription) *common.PushSubscriptionResult {
	if sub == nil {
		return nil
	}
	return &common.PushSubscriptionResult{ID: sub.ID()}
}
