package userservice

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestUserAdd(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "PassingCase",
			id:   uuid.NewString(),
		},
	}

	repo := NewInMemoryUserRepository()
	service := Service{Store: repo}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := service.Add(context.TODO(), tc.id)

			if err != nil {
				t.Errorf("expected no error, got %d", err)
			}
		})
	}
}

func TestUserDelete(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{
			name: "PassingCase",
			id:   uuid.NewString(),
		},
	}

	repo := NewInMemoryUserRepository()
	service := Service{Store: repo}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := service.Add(context.TODO(), tc.id)

			if err != nil {
				t.Errorf("expected no error on Add(), got %d", err)
			}

			err = service.Delete(context.TODO(), tc.id)

			if err != nil {
				t.Errorf("expected no error on Delete(), got %d", err)
			}
		})
	}
}
