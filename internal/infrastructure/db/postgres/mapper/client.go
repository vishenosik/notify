package mapper

import (
	"time"
	"uuid"

	"github.com/vishenosik/notify/internal/domain/entity"
	db "github.com/vishenosik/notify/internal/infrastructure/db/sqlc"
)

func ToDbCreateClientParams(client entity.ValidatedClient) db.CreateClientParams {
	c := client.Value
	return db.CreateClientParams{
		ID:          c.ID(),
		Name:        c.Name(),
		Description: textFromString(c.Description()),
		CreatedAt:   unixFromTime(c.CreatedAt()),
		UpdatedAt:   unixFromTime(c.UpdatedAt()),
	}
}

func FromDbClient(row db.Client) *entity.Client {
	return entity.NewClient(
		entity.ClientVariants{
			Name:        row.Name,
			Description: stringFromText(row.Description),
		},
		entity.CreatedClient(entity.Base{
			ID:        row.ID,
			CreatedAt: timeFromUnix(row.CreatedAt),
			UpdatedAt: timeFromUnix(row.UpdatedAt),
		}),
	)
}

func FromDbClientList(rows []db.Client) []*entity.Client {
	out := make([]*entity.Client, len(rows))
	for i, row := range rows {
		out[i] = FromDbClient(row)
	}
	return out
}

func ToDbDeleteClientParams(id uuid.UUID, archivedAt time.Time) db.DeleteClientParams {
	return db.DeleteClientParams{
		ID:         id,
		ArchivedAt: int8FromTime(archivedAt),
	}
}
