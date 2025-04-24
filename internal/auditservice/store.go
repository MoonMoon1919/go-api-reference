package auditservice

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moonmoon1919/go-api-reference/pkg/events"
)

type Storer interface {
	Add(ctx context.Context, item events.Event) error
}

// MARK: Memory
type auditLogMemoryRepository struct {
	items []events.Event
}

func NewInMemoryAuditLogRepository() *auditLogMemoryRepository {
	return &auditLogMemoryRepository{
		items: make([]events.Event, 0),
	}
}

func (a *auditLogMemoryRepository) Add(ctx context.Context, item events.Event) error {
	a.items = append(a.items, item)

	return nil
}

// MARK: SQL
type auditlogSQLRepository struct {
	pool *pgxpool.Pool
}

func NewSQLRepository(pool *pgxpool.Pool) *auditlogSQLRepository {
	return &auditlogSQLRepository{
		pool: pool,
	}
}

func (a *auditlogSQLRepository) Add(ctx context.Context, item events.Event) error {
	x, err := json.Marshal(item.Data)
	if err != nil {
		return err
	}

	_, err = a.pool.Exec(
		ctx,
		"INSERT INTO auditlog (eventname, uid, entityid, timestamp, event) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (eventname, entityid, uid, timestamp) DO NOTHING",
		item.Name,
		item.UserId,
		item.EntityId,
		item.Timestamp,
		x,
	)

	if err != nil {
		return err
	}

	return nil
}
