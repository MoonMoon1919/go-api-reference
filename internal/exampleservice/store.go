package exampleservice

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
	"github.com/valkey-io/valkey-go/valkeyaside"
)

const (
	notFoundMsg = "NOT FOUND"
	errKey      = "ERR"
)

var notFoundError = errors.New("example not found")

// MARK: Interface
type Storer interface {
	Add(ctx context.Context, item example.Example) (example.Example, error)
	Get(ctx context.Context, id string) (example.Example, error)
	List(ctx context.Context, id string, limit int, page int) ([]example.Example, error)
	Update(ctx context.Context, item example.Example) (example.Example, error)
	Delete(ctx context.Context, userId string) error
}

// MARK: Memory
type exampleRepository struct {
	items       map[string]example.Example
	byUserIndex map[string][]example.Example
}

func NewInMemoryExampleRepository() *exampleRepository {
	return &exampleRepository{
		items:       make(map[string]example.Example),
		byUserIndex: make(map[string][]example.Example),
	}
}

func (e *exampleRepository) Add(ctx context.Context, item example.Example) (example.Example, error) {
	// Pretend to be a DB
	item.Id = uuid.NewString()

	e.items[item.Id] = item

	_, ok := e.byUserIndex[item.UserId]

	if !ok {
		e.byUserIndex[item.UserId] = []example.Example{item}
	} else {
		e.byUserIndex[item.UserId] = append(e.byUserIndex[item.UserId], item)
	}

	return item, nil
}

func (e *exampleRepository) Update(ctx context.Context, item example.Example) (example.Example, error) {
	e.items[item.Id] = item

	return item, nil
}

func (e *exampleRepository) Get(ctx context.Context, i string) (example.Example, error) {
	if item, ok := e.items[i]; !ok {
		return example.Nil(), notFoundError
	} else {
		return item, nil
	}
}

func (e *exampleRepository) List(ctx context.Context, i string, limit int, page int) ([]example.Example, error) {
	items := e.byUserIndex[i]

	return items, nil
}

func (e *exampleRepository) Delete(ctx context.Context, i string) error {
	item := e.items[i]

	// Delete from user index
	b := e.byUserIndex[item.UserId][:0]
	for _, x := range e.byUserIndex[item.UserId] {
		if x.Id != i {
			b = append(b, x)
		}
	}
	e.byUserIndex[item.UserId] = b

	delete(e.items, i)

	return nil
}

// MARK: SQL
func serializer(val *example.Example) (string, error) {
	b, err := json.Marshal(val)
	return string(b), err
}

func deserializer(s string) (*example.Example, error) {
	var result example.Example
	if err := json.Unmarshal([]byte(s), &result); err != nil {
		return nil, err
	}

	return &result, nil
}

type exampleSQLRepository struct {
	pool        *pgxpool.Pool
	cacheClient valkeyaside.TypedCacheAsideClient[example.Example]
}

func NewSQLRepository(pool *pgxpool.Pool, cacheClient valkeyaside.CacheAsideClient) *exampleSQLRepository {
	typedClient := valkeyaside.NewTypedCacheAsideClient(cacheClient, serializer, deserializer)

	return &exampleSQLRepository{
		pool:        pool,
		cacheClient: typedClient,
	}
}

func (e *exampleSQLRepository) refreshCache(ctx context.Context, item *example.Example) error {
	// Force a cache reset by deleting the existing value from the cache
	// This could result in a race condition, in which case the requester
	// Will need to wait until the TTL expires on the record (60s) and request again
	// This is required because valkeyaside has a local in-memory cache to reduce the load on
	// the remote instance
	err := e.cacheClient.Del(ctx, item.Id)
	if err != nil {
		return err
	}

	// TODO: Make repo specific error here...
	cacheVal, err := serializer(item)
	if err != nil {
		return err
	}

	// Update the cache
	err = e.cacheClient.Client().Client().Do(
		ctx,
		e.cacheClient.Client().Client().B().Set().Key(item.Id).Value(cacheVal).Build(),
	).Error()

	if err != nil {
		return err
	}

	return nil
}

func (e *exampleSQLRepository) Add(ctx context.Context, item example.Example) (example.Example, error) {
	var result example.Example
	err := e.pool.QueryRow(ctx, "INSERT INTO examples (message, uid) VALUES ($1, $2) RETURNING *", item.Message, item.UserId).Scan(&result.Id, &result.Message, &result.UserId)

	if err != nil {
		return example.Nil(), err
	}

	return result, nil
}

func (e *exampleSQLRepository) Update(ctx context.Context, item example.Example) (example.Example, error) {
	var result example.Example
	err := e.pool.QueryRow(ctx, "UPDATE examples SET message=$1 WHERE id=$2 RETURNING *", item.Message, item.Id).Scan(&result.Id, &result.Message, &result.UserId)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return example.Nil(), notFoundError
		default:
			return example.Nil(), err
		}
	}

	e.refreshCache(ctx, &result)

	return result, nil
}

func (e *exampleSQLRepository) Get(ctx context.Context, i string) (example.Example, error) {
	result, err := e.cacheClient.Get(ctx, time.Minute, i, func(ctx context.Context, key string) (val *example.Example, err error) {
		var result example.Example
		err = e.pool.QueryRow(ctx, "SELECT * FROM examples WHERE id=$1", i).Scan(&result.Id, &result.Message, &result.UserId)

		if err != nil {
			slog.LogAttrs(ctx, slog.LevelError, notFoundMsg, slog.String(errKey, err.Error()))
			return nil, notFoundError
		}

		return &result, nil
	})

	if err != nil {
		return example.Nil(), err
	}

	return *result, nil
}

func (e *exampleSQLRepository) List(ctx context.Context, userId string, limit int, page int) ([]example.Example, error) {
	offset := (page - 1) * limit

	res, err := e.pool.Query(ctx, "SELECT * FROM examples WHERE uid=$1 ORDER BY id LIMIT $2 OFFSET $3", userId, limit, offset)

	if err != nil {
		return nil, err
	}

	var results []example.Example
	for res.Next() {
		var e example.Example
		res.Scan(&e.Id, &e.Message, &e.UserId)
		results = append(results, e)
	}

	resErr := res.Err()
	if resErr != nil {
		return nil, resErr
	}

	return results, nil
}

func (e *exampleSQLRepository) Delete(ctx context.Context, i string) error {
	var id string
	err := e.pool.QueryRow(ctx, "DELETE FROM examples WHERE id=$1 RETURNING id", i).Scan(&id)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return notFoundError
		default:
			return err
		}
	}

	// Delete from cache
	err = e.cacheClient.Del(ctx, id)

	if err != nil {
		return nil
	}

	return nil
}
