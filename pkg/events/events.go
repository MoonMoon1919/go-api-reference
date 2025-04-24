package events

import (
	"encoding/json"
	"time"

	"github.com/moonmoon1919/go-api-reference/pkg/example"
)

const emptyString = ""

type EventData interface {
	Name() example.ExampleEvent
	EntityId() string
}

type Event struct {
	Name      example.ExampleEvent
	UserId    string
	EntityId  string
	Timestamp int64
	Data      EventData
}

func NewEvent(userId string, data EventData) Event {
	return Event{
		Name:      data.Name(),
		UserId:    userId,
		EntityId:  data.EntityId(),
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
}

func Nil() Event {
	return Event{}
}

func (e Event) String() (string, error) {
	s, err := json.Marshal(e)

	if err != nil {
		return emptyString, err
	}

	return string(s), err
}
