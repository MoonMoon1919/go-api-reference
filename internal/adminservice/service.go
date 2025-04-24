package adminservice

import (
	"context"
	"errors"
	"log/slog"

	"github.com/moonmoon1919/go-api-reference/internal/bus"
	"github.com/moonmoon1919/go-api-reference/pkg/events"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
	"github.com/moonmoon1919/go-api-reference/pkg/users"
)

const (
	addUserMsg          = "ADMIN_SERVICE_ADD_USER"
	getUserMsg          = "ADMIN_SERVICE_GET_USER"
	listUserMsg         = "ADMIN_SERVICE_LIST_USERS"
	deleteUserMsg       = "ADMIN_SERVICE_DELETE_USER"
	getExampleMsg       = "ADMIN_SERVICE_GET_EXAMPLE"
	listExamplesMsg     = "ADMIN_SERVICE_LIST_EXAMPLE_FOR_USER"
	deleteExampleMsg    = "ADMIN_SERVICE_DELETE_EXAMPLE"
	getEventsForItemMsg = "ADMIN_SERVICE_GET_EVENTS_FOR_ITEM"
	listByEventUserMsg  = "ADMIN_SERVICE_LIST_BY_EVENT_FOR_USER"
	listEventsByUserMsg = "ADMIN_SERVICE_LIST_EVENTS_FOR_USER"

	// Errors
	storeErrorMsg = "STORE_ERROR"
	logKeyId      = "ID"
	logKeyErr     = "ERR"
)

var limitToLargeError = errors.New("maximum limit is 50")
var invalidPageError = errors.New("page must be greater than 0")
var userServiceNotFound = errors.New("user not found")
var exampleServiceNotFound = errors.New("example not found")
var storeError = errors.New("store error")

type Service struct {
	UserStore    UserStorer
	ExampleStore ExampleStorer
	AuditStore   AuditStorer
	Bus          bus.Busser
}

// MARK: Users
func (s Service) AddUser(ctx context.Context, id string) error {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		addUserMsg,
		slog.String(logKeyId, id),
	)

	user := users.NewUserWithId(id)

	err := s.UserStore.Add(ctx, user)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			storeErrorMsg,
			slog.String(logKeyErr, err.Error()),
		)
		return err
	}

	return nil
}

func (s Service) GetUser(ctx context.Context, id string) (users.User, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		getUserMsg,
		slog.String(logKeyId, id),
	)

	user, err := s.UserStore.Get(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, userNotFoundError):
			return users.Nil(), userServiceNotFound
		default:
			slog.LogAttrs(
				ctx,
				slog.LevelError,
				storeErrorMsg,
				slog.String(logKeyErr, err.Error()),
			)
			return users.Nil(), storeError
		}
	}

	return user, nil
}

func (s Service) ListUsers(ctx context.Context, limit, page int) ([]users.User, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		listUserMsg,
	)

	if limit > 50 {
		return nil, limitToLargeError
	}

	if page < 1 {
		return nil, invalidPageError
	}

	retrievedUsers, err := s.UserStore.List(ctx, limit, page)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			storeErrorMsg,
			slog.String(logKeyErr, err.Error()),
		)
		return make([]users.User, 0), nil
	}

	return retrievedUsers, nil
}

func (s Service) DeleteUser(ctx context.Context, id string) error {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		deleteUserMsg,
		slog.String(logKeyId, id),
	)

	err := s.UserStore.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, userNotFoundError):
			slog.LogAttrs(
				ctx,
				slog.LevelInfo,
				notFoundMsg,
				slog.String(logKeyId, id),
			)
			return userServiceNotFound
		default:
			slog.LogAttrs(
				ctx,
				slog.LevelError,
				storeErrorMsg,
				slog.String(logKeyErr, err.Error()),
			)
			return err
		}
	}

	return nil
}

// MARK: Examples
func (s Service) GetExample(ctx context.Context, id string) (example.Example, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		getExampleMsg,
		slog.String(logKeyId, id),
	)

	item, err := s.ExampleStore.Get(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, exampleNotFoundError):
			slog.LogAttrs(
				ctx,
				slog.LevelInfo,
				notFoundMsg,
				slog.String(logKeyId, id),
			)

			return example.Nil(), exampleServiceNotFound
		default:
			slog.LogAttrs(
				ctx,
				slog.LevelError,
				storeErrorMsg,
				slog.String(logKeyErr, err.Error()),
			)
			return example.Nil(), storeError
		}

	}

	return item, nil
}

func (s Service) GetExamplesForUser(ctx context.Context, id string, limit, page int) ([]example.Example, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		listExamplesMsg,
		slog.String(logKeyId, id),
	)

	if limit > 50 {
		return nil, limitToLargeError
	}

	if page < 1 {
		return nil, invalidPageError
	}

	items, err := s.ExampleStore.GetForUser(ctx, id, limit, page)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			storeErrorMsg,
			slog.String(logKeyErr, err.Error()),
		)
		return nil, err
	}

	return items, nil
}

func (s Service) DeleteExample(ctx context.Context, userId, id string) error {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		deleteExampleMsg,
		slog.String(logKeyId, id),
	)

	err := s.ExampleStore.Delete(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, exampleNotFoundError):
			slog.LogAttrs(
				ctx,
				slog.LevelInfo,
				notFoundMsg,
				slog.String(logKeyId, id),
			)

			return exampleServiceNotFound
		default:
			slog.LogAttrs(
				ctx,
				slog.LevelError,
				storeErrorMsg,
				slog.String(logKeyErr, err.Error()),
			)
			return err
		}

	}

	s.Bus.Notify(events.NewEvent(
		userId,
		example.ExampleDeleted{Id: id},
	))

	return nil
}

// MARK: Audit
func (s Service) GetEventsForItem(ctx context.Context, itemId string, limit, page int) ([]events.Event, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		getEventsForItemMsg,
		slog.String(logKeyId, itemId),
	)

	if limit > 50 {
		return nil, limitToLargeError
	}

	if page < 1 {
		return nil, invalidPageError
	}
	data, err := s.AuditStore.GetEventsForItem(ctx, itemId, limit, page)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			storeErrorMsg,
			slog.String(logKeyErr, err.Error()),
		)
		return nil, storeError
	}

	return data, nil
}

func (s Service) GetEventsForUser(ctx context.Context, userId string, limit, page int) ([]events.Event, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		listEventsByUserMsg,
		slog.String(logKeyId, userId),
	)

	if limit > 50 {
		return nil, limitToLargeError
	}

	if page < 1 {
		return nil, invalidPageError
	}

	data, err := s.AuditStore.GetEventsForUser(ctx, userId, limit, page)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			storeErrorMsg,
			slog.String(logKeyErr, err.Error()),
		)
		return nil, storeError
	}

	return data, nil
}

func (s Service) GetByEventAndUser(ctx context.Context, userId, eventName string, limit, page int) ([]events.Event, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		listByEventUserMsg,
		slog.String("userId", userId),
		slog.String("eventName", eventName),
	)

	if limit > 50 {
		return nil, limitToLargeError
	}

	if page < 1 {
		return nil, invalidPageError
	}

	data, err := s.AuditStore.GetByEventAndUser(ctx, userId, eventName, limit, page)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			storeErrorMsg,
			slog.String(logKeyErr, err.Error()),
		)
		return nil, storeError
	}

	return data, nil
}
