package exampleservice

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/moonmoon1919/go-api-reference/internal/bus"
)

// MARK: Add
func TestExampleAdd(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		userId     string
		errMessage string
	}{
		{
			name:       "PassingCase",
			userId:     uuid.NewString(),
			message:    "Hello",
			errMessage: "",
		},
		{
			name:       "FailingCase",
			userId:     uuid.NewString(),
			message:    "",
			errMessage: "invalid message. error: message length must be greater than 0",
		},
	}

	b := bus.NewFake()
	repo := NewInMemoryExampleRepository()
	service := Service{Store: repo, Bus: b}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item, err := service.Add(context.TODO(), tc.userId, tc.message)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMessage)
			}

			retrievedItem, _ := service.Get(context.TODO(), item.Id)
			if retrievedItem.Message != tc.message {
				t.Errorf("expected message %s, got %s", retrievedItem.Message, tc.message)
			}
		})
	}
}

// MARK: Update
func TestExampleUpdate(t *testing.T) {
	tests := []struct {
		name           string
		initialMessage string
		userId         string
		updatedMessage string
		errMessage     string
	}{
		{
			name:           "PassingCase",
			initialMessage: "Hi",
			userId:         uuid.NewString(),
			updatedMessage: "Bye",
			errMessage:     "",
		},
		{
			name:           "FailingCase",
			initialMessage: "Hi",
			userId:         uuid.NewString(),
			updatedMessage: "",
			errMessage:     "invalid message. error: message length must be greater than 0",
		},
	}

	b := bus.NewFake()
	repo := NewInMemoryExampleRepository()
	service := Service{Store: repo, Bus: b}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item, err := service.Add(context.TODO(), tc.userId, tc.initialMessage)

			if err != nil {
				t.Errorf("unexpected error message %s", err)
			}

			updatedItem, err := service.Update(context.TODO(), tc.userId, item.Id, tc.updatedMessage)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if errMessage != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMessage)
			}

			if updatedItem.Message != tc.updatedMessage {
				t.Errorf("expected message %s, got %s", tc.updatedMessage, updatedItem.Message)
			}
		})
	}
}

// MARK: Delete
func TestExampleDelete(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		userId     string
		errMessage string
	}{
		{
			name:    "PassingCase",
			userId:  uuid.NewString(),
			message: "Hi",
		},
	}

	b := bus.NewFake()
	repo := NewInMemoryExampleRepository()
	service := Service{Store: repo, Bus: b}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			item, err := service.Add(context.TODO(), tc.userId, tc.message)

			if err != nil {
				t.Errorf("unexpeced error %s", err)
			}

			err = service.Delete(context.TODO(), tc.userId, item.Id)

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

// MARK: Get
func TestExampleGet(t *testing.T) {
	tests := []struct {
		name       string
		message    string
		userId     string
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
			errMessage: "item not found",
			create:     false,
		},
	}

	b := bus.NewFake()
	repo := NewInMemoryExampleRepository()
	service := Service{Store: repo, Bus: b}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			itemId := uuid.NewString()

			if tc.create {
				item, _ := service.Add(context.TODO(), tc.userId, tc.message)
				itemId = item.Id
			}

			retrievedItem, err := service.Get(context.TODO(), itemId)

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

// MARK: List
func TestList(t *testing.T) {
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
			numItems:   2,
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

	b := bus.NewFake()
	repo := NewInMemoryExampleRepository()
	service := Service{Store: repo, Bus: b}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create the expected items
			for i := range tc.numItems {
				_, err := service.Add(context.TODO(), tc.userId, fmt.Sprintf("Item number %d", i+1))

				if err != nil {
					t.Errorf("Got unexpected error %s when adding item", err.Error())
				}
			}

			retrievedItems, err := service.List(context.TODO(), tc.userId, tc.limit, tc.page)

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
