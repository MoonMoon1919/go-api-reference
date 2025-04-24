package exampleservice

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/moonmoon1919/go-api-reference/internal/bus"
	"github.com/moonmoon1919/go-api-reference/pkg/events"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
)

// MARK: Logging constants
const (
	exampleServiceAdd    = "EXAMPLE_SERVICE_ADD"
	exampleServiceGet    = "EXAMPLE_SERVICE_GET"
	exampleServiceList   = "EXAMPLE_SERVICE_LIST"
	exampleServicePatch  = "EXAMPLE_SERVICE_PATCH"
	exampleServiceDelete = "EXAMPLE_SERVICE_DELETE"
	storeError           = "STORE_ERROR"
	domainError          = "DOMAIN_ERROR"
	logKeyMessage        = "message"
	logKeyId             = "id"
	logKeyUserId         = "uid"
	logKeyError          = "ERROR"
)

// MARK: Errors
type InvalidMessageError struct {
	wrappedErr error
}

func (i InvalidMessageError) Error() string {
	return fmt.Sprintf("invalid message. error: %s", i.wrappedErr.Error())
}

var serviceError = errors.New("service layer error")
var repositoryAddError = errors.New("error storing record")
var repositoryNotFoundError = errors.New("item not found")
var RepositoryListError = errors.New("error listing items")
var limitToLargeError = errors.New("maximum limit is 50")
var invalidPageError = errors.New("page must be greater than 0")

// MARK: Service
type Service struct {
	Store Storer
	Bus   bus.Busser
}

// MARK: Add
func (e Service) Add(ctx context.Context, userId, message string) (example.Example, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		exampleServiceAdd,
		slog.String(logKeyMessage, message),
	)

	item, err := example.New(userId, message)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			domainError,
			slog.String(logKeyError, err.Error()),
		)
		return example.Nil(), &InvalidMessageError{wrappedErr: err}
	}

	storedItem, err := e.Store.Add(ctx, item)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			storeError,
			slog.String(logKeyError, err.Error()),
		)
		return example.Nil(), repositoryAddError
	}

	e.Bus.Notify(
		events.NewEvent(
			userId,
			example.ExampleCreated{Id: storedItem.Id},
		),
	)

	return storedItem, nil
}

// MARK: Update
func (e Service) Update(ctx context.Context, userId, id, message string) (example.Example, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		exampleServicePatch,
		slog.String(logKeyMessage, message),
	)

	item, err := e.Store.Get(ctx, id)

	// Check if the error is not found or something else
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			storeError,
			slog.String(logKeyError, err.Error()),
		)
		// item is example.Nil, so dont bother creating another empty struct
		return item, repositoryNotFoundError
	}

	err = item.SetMessage(message)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			domainError,
			slog.String(logKeyError, err.Error()),
		)
		return example.Nil(), &InvalidMessageError{wrappedErr: err}
	}

	storedItem, err := e.Store.Update(ctx, item)
	if err != nil {
		switch {
		case errors.Is(err, notFoundError):
			slog.LogAttrs(
				ctx,
				slog.LevelInfo,
				notFoundMsg,
				slog.String(logKeyId, id),
			)
			// storedItem is example.Nil, no need to create an empty struct again
			return storedItem, repositoryNotFoundError
		default:
			slog.LogAttrs(
				ctx,
				slog.LevelError,
				storeError,
				slog.String(logKeyError, err.Error()),
			)
			// storedItem is example.Nil, no need to create an empty struct again
			return storedItem, repositoryAddError
		}
	}

	e.Bus.Notify(events.NewEvent(
		userId,
		example.ExampleUpdated{Id: storedItem.Id},
	))

	return storedItem, nil
}

// MARK: DELETE
func (e Service) Delete(ctx context.Context, userId, id string) error {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		exampleServiceDelete,
		slog.String(logKeyId, id),
	)

	err := e.Store.Delete(ctx, id)

	if err != nil {
		switch {
		case errors.Is(err, notFoundError):
			slog.LogAttrs(
				ctx,
				slog.LevelInfo,
				notFoundMsg,
				slog.String(logKeyId, id),
			)
			return repositoryNotFoundError
		default:
			slog.LogAttrs(
				ctx,
				slog.LevelError,
				storeError,
				slog.String(logKeyError, err.Error()),
			)
			return err
		}
	}

	e.Bus.Notify(events.NewEvent(
		userId,
		example.ExampleDeleted{Id: id},
	))
	return nil
}

// MARK: GET
func (e Service) Get(ctx context.Context, id string) (example.Example, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		exampleServiceGet,
		slog.String(logKeyId, id),
	)

	item, err := e.Store.Get(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, notFoundError):
			slog.LogAttrs(
				ctx,
				slog.LevelInfo,
				notFoundMsg,
				slog.String(logKeyId, id),
			)
			return example.Nil(), repositoryNotFoundError
		default:
			slog.LogAttrs(
				ctx,
				slog.LevelError,
				storeError,
				slog.String(logKeyError, err.Error()),
			)
			return example.Nil(), err
		}
	}

	return item, nil
}

// MARK: LIST
func (e Service) List(ctx context.Context, userId string, limit int, page int) ([]example.Example, error) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		exampleServiceList,
		slog.String(logKeyUserId, userId),
	)

	// Maximimum limit we support is 50
	// Chosen arbitrarily to keep load on DB to "something reasonable"
	if limit > 50 {
		return nil, limitToLargeError
	}

	if page < 1 {
		return nil, invalidPageError
	}

	res, err := e.Store.List(ctx, userId, limit, page)
	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			storeError,
			slog.String(logKeyError, err.Error()),
		)
		return []example.Example{}, RepositoryListError
	}

	return res, err
}
