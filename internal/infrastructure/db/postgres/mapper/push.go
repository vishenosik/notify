package mapper

import (
	"uuid"

	"github.com/vishenosik/notify/internal/domain/entity"
	db "github.com/vishenosik/notify/internal/infrastructure/db/sqlc"
)

func ToDbUpsertPushSubscriptionParams(sub entity.ValidatedPushSubscription) db.UpsertPushSubscriptionParams {
	s := sub.Value
	return db.UpsertPushSubscriptionParams{
		ID:        s.ID(),
		ClientID:  s.ClientID(),
		Endpoint:  s.Endpoint(),
		P256dh:    s.P256dh(),
		Auth:      s.Auth(),
		UserAgent: textFromString(s.UserAgent()),
		CreatedAt: unixFromTime(s.CreatedAt()),
		UpdatedAt: unixFromTime(s.UpdatedAt()),
	}
}

func FromDbPushSubscription(row db.PushSubscription) *entity.PushSubscription {
	return entity.NewPushSubscription(
		entity.PushSubscriptionVariants{
			ClientID:  row.ClientID,
			Endpoint:  row.Endpoint,
			P256dh:    row.P256dh,
			Auth:      row.Auth,
			UserAgent: stringFromText(row.UserAgent),
		},
		entity.CreatedPushSubscription(entity.Base{
			ID:        row.ID,
			CreatedAt: timeFromUnix(row.CreatedAt),
			UpdatedAt: timeFromUnix(row.UpdatedAt),
		}),
	)
}

func FromDbPushSubscriptionList(rows []db.PushSubscription) []*entity.PushSubscription {
	out := make([]*entity.PushSubscription, len(rows))
	for i, row := range rows {
		out[i] = FromDbPushSubscription(row)
	}
	return out
}

func ToDbDeletePushSubscriptionParams(
	clientID uuid.UUID,
	endpoint string,
) db.DeletePushSubscriptionByClientIdAndEndpointParams {
	return db.DeletePushSubscriptionByClientIdAndEndpointParams{
		ClientID: clientID,
		Endpoint: endpoint,
	}
}

func ToDbGetPushSubscriptionByClientIdAndEndpointParams(
	clientID uuid.UUID,
	endpoint string,
) db.GetPushSubscriptionByClientIdAndEndpointParams {
	return db.GetPushSubscriptionByClientIdAndEndpointParams{
		ClientID: clientID,
		Endpoint: endpoint,
	}
}
