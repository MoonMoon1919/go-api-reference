package userservice

import (
	"context"
	"errors"
	"log/slog"

	"github.com/moonmoon1919/go-api-reference/pkg/users"
)

const (
	userServiceAdd    = "USER_SERVICE_ADD"
	userServiceDelete = "USER_SERVICE_DELETE"
	logKeyId          = "ID"
	userNotFoundMsg   = "USER_NOT_FOUND"
)

type Service struct {
	Store Storer
}

func (us Service) Add(ctx context.Context, id string) error {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		userServiceAdd,
		slog.String(logKeyId, id),
	)

	user := users.NewUserWithId(id)

	_, err := us.Store.Add(ctx, user)
	if err != nil {
		return err
	}

	return nil
}

func (us Service) Delete(ctx context.Context, id string) error {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		userServiceDelete,
		slog.String(logKeyId, id),
	)

	err := us.Store.Delete(ctx, id)

	if err != nil {
		switch {
		case errors.Is(err, notFoundError):
			// If we cannot find the user the event is not retryable
			// So instead of letting something retry and fail
			// We log the error and return nil
			slog.LogAttrs(ctx, slog.LevelError, userNotFoundMsg, slog.String(logKeyId, id))
			return nil
		default:
			return err
		}
	}

	return nil
}
