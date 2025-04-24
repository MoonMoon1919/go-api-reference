package auditservice

import (
	"context"
	"log/slog"

	"github.com/moonmoon1919/go-api-reference/pkg/events"
)

type Service struct {
	Store Storer
}

func (s Service) Add(ctx context.Context, e events.Event) error {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"EVENT_RECEIVED",
		slog.Any("event", e),
	)
	err := s.Store.Add(ctx, e)

	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			"STORE_ERROR",
			slog.String("err", err.Error()),
		)
		return err
	}

	return nil
}
