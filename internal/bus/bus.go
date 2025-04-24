package bus

import (
	"context"
	"log/slog"
	"sync"

	"github.com/moonmoon1919/go-api-reference/pkg/events"
)

const QUEUE_CLOSED = "QUEUE_CLOSED"

type Subscriber func(event events.Event) error
type Subscribers []Subscriber

type Busser interface {
	Listen(done <-chan struct{})
	Notify(event events.Event)
	CloseAndDrain(ctx context.Context)
}

type Bus struct {
	ch          chan events.Event
	subscribers Subscribers
	wg          sync.WaitGroup
}

func New(subscribers Subscribers) Bus {
	return Bus{
		ch:          make(chan events.Event),
		subscribers: subscribers,
		wg:          sync.WaitGroup{},
	}
}

func (b *Bus) Listen(done <-chan struct{}) {
	for {
		select {
		case <-done:
			return
		case v := <-b.ch:
			for _, subscriber := range b.subscribers {
				b.wg.Add(1)
				go func(v events.Event, sub Subscriber) {
					defer b.wg.Done()
					sub(v)
				}(v, subscriber)
			}
		}
	}
}

func (b *Bus) CloseAndDrain(ctx context.Context) {
	// Drain the work queue by closing it and processing all remaining items
	close(b.ch)
	slog.LogAttrs(ctx, slog.LevelInfo, QUEUE_CLOSED)

	for v := range b.ch {
		for _, subscriber := range b.subscribers {
			b.wg.Add(1)
			go func(v events.Event, sub Subscriber) {
				defer b.wg.Done()
				sub(v)
			}(v, subscriber)
		}
	}

	// Wait for all in flight work to finish
	doneChan := make(chan struct{}, 1)
	go func() {
		b.wg.Wait()
		close(doneChan)
	}()

	select {
	case <-ctx.Done():
		return
	case <-doneChan:
	}
}

func (b *Bus) Notify(event events.Event) {
	b.ch <- event
}
