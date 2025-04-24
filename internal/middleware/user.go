package middleware

import (
	"context"
	"net/http"

	"github.com/moonmoon1919/go-api-reference/internal/responses"
)

/*
Key for storing the user in the context

We use an empty struct as the key value to avoid allocations
*/
type userKey struct{}

/*
Role of the user
*/
type Role string

const (
	AdministratorRole Role = "Administrator"
	UserRole          Role = "User"
)

/*
User information retrieved from the JWT token in the request
*/
type RequestingUser struct {
	Id          string
	Role        Role
	Permissions PermissionSet
}

/*
Add the requesting user to the context
*/
func ContextWithUser(ctx context.Context, user RequestingUser) context.Context {
	return context.WithValue(ctx, userKey{}, user)
}

/*
Get the requesting user from the context
*/
func UserFromContext(ctx context.Context) (RequestingUser, bool) {
	user, ok := ctx.Value(userKey{}).(RequestingUser)
	return user, ok
}

/*
Get the requesting user from the request
*/
func getUserFromRequest(_ *http.Request) (RequestingUser, error) {
	// TEMPORARY
	return RequestingUser{
		Id:   "f697115f-f723-4c45-8301-e482a21dfd89",
		Role: AdministratorRole,
		Permissions: NewPermissionSet([]string{
			"example::read",
			"example::create",
			"example::delete",
			"admin::example::read",
			"admin::example::delete",
			"admin::user::read",
			"admin::user::create",
			"admin::user::delete",
			"admin::auditlog::read",
		}),
	}, nil
}

// MARK: Middleware
type userMiddleware struct {
	next http.HandlerFunc
}

func (m *userMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromRequest(r)
	if err != nil {
		responses.WriteUnauthorizedResponse(w)
		return
	}

	r = r.WithContext(ContextWithUser(r.Context(), user))
	m.next.ServeHTTP(w, r)
}

func InsertRequestingUser(next http.HandlerFunc) http.Handler {
	return &userMiddleware{next: next}
}
