package bus

import (
	"context"
	"fmt"

	"github.com/moonmoon1919/go-api-reference/pkg/events"
)

type FakeBus struct {
	Messages []events.Event
}

func NewFake() *FakeBus {
	return &FakeBus{}
}

func (b *FakeBus) Listen(done <-chan struct{}) {
	fmt.Printf("Not implemented")
}

func (b *FakeBus) Notify(event events.Event) {
	b.Messages = append(b.Messages, event)
}

func (b *FakeBus) CloseAndDrain(ctx context.Context) {
	fmt.Printf("Not implemented")
}
