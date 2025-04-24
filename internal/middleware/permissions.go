package middleware

import (
	"errors"
	"net/http"

	"github.com/moonmoon1919/go-api-reference/internal/responses"
)

type Exists struct{}

type PermissionSet map[string]Exists

func NewPermissionSet(permissions []string) PermissionSet {
	perms := make(PermissionSet, len(permissions))

	for _, permission := range permissions {
		perms[permission] = Exists{}
	}

	return perms
}

// MARK: Validators
type Validator interface {
	Validate(permissions PermissionSet) error
}

// MARK: HasAll
type HasAll struct {
	permissions PermissionSet
}

func NewHasAll(permissions []string) HasAll {
	return HasAll{
		permissions: NewPermissionSet(permissions),
	}
}

/*
Validates that a requesting user has all of the permissions in the set
*/
func (h HasAll) Validate(permissions PermissionSet) error {
	expectedNumPermissions := len(h.permissions)
	found := make(map[string]Exists, expectedNumPermissions)

	for permission := range h.permissions {
		_, ok := permissions[permission]

		if ok {
			found[permission] = Exists{}
		}
	}

	if len(found) != expectedNumPermissions {
		return errors.New("MISSING_PERMISSION")
	}

	return nil
}

// MARK: HasOne
type HasOne struct {
	permissions PermissionSet
}

func NewHasOne(permissions []string) HasOne {
	return HasOne{
		permissions: NewPermissionSet(permissions),
	}
}

/*
Validates that a requesting user has at least one of the permissions in the set
*/
func (h HasOne) Validate(permissions PermissionSet) error {
	for permission := range h.permissions {
		_, ok := permissions[permission]

		if ok {
			return nil
		}
	}

	return errors.New("MISSING_PERMISSION")
}

// MARK: Has
type Has struct {
	permission string
}

func NewHas(permission string) Has {
	return Has{
		permission: permission,
	}
}

/*
Validates that a requesting user has a specific permission
*/
func (h Has) Validate(permissions PermissionSet) error {
	_, ok := permissions[h.permission]

	if ok {
		return nil
	}

	return errors.New("MISSING_PERMISSION")
}

// MARK: ValidationMiddleware
func PermissionValidationMiddleware(validator Validator) func(http.HandlerFunc) http.HandlerFunc {
	return func(h http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, ok := UserFromContext(r.Context())

			if !ok {
				responses.WriteUnauthorizedResponse(w)
				return
			}

			err := validator.Validate(user.Permissions)

			if err != nil {
				responses.WriteUnauthorizedResponse(w)
				return
			}

			h.ServeHTTP(w, r)
		}
	}
}
