package middleware

import (
	"context"
	"testing"
)

func TestContextWithUserUserFromContext(t *testing.T) {
	user := RequestingUser{
		Id:          "123",
		Role:        AdministratorRole,
		Permissions: NewPermissionSet([]string{}),
	}

	ctx := ContextWithUser(context.Background(), user)

	user, ok := UserFromContext(ctx)

	if !ok {
		t.Errorf("Expected user to be in context")
	}

	if user.Id != user.Id {
		t.Errorf("Expected user to be %v, got %v", user, user)
	}

	if user.Role != user.Role {
		t.Errorf("Expected user to be %v, got %v", user, user)
	}
}
