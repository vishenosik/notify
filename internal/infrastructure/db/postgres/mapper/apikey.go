package mapper

import (
	"time"
	"uuid"

	"github.com/vishenosik/notify/internal/domain/entity"
	db "github.com/vishenosik/notify/internal/infrastructure/db/sqlc"
)

func ToDbCreateAPIKeyParams(key entity.ValidatedAPIKey) db.CreateAPIKeyParams {
	k := key.Value
	return db.CreateAPIKeyParams{
		ID:        k.ID(),
		ClientID:  k.ClientID(),
		Name:      k.Name(),
		KeyHash:   k.KeyHash(),
		Prefix:    k.Prefix(),
		Last4:     k.Last4(),
		CreatedAt: unixFromTime(k.CreatedAt()),
		UpdatedAt: unixFromTime(k.UpdatedAt()),
	}
}

func FromDbAPIKey(row db.ApiKey) *entity.APIKey {
	return entity.NewAPIKey(
		entity.APIKeyVariants{
			ClientID: row.ClientID,
			Name:     row.Name,
			KeyHash:  row.KeyHash,
			Prefix:   row.Prefix,
			Last4:    row.Last4,
		},
		entity.CreatedAPIKey(entity.Base{
			ID:        row.ID,
			CreatedAt: timeFromUnix(row.CreatedAt),
			UpdatedAt: timeFromUnix(row.UpdatedAt),
		}),
	)
}

func FromDbAPIKeys(rows []db.ApiKey) []*entity.APIKey {
	out := make([]*entity.APIKey, len(rows))
	for i, row := range rows {
		out[i] = FromDbAPIKey(row)
	}
	return out
}

func ToDbDeleteAPIKeysByClientIdParams(clientID uuid.UUID, archivedAt time.Time) db.DeleteAPIKeysByClientIdParams {
	return db.DeleteAPIKeysByClientIdParams{
		ClientID:   clientID,
		ArchivedAt: int8FromTime(archivedAt),
	}
}

func ToDbDeleteAPIKeyParams(id uuid.UUID, archivedAt time.Time) db.DeleteAPIKeyParams {
	return db.DeleteAPIKeyParams{
		ID:         id,
		ArchivedAt: int8FromTime(archivedAt),
	}
}
