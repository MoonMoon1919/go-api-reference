package adminservice

import (
	"github.com/moonmoon1919/go-api-reference/pkg/events"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
	"github.com/moonmoon1919/go-api-reference/pkg/users"
)

type GetUserResponse struct {
	Id string `json:"id"`
}

func NewGetUserResponseFromUser(user *users.User) GetUserResponse {
	return GetUserResponse{
		Id: user.Id,
	}
}

type ListUserResponse struct {
	Users []GetUserResponse `json:"users"`
}

func NewListUsersResponseFromUsers(users *[]users.User) ListUserResponse {
	items := make([]GetUserResponse, len(*users))

	for idx, i := range *users {
		items[idx] = NewGetUserResponseFromUser(&i)
	}

	return ListUserResponse{
		Users: items,
	}
}

type GetExampleResponse struct {
	Id     string `json:"id"`
	UserId string `json:"message"`
}

func NewGetExampleResponseFromExample(e example.Example) GetExampleResponse {
	return GetExampleResponse{
		Id:     e.Id,
		UserId: e.UserId,
	}
}

type ListExampleResponse struct {
	Items []GetExampleResponse `json:"items"`
}

func NewListExampleResponseFromExamples(e []example.Example) ListExampleResponse {
	items := make([]GetExampleResponse, len(e))

	for idx, i := range e {
		items[idx] = NewGetExampleResponseFromExample(i)
	}

	return ListExampleResponse{
		Items: items,
	}
}

type EventResponse struct {
	Name      example.ExampleEvent `json:"name"`
	UserId    string               `json:"user_id"`
	EntityId  string               `json:"entity_id"`
	Timestamp int64                `json:"timestamp"`
	Data      events.EventData     `json:"data"`
}

func NewEventResponseFromEvent(e *events.Event) EventResponse {
	return EventResponse{
		Name:      e.Name,
		UserId:    e.UserId,
		EntityId:  e.EntityId,
		Timestamp: e.Timestamp,
		Data:      e.Data,
	}
}

type ListEventResponse struct {
	Events []EventResponse `json:"events"`
}

func NewListEventsFromEventsList(e *[]events.Event) ListEventResponse {
	events := make([]EventResponse, len(*e))

	for idx, i := range *e {
		events[idx] = NewEventResponseFromEvent(&i)
	}

	return ListEventResponse{
		Events: events,
	}
}
