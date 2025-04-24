package auditservice

import (
	"context"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/moonmoon1919/go-api-reference/pkg/events"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		name       string
		event      events.Event
		errMessage string
	}{
		{
			name:       "PassingCase",
			event:      events.NewEvent(uuid.NewString(), example.ExampleCreated{}),
			errMessage: "",
		},
	}

	repo := NewInMemoryAuditLogRepository()
	service := Service{Store: repo}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := service.Add(context.TODO(), tc.event)

			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}

			if errMsg != tc.errMessage {
				t.Errorf("expected error message %s, got %s", tc.errMessage, errMsg)
			}

			// We've only added one item
			// so retrieve the first item from the items in the repo to verify
			firstItem := repo.items[len(repo.items)-1]
			if !reflect.DeepEqual(tc.event, firstItem) {
				t.Errorf("expected %v, found %v", tc.event, firstItem)
			}
		})
	}
}
