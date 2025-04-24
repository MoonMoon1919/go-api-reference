package adminservice

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/moonmoon1919/go-api-reference/internal/bus"
	"github.com/moonmoon1919/go-api-reference/pkg/events"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
)

// MARK: Users
func TestAddUser(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "PassingCase",
			id:   uuid.NewString(),
		},
	}

	userStore := newInMemoryUserStore()
	exampleStore := newInMemoryExampleStore()
	service := Service{UserStore: userStore, ExampleStore: exampleStore}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := service.AddUser(context.TODO(), tc.id)

			if err != nil {
				t.Errorf("expected no error, got %s", err.Error())
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		errMessage string
		create     bool
	}{
		{
			name:       "PassingCase",
			id:         uuid.NewString(),
			errMessage: "",
			create:     true,
		},
		{
			name:       "FailingCase",
			id:         uuid.NewString(),
			errMessage: "user not found",
			create:     false,
		},
	}

	userStore := newInMemoryUserStore()
	exampleStore := newInMemoryExampleStore()
	service := Service{UserStore: userStore, ExampleStore: exampleStore}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.create {
				err := service.AddUser(context.TODO(), tc.id)

				if err != nil {
					t.Errorf("expected no error adding user, got %s", err.Error())
				}
			}

			retrievedItem, err := service.GetUser(context.TODO(), tc.id)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMessage)
			}

			// Only validate a match if we pre-created the user
			if tc.create {
				if retrievedItem.Id != tc.id {
					t.Errorf("expected user with id %s, got %s", tc.id, retrievedItem.Id)
				}
			}
		})
	}
}

func TestListUsers(t *testing.T) {
	tests := []struct {
		name       string
		numItems   int
		errMessage string
		limit      int
		page       int
	}{
		{
			name:       "PassingCase-SomeItems",
			numItems:   2,
			errMessage: "",
			limit:      10,
			page:       1,
		},
		{
			name:       "PassingCase-NoItems",
			numItems:   0,
			errMessage: "",
			limit:      10,
			page:       1,
		},
		{
			name:       "FailingCase-TooManyItems",
			numItems:   0,
			errMessage: "maximum limit is 50",
			limit:      51,
			page:       1,
		},
		{
			name:       "FailingCase-InvalidPage",
			numItems:   0,
			errMessage: "page must be greater than 0",
			limit:      49,
			page:       0,
		},
	}

	for _, tc := range tests {
		// Create these here so we don't have to clear the repos manually
		userStore := newInMemoryUserStore()
		exampleStore := newInMemoryExampleStore()
		service := Service{UserStore: userStore, ExampleStore: exampleStore}

		t.Run(tc.name, func(t *testing.T) {
			for range tc.numItems {
				err := service.AddUser(context.TODO(), uuid.NewString())

				if err != nil {
					t.Errorf("got unexpected error %s wheen adding user", err.Error())
				}
			}

			retrievedItems, err := service.ListUsers(context.TODO(), tc.limit, tc.page)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errMessage {
				t.Errorf("Got unexpected error %s, expecting %s", errMessage, tc.errMessage)
			}

			if len(retrievedItems) != tc.numItems {
				t.Errorf("Found %d items, expected %d", len(retrievedItems), tc.numItems)
			}
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name       string
		id         string
		errMessage string
		create     bool
	}{
		{
			name:       "PassingCase",
			id:         uuid.NewString(),
			errMessage: "",
			create:     true,
		},
		{
			name:       "FailingCase",
			id:         uuid.NewString(),
			errMessage: "user not found",
			create:     false,
		},
	}

	userStore := newInMemoryUserStore()
	exampleStore := newInMemoryExampleStore()
	service := Service{UserStore: userStore, ExampleStore: exampleStore}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.create {
				err := service.AddUser(context.TODO(), tc.id)

				if err != nil {
					t.Errorf("unexpected error adding user, %s", err.Error())
				}
			}

			err := service.DeleteUser(context.TODO(), tc.id)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMessage)
			}
		})
	}
}

// MARK: Examples
func TestGetExample(t *testing.T) {
	tests := []struct {
		name       string
		userId     string
		message    string
		errMessage string
		create     bool
	}{
		{
			name:       "PassingCase",
			message:    "Success",
			userId:     uuid.NewString(),
			errMessage: "",
			create:     true,
		},
		{
			name:       "FailingCase",
			message:    "",
			userId:     uuid.NewString(),
			errMessage: "example not found",
			create:     false,
		},
	}

	userStore := newInMemoryUserStore()
	exampleStore := newInMemoryExampleStore()
	service := Service{UserStore: userStore, ExampleStore: exampleStore}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var id string

			if tc.create {
				item, err := exampleStore.add(context.TODO(), example.Example{
					UserId:  tc.userId,
					Message: tc.message,
				})

				if err != nil {
					t.Errorf("unexpected error adding example %s", err.Error())
				}

				id = item.Id
			}

			retrievedItem, err := service.GetExample(context.TODO(), id)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMessage)
			}

			if retrievedItem.Message != tc.message {
				t.Errorf("expected message %s, got %s", tc.message, retrievedItem.Message)
			}
		})
	}
}

func TestGetExamplesForUser(t *testing.T) {
	tests := []struct {
		name       string
		userId     string
		numItems   int
		errMessage string
		limit      int
		page       int
	}{
		{
			name:       "PassingCase-SomeItems",
			userId:     uuid.NewString(),
			numItems:   10,
			errMessage: "",
			limit:      10,
			page:       1,
		},
		{
			name:       "PassingCase-NoItems",
			userId:     uuid.NewString(),
			numItems:   0,
			errMessage: "",
			limit:      10,
			page:       1,
		},
		{
			name:       "FailingCase-TooManyItems",
			userId:     uuid.NewString(),
			numItems:   0,
			errMessage: "maximum limit is 50",
			limit:      51,
			page:       1,
		},
		{
			name:       "FailingCase-InvalidPage",
			userId:     uuid.NewString(),
			numItems:   0,
			errMessage: "page must be greater than 0",
			limit:      49,
			page:       -10,
		},
	}

	userStore := newInMemoryUserStore()
	exampleStore := newInMemoryExampleStore()
	service := Service{UserStore: userStore, ExampleStore: exampleStore}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create the expected items
			for i := range tc.numItems {
				_, err := exampleStore.add(context.TODO(), example.Example{UserId: tc.userId, Message: fmt.Sprintf("message-%d", i)})

				if err != nil {
					t.Errorf("Unexpected error %s when adding item", err.Error())
				}
			}

			retrievedItems, err := service.GetExamplesForUser(context.TODO(), tc.userId, tc.limit, tc.page)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errMessage {
				t.Errorf("Got unexpected error %s, expected %s", errMessage, tc.errMessage)
			}

			if len(retrievedItems) != tc.numItems {
				t.Errorf("Found %d items, expected %d", len(retrievedItems), tc.numItems)
			}
		})
	}
}

func TestDeleteExample(t *testing.T) {
	tests := []struct {
		name         string
		userId       string
		message      string
		create       bool
		errorMessage string
	}{
		{
			name:         "PassingCase",
			userId:       uuid.NewString(),
			message:      "hello",
			create:       true,
			errorMessage: "",
		},
		{
			name:         "FailingCase-NotFound",
			userId:       uuid.NewString(),
			message:      "",
			create:       false,
			errorMessage: "example not found",
		},
	}

	b := bus.NewFake()
	userStore := newInMemoryUserStore()
	exampleStore := newInMemoryExampleStore()
	service := Service{UserStore: userStore, ExampleStore: exampleStore, Bus: b}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var itemId string
			if tc.create {
				item, err := exampleStore.add(context.TODO(), example.Example{UserId: tc.userId, Message: tc.message})

				if err != nil {
					t.Errorf("unexpected error while adding item %s", err.Error())
				}

				itemId = item.Id
			}

			err := service.DeleteExample(context.TODO(), tc.userId, itemId)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errorMessage {
				t.Errorf("expected error message %s, got %s", tc.errorMessage, errMessage)
			}
		})
	}
}

// MARK: Audit
func TestGetEventsForItem(t *testing.T) {
	tests := []struct {
		name       string
		userId     string
		itemId     string
		numEvents  int
		errMessage string
		limit      int
		page       int
	}{
		{
			name:       "PassingCase",
			userId:     uuid.NewString(),
			itemId:     uuid.NewString(),
			errMessage: "",
			numEvents:  50,
			limit:      50,
			page:       1,
		},
		{
			name:       "FailingCase-TooManyItems",
			userId:     uuid.NewString(),
			itemId:     uuid.NewString(),
			errMessage: "maximum limit is 50",
			numEvents:  0,
			limit:      51,
			page:       1,
		},
		{
			name:       "FailingCase-InvalidPage",
			userId:     uuid.NewString(),
			itemId:     uuid.NewString(),
			errMessage: "page must be greater than 0",
			numEvents:  0,
			limit:      49,
			page:       -10,
		},
	}

	us := newInMemoryUserStore()
	es := newInMemoryExampleStore()
	as := newInMemoryAuditLogStore()
	svc := Service{UserStore: us, ExampleStore: es, AuditStore: as}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			for i := range tc.numEvents {
				if i == 1 {
					as.add(events.NewEvent(tc.userId, example.ExampleCreated{
						Id: tc.itemId,
					}))
				}

				if i < tc.numEvents-1 {
					as.add(events.NewEvent(tc.userId, example.ExampleUpdated{
						Id: tc.itemId,
					}))
				}

				if i == tc.numEvents {
					as.add(events.NewEvent(tc.userId, example.ExampleDeleted{
						Id: tc.itemId,
					}))
				}
			}

			// When
			retrievedItems, err := svc.GetEventsForItem(context.TODO(), tc.itemId, tc.limit, tc.page)

			// Then
			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if tc.errMessage != errMessage {
				t.Errorf("Expected error %s, got %s", tc.errMessage, errMessage)
			}

			if len(retrievedItems) != tc.numEvents {
				t.Errorf("Found %d items, expected %d", len(retrievedItems), tc.numEvents)
			}
		})
	}
}

func TestGetEventsForUser(t *testing.T) {
	tests := []struct {
		name       string
		userId     string
		numEvents  int
		errMessage string
		limit      int
		page       int
	}{
		{
			name:       "PassingCase",
			userId:     uuid.NewString(),
			numEvents:  50,
			errMessage: "",
			limit:      50,
			page:       1,
		},
		{
			name:       "FailingCase-TooManyItems",
			userId:     uuid.NewString(),
			errMessage: "maximum limit is 50",
			numEvents:  0,
			limit:      51,
			page:       1,
		},
		{
			name:       "FailingCase-InvalidPage",
			userId:     uuid.NewString(),
			errMessage: "page must be greater than 0",
			numEvents:  0,
			limit:      49,
			page:       -10,
		},
	}

	us := newInMemoryUserStore()
	es := newInMemoryExampleStore()
	as := newInMemoryAuditLogStore()
	svc := Service{UserStore: us, ExampleStore: es, AuditStore: as}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			for i := range tc.numEvents {
				// We don't care about maintaining relationships between events here
				// We're just looking at getting a mix of events, so we generate
				// IDs randomly
				if i%2 == 0 {
					as.add(events.NewEvent(tc.userId, example.ExampleCreated{
						Id: uuid.NewString(),
					}))
				} else if i%3 == 0 {
					as.add(events.NewEvent(tc.userId, example.ExampleUpdated{
						Id: uuid.NewString(),
					}))
				} else {
					as.add(events.NewEvent(tc.userId, example.ExampleDeleted{
						Id: uuid.NewString(),
					}))
				}
			}

			// When
			retrievedItems, err := svc.GetEventsForUser(context.TODO(), tc.userId, tc.limit, tc.page)

			// Then
			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if tc.errMessage != errMessage {
				t.Errorf("Expected error %s, got %s", tc.errMessage, errMessage)
			}

			if len(retrievedItems) != tc.numEvents {
				t.Errorf("Found %d items, expected %d", len(retrievedItems), tc.numEvents)
			}
		})
	}
}

func TestGetByEventAndUser(t *testing.T) {
	tests := []struct {
		name       string
		userId     string
		errMessage string
		eventType  example.ExampleEvent
		numEvents  int
		limit      int
		page       int
	}{
		{
			name:       "PassingCase",
			userId:     uuid.NewString(),
			errMessage: "",
			eventType:  example.ExampleCreatedEvent,
			numEvents:  10,
			limit:      50,
			page:       1,
		},
		{
			name:       "FailingCase-TooManyItems",
			userId:     uuid.NewString(),
			errMessage: "maximum limit is 50",
			eventType:  example.ExampleCreatedEvent,
			numEvents:  0,
			limit:      51,
			page:       1,
		},
		{
			name:       "FailingCase-InvalidPage",
			userId:     uuid.NewString(),
			errMessage: "page must be greater than 0",
			eventType:  example.ExampleCreatedEvent,
			numEvents:  0,
			limit:      49,
			page:       -10,
		},
	}

	us := newInMemoryUserStore()
	es := newInMemoryExampleStore()
	as := newInMemoryAuditLogStore()
	svc := Service{UserStore: us, ExampleStore: es, AuditStore: as}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			for range tc.numEvents {
				switch tc.eventType {
				case example.ExampleCreatedEvent:
					as.add(events.NewEvent(tc.userId, example.ExampleCreated{
						Id: uuid.NewString(),
					}))
				case example.ExampleUpdatedEvent:
					as.add(events.NewEvent(tc.userId, example.ExampleUpdated{
						Id: uuid.NewString(),
					}))
				case example.ExampleDeletedEvent:
					as.add(events.NewEvent(tc.userId, example.ExampleDeleted{
						Id: uuid.NewString(),
					}))
				}
			}

			// When
			retrievedItems, err := svc.GetByEventAndUser(context.TODO(), tc.userId, string(tc.eventType), tc.limit, tc.page)

			// Then
			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if tc.errMessage != errMessage {
				t.Errorf("Expected error %s, got %s", tc.errMessage, errMessage)
			}

			if len(retrievedItems) != tc.numEvents {
				t.Errorf("Found %d items, expected %d", len(retrievedItems), tc.numEvents)
			}
		})
	}
}
