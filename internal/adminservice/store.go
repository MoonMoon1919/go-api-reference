package adminservice

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moonmoon1919/go-api-reference/pkg/events"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
	"github.com/moonmoon1919/go-api-reference/pkg/users"
	"github.com/valkey-io/valkey-go/valkeyaside"
)

const (
	notFoundMsg = "NOT_FOUND"
	userKey     = "USER"
	errKey      = "ERR"
)

var userNotFoundError = errors.New("user not found")
var exampleNotFoundError = errors.New("example not found")

// MARK: Users
type UserStorer interface {
	Add(ctx context.Context, item users.User) error
	Get(ctx context.Context, id string) (users.User, error)
	List(ctx context.Context, limit, page int) ([]users.User, error)
	Delete(ctx context.Context, id string) error
}

type userMemoryStore struct {
	items map[string]users.User
}

func newInMemoryUserStore() *userMemoryStore {
	return &userMemoryStore{
		items: make(map[string]users.User),
	}
}

func (u *userMemoryStore) Add(ctx context.Context, item users.User) error {
	u.items[item.Id] = item

	return nil
}

func (u *userMemoryStore) Get(ctx context.Context, id string) (users.User, error) {
	if item, ok := u.items[id]; !ok {
		return users.Nil(), userNotFoundError
	} else {
		return item, nil
	}
}

func (u *userMemoryStore) List(ctx context.Context, _, _ int) ([]users.User, error) {
	var results []users.User

	for _, usr := range u.items {
		results = append(results, usr)
	}

	return results, nil
}

func (u *userMemoryStore) Delete(ctx context.Context, id string) error {
	if _, ok := u.items[id]; !ok {
		return errors.New("user not found")
	}

	delete(u.items, id)

	return nil
}

// SQL
type userSQLRepository struct {
	pool *pgxpool.Pool
}

func NewUserSQLRepository(pool *pgxpool.Pool) *userSQLRepository {
	return &userSQLRepository{
		pool: pool,
	}
}

func (u *userSQLRepository) Add(ctx context.Context, item users.User) error {
	_, err := u.pool.Exec(ctx, "INSERT INTO users (id) values ($1) ON CONFLICT (id) DO NOTHING", item.Id)

	if err != nil {
		return err
	}

	return nil
}

func (u *userSQLRepository) Get(ctx context.Context, id string) (users.User, error) {
	var user users.User
	err := u.pool.QueryRow(ctx, "SELECT * FROM users WHERE id=$1", id).Scan(&user.Id)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return users.Nil(), userNotFoundError
		default:
			return users.Nil(), err
		}
	}

	return user, nil
}

func (u *userSQLRepository) List(ctx context.Context, limit, page int) ([]users.User, error) {
	offset := (page - 1) * limit

	res, err := u.pool.Query(ctx, "SELECT * FROM users ORDER BY id LIMIT $1 OFFSET $2", limit, offset)

	var results []users.User
	for res.Next() {
		var u users.User
		res.Scan(&u.Id)
		results = append(results, u)
	}

	resErr := res.Err()
	if resErr != nil {
		return nil, err
	}

	return results, err
}

/*
Deletes a user AND PERFORMS A CASCADING DELETE!
*/
func (u *userSQLRepository) Delete(ctx context.Context, id string) error {
	var i string
	err := u.pool.QueryRow(ctx, "DELETE FROM users where id=$1 RETURNING id", id).Scan(&i)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return userNotFoundError
		default:
			return err
		}
	}

	return nil
}

// MARK: Examples
type ExampleStorer interface {
	Get(ctx context.Context, id string) (example.Example, error)
	GetForUser(ctx context.Context, userId string, limit, page int) ([]example.Example, error)
	Delete(ctx context.Context, id string) error
}

type exampleMemoryStore struct {
	items       map[string]example.Example
	byUserIndex map[string][]example.Example
}

func newInMemoryExampleStore() *exampleMemoryStore {
	return &exampleMemoryStore{
		items:       make(map[string]example.Example),
		byUserIndex: make(map[string][]example.Example),
	}
}

// TESTING ONLY!
func (e *exampleMemoryStore) add(ctx context.Context, item example.Example) (example.Example, error) {
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

func (e *exampleMemoryStore) Get(ctx context.Context, id string) (example.Example, error) {
	if item, ok := e.items[id]; !ok {
		return example.Nil(), exampleNotFoundError
	} else {
		return item, nil
	}
}

func (e *exampleMemoryStore) GetForUser(ctx context.Context, userId string, _, _ int) ([]example.Example, error) {
	items := e.byUserIndex[userId]

	return items, nil
}

func (e *exampleMemoryStore) Delete(ctx context.Context, id string) error {
	if _, ok := e.items[id]; !ok {
		return exampleNotFoundError
	}

	item := e.items[id]

	// Delete from user index
	b := e.byUserIndex[item.UserId][:0]
	for _, x := range e.byUserIndex[item.UserId] {
		if x.Id != id {
			b = append(b, x)
		}
	}
	e.byUserIndex[item.UserId] = b

	delete(e.items, id)

	return nil
}

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
	cacheClient valkeyaside.CacheAsideClient
}

func NewExampleSQLRepository(pool *pgxpool.Pool, cacheClient valkeyaside.CacheAsideClient) *exampleSQLRepository {
	return &exampleSQLRepository{
		pool:        pool,
		cacheClient: cacheClient,
	}
}

/*
Retrieves a single example for an administrator

Admins cannot see message content of "example" objects.
*/
func (e *exampleSQLRepository) Get(ctx context.Context, id string) (example.Example, error) {
	var result example.Example
	err := e.pool.QueryRow(ctx, "SELECT id, uid FROM examples WHERE id=$1", id).Scan(&result.Id, &result.UserId)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return example.Nil(), exampleNotFoundError
		default:
			return example.Nil(), err
		}
	}

	return result, nil
}

func (e *exampleSQLRepository) GetForUser(ctx context.Context, userId string, limit, page int) ([]example.Example, error) {
	offset := (page - 1) * limit

	res, err := e.pool.Query(ctx, "SELECT id, uid FROM examples WHERE uid=$1 ORDER BY id LIMIT $2 OFFSET $3", userId, limit, offset)
	if err != nil {
		return make([]example.Example, 0), err
	}

	var results []example.Example
	for res.Next() {
		var e example.Example
		res.Scan(&e.Id, &e.UserId)
		results = append(results, e)
	}

	resErr := res.Err()
	if resErr != nil {
		return make([]example.Example, 0), resErr
	}

	return results, nil
}

/*
Deletes an example on a users behalf

Removes from cache so we don't serve stale cache records to users
*/
func (e *exampleSQLRepository) Delete(ctx context.Context, id string) error {
	var i string
	err := e.pool.QueryRow(ctx, "DELETE FROM examples WHERE id=$1 RETURNING id", id).Scan(&i)

	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return exampleNotFoundError
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

// MARK: Audit
type AuditStorer interface {
	GetEventsForItem(ctx context.Context, itemId string, limit, page int) ([]events.Event, error)
	GetEventsForUser(ctx context.Context, userId string, limit, page int) ([]events.Event, error)
	GetByEventAndUser(ctx context.Context, userId, eventName string, limit, page int) ([]events.Event, error)
}

type auditLogMemoryStore struct {
	byItemIdx           map[string][]events.Event
	byUserIndex         map[string][]events.Event
	byEventAndUserIndex map[string][]events.Event
}

func newInMemoryAuditLogStore() *auditLogMemoryStore {
	return &auditLogMemoryStore{
		byItemIdx:           make(map[string][]events.Event),
		byUserIndex:         make(map[string][]events.Event),
		byEventAndUserIndex: make(map[string][]events.Event),
	}
}

func (a *auditLogMemoryStore) generateCompositeKey(userId, eventName string) string {
	return fmt.Sprintf("UserId#%s+EventName#%s", userId, eventName)
}

func (a *auditLogMemoryStore) add(e events.Event) error {
	// Add to item idx
	_, entityOk := a.byItemIdx[e.EntityId]
	if !entityOk {
		a.byItemIdx[e.EntityId] = make([]events.Event, 0)
	}
	a.byItemIdx[e.EntityId] = append(a.byItemIdx[e.EntityId], e)

	// Add to userIndex
	_, userOk := a.byUserIndex[e.UserId]
	if !userOk {
		a.byUserIndex[e.EntityId] = make([]events.Event, 0)
	}
	a.byUserIndex[e.UserId] = append(a.byUserIndex[e.UserId], e)

	// Add to event + user index
	compositeKey := a.generateCompositeKey(e.UserId, string(e.Name))
	_, userEventOk := a.byEventAndUserIndex[compositeKey]
	if !userEventOk {
		a.byEventAndUserIndex[compositeKey] = make([]events.Event, 0)
	}
	a.byEventAndUserIndex[compositeKey] = append(a.byEventAndUserIndex[compositeKey], e)

	return nil
}

func (a *auditLogMemoryStore) GetEventsForItem(ctx context.Context, itemId string, limit, page int) ([]events.Event, error) {
	items := a.byItemIdx[itemId]

	return items, nil
}

func (a *auditLogMemoryStore) GetEventsForUser(ctx context.Context, userId string, limit, page int) ([]events.Event, error) {
	items := a.byUserIndex[userId]

	return items, nil
}

func (a *auditLogMemoryStore) GetByEventAndUser(ctx context.Context, userId, eventName string, limit, page int) ([]events.Event, error) {
	compositeKey := a.generateCompositeKey(userId, eventName)

	items := a.byEventAndUserIndex[compositeKey]

	return items, nil
}

type auditLogSQLRepository struct {
	pool *pgxpool.Pool
}

/*
We don't know the event type when scanning a row returned from the DB
so we load into this intermediate struct first.

This allows us to inspect the name of the event, then load it into
the appropriate event struct
*/
type intermediateEvent struct {
	Name      example.ExampleEvent
	UserId    string
	EntityId  string
	Timestamp int64
	Data      interface{}
}

/*
Loads an intermediateEvent into an Event

Determines the value of event.Data by inspecting intermediateEvent.Name
then parsing the value of intermediateEvent.Data as the appropriate
event type
*/
func (i intermediateEvent) ToEvent(ctx context.Context) (events.Event, error) {
	var e events.Event

	e.Name = i.Name
	e.UserId = i.UserId
	e.EntityId = i.EntityId
	e.Timestamp = i.Timestamp

	// Marshalling and unmarshalling to json is smelly
	// But we'll live with it for now
	eventDataStr, err := json.Marshal(i.Data)
	if err != nil {
		return events.Nil(), err
	}

	switch i.Name {
	case example.ExampleCreatedEvent:
		var ev example.ExampleCreated

		if err := json.Unmarshal(eventDataStr, &ev); err != nil {
			return events.Nil(), err
		}
		e.Data = ev
	case example.ExampleUpdatedEvent:
		var ev example.ExampleUpdated

		if err := json.Unmarshal(eventDataStr, &ev); err != nil {
			return events.Nil(), err
		}
		e.Data = ev
	case example.ExampleDeletedEvent:
		var ev example.ExampleDeleted

		if err := json.Unmarshal(eventDataStr, &ev); err != nil {
			return events.Nil(), err
		}
		e.Data = ev
	default:
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"INTERMEDIATE_EVENT_TO_EVENT_ERROR",
			slog.String("UNKOWN_EVENT_TYPE", string(i.Name)),
		)
		return events.Nil(), errors.New("unknown example event type")
	}

	return e, nil
}

func NewAuditLogSQLRepository(pool *pgxpool.Pool) *auditLogSQLRepository {
	return &auditLogSQLRepository{
		pool: pool,
	}
}

func (a *auditLogSQLRepository) loadResults(ctx context.Context, rows pgx.Rows) ([]events.Event, error) {
	var results []events.Event
	for rows.Next() {
		var e intermediateEvent

		err := rows.Scan(&e.Name, &e.UserId, &e.EntityId, &e.Timestamp, &e.Data)
		if err != nil {
			return nil, err
		}

		loadedEvent, err := e.ToEvent(ctx)
		if err != nil {
			return nil, err
		}

		results = append(results, loadedEvent)
	}

	rowsErr := rows.Err()
	if rowsErr != nil {
		return nil, rowsErr
	}

	return results, nil
}

func (a *auditLogSQLRepository) GetEventsForItem(ctx context.Context, itemId string, limit, page int) ([]events.Event, error) {
	offset := (page - 1) * limit

	rows, err := a.pool.Query(
		ctx,
		"SELECT eventname, uid, entityid, timestamp, event FROM auditlog WHERE entityid=$1 ORDER BY timestamp LIMIT $2 OFFSET $3",
		itemId,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	return a.loadResults(ctx, rows)
}

func (a *auditLogSQLRepository) GetEventsForUser(ctx context.Context, userId string, limit, page int) ([]events.Event, error) {
	offset := (page - 1) * limit

	rows, err := a.pool.Query(
		ctx,
		"SELECT eventname, uid, entityid, timestamp, event FROM auditlog WHERE uid=$1 ORDER BY timestamp LIMIT $2 OFFSET $3",
		userId,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	return a.loadResults(ctx, rows)
}

func (a *auditLogSQLRepository) GetByEventAndUser(ctx context.Context, userId, eventName string, limit, page int) ([]events.Event, error) {
	offset := (page - 1) * limit

	rows, err := a.pool.Query(
		ctx,
		"SELECT eventname, uid, entityid, timestamp, event FROM auditlog WHERE uid=$1 AND eventName=$2 ORDER BY timestamp LIMIT $3 OFFSET $4",
		userId,
		eventName,
		limit,
		offset,
	)
	if err != nil {
		return nil, err
	}

	return a.loadResults(ctx, rows)
}
