package adminservice

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/moonmoon1919/go-api-reference/internal/bus"
	"github.com/moonmoon1919/go-api-reference/internal/cache"
	"github.com/moonmoon1919/go-api-reference/internal/middleware"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
	"github.com/moonmoon1919/go-api-reference/pkg/users"
)

// MARK: ADD USER
func TestControllerAddUser(t *testing.T) {
	tests := []struct {
		name           string
		responseWriter *httptest.ResponseRecorder
		expectedStatus int
		userId         string
	}{
		{
			name:           "PassingCase",
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusCreated,
			userId:         uuid.NewString(),
		},
	}

	cache := cache.NewInMemoryCache()
	es := newInMemoryExampleStore()
	us := newInMemoryUserStore()
	as := newInMemoryAuditLogStore()
	service := Service{ExampleStore: es, UserStore: us, AuditStore: as}
	controller := Controller{Service: service, Cache: cache}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			body, err := json.Marshal(CreateUserRequest{
				UserId: tc.userId,
			})
			if err != nil {
				t.Errorf("Got unexpected error marshalling json: %s", err.Error())
			}

			request := httptest.NewRequest(http.MethodPut, "/admin/users", strings.NewReader(string(body)))

			// When
			controller.AddUser(tc.responseWriter, request)

			// Then
			if tc.responseWriter.Code != tc.expectedStatus {
				t.Errorf("expected status code to be %d, got %d", tc.expectedStatus, tc.responseWriter.Code)
			}

			_, err = us.Get(context.TODO(), tc.userId)
			if err != nil {
				t.Errorf("expected to find user in store, got error %s", err.Error())
			}
		})
	}
}

// MARK: GET USER
func TestControllerGetUser(t *testing.T) {
	tests := []struct {
		name           string
		responseWriter *httptest.ResponseRecorder
		expectedStatus int
		userId         string
	}{
		{
			name:           "PassingCase",
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusOK,
			userId:         uuid.NewString(),
		},
	}

	cache := cache.NewInMemoryCache()
	es := newInMemoryExampleStore()
	us := newInMemoryUserStore()
	as := newInMemoryAuditLogStore()
	service := Service{ExampleStore: es, UserStore: us, AuditStore: as}
	controller := Controller{Service: service, Cache: cache}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var expected GetUserResponse
			id := tc.userId

			if tc.expectedStatus != http.StatusNotFound {
				err := service.AddUser(context.TODO(), tc.userId)
				if err != nil {
					t.Errorf("Unexpected error adding user %s", err.Error())
				}

				expected = NewGetUserResponseFromUser(&users.User{
					Id: tc.userId,
				})
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/users/%s", id), nil)
			req.SetPathValue("id", id)

			// When
			controller.GetUser(tc.responseWriter, req)

			// Then
			if tc.responseWriter.Code != tc.expectedStatus {
				t.Errorf("expected status code to be %d, got %d", tc.expectedStatus, tc.responseWriter.Code)
			}

			var actual GetUserResponse
			json.Unmarshal(tc.responseWriter.Body.Bytes(), &actual)

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("expected body to be:\n%v\ngot:\n%v", expected, actual)
			}
		})
	}
}

// MARK: LIST USERS
func TestControllerListUsers(t *testing.T) {
	tests := []struct {
		name           string
		responseWriter *httptest.ResponseRecorder
		request        *http.Request
		expectedStatus int
		numItems       int
	}{
		{
			name:           "PassingCase-50-items",
			responseWriter: httptest.NewRecorder(),
			request:        httptest.NewRequest(http.MethodGet, "/admin/users?limit=50&page=1", nil),
			expectedStatus: http.StatusOK,
			numItems:       50,
		},
		{
			name:           "PassingCase-0-items",
			responseWriter: httptest.NewRecorder(),
			request:        httptest.NewRequest(http.MethodGet, "/admin/users?limit=50&page=1", nil),
			expectedStatus: http.StatusOK,
			numItems:       0,
		},
		{
			name:           "FailingCase-InvalidLimit",
			responseWriter: httptest.NewRecorder(),
			request:        httptest.NewRequest(http.MethodGet, "/admin/users?limit=51&page=1", nil),
			expectedStatus: http.StatusBadRequest,
			numItems:       0,
		},
		{
			name:           "FailingCase-InvalidPage",
			responseWriter: httptest.NewRecorder(),
			request:        httptest.NewRequest(http.MethodGet, "/admin/users?limit=50&page=0", nil),
			expectedStatus: http.StatusBadRequest,
			numItems:       0,
		},
	}

	for _, tc := range tests {
		// Create these here so we don't have to manage resetting the stores
		// Between test cases
		cache := cache.NewInMemoryCache()
		es := newInMemoryExampleStore()
		us := newInMemoryUserStore()
		as := newInMemoryAuditLogStore()
		service := Service{ExampleStore: es, UserStore: us, AuditStore: as}
		controller := Controller{Service: service, Cache: cache}

		t.Run(tc.name, func(t *testing.T) {
			// Given
			for range tc.numItems {
				service.AddUser(context.TODO(), uuid.NewString())
			}

			// When
			controller.ListUsers(tc.responseWriter, tc.request)

			// Then
			if tc.responseWriter.Code != tc.expectedStatus {
				t.Errorf("expected status code to be %d, got %d", tc.expectedStatus, tc.responseWriter.Code)
			}

			// Check the response on positive cases
			if tc.expectedStatus != http.StatusBadRequest {
				var actual ListUserResponse
				json.Unmarshal(tc.responseWriter.Body.Bytes(), &actual)

				if len(actual.Users) != tc.numItems {
					t.Errorf("expected %d users, found %d", tc.numItems, len(actual.Users))
				}
			}

		})
	}
}

// MARK: DELETE USER
func TestControllerDeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		responseWriter *httptest.ResponseRecorder
		expectedStatus int
		userId         string
		validate       bool
	}{
		{
			name:           "PassingCase",
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusNoContent,
			userId:         uuid.NewString(),
			validate:       true,
		},
	}

	es := newInMemoryExampleStore()
	us := newInMemoryUserStore()
	as := newInMemoryAuditLogStore()
	service := Service{ExampleStore: es, UserStore: us, AuditStore: as}
	controller := Controller{Service: service}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			body, err := json.Marshal(CreateUserRequest{
				UserId: tc.userId,
			})
			if err != nil {
				t.Errorf("Got unexpected error marshalling json: %s", err.Error())
			}

			initialResponseWriter := httptest.NewRecorder()
			initialRequest := httptest.NewRequest(http.MethodPut, "/admin/users", strings.NewReader(string(body)))

			controller.AddUser(initialResponseWriter, initialRequest)

			if initialResponseWriter.Code != http.StatusCreated {
				t.Errorf("expected status %d, got %d when putting initial data", http.StatusCreated, initialResponseWriter.Code)
			}

			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/admin/users/%s", tc.userId), nil)
			req.SetPathValue("id", tc.userId)

			// When
			controller.DeleteUser(tc.responseWriter, req)

			// Then
			if tc.validate {
				_, err := us.Get(context.TODO(), tc.userId)

				if err == nil {
					t.Errorf("expected repository to return user not found")
				}
			}
		})
	}
}

// MARK: GET EXAMPLE
func TestControllerGetExample(t *testing.T) {
	tests := []struct {
		name           string
		responseWriter *httptest.ResponseRecorder
		expectedStatus int
		userId         string
	}{
		{
			name:           "PassingCase",
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusOK,
			userId:         uuid.NewString(),
		},
		{
			name:           "FailingCase",
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusNotFound,
			userId:         uuid.NewString(),
		},
	}

	cache := cache.NewInMemoryCache()
	es := newInMemoryExampleStore()
	us := newInMemoryUserStore()
	as := newInMemoryAuditLogStore()
	service := Service{ExampleStore: es, UserStore: us, AuditStore: as}
	controller := Controller{Service: service, Cache: cache}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var expected GetExampleResponse
			var id string = "1234"

			if tc.expectedStatus != http.StatusNotFound {
				d, _ := es.add(context.TODO(), example.Example{UserId: tc.userId, Id: uuid.NewString()})
				id = d.Id
				expected = GetExampleResponse{
					Id:     d.Id,
					UserId: tc.userId,
				}
			}

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/users/%s", id), nil)
			req.SetPathValue("id", id)

			// When
			controller.GetExample(tc.responseWriter, req)

			// Then
			if tc.responseWriter.Code != tc.expectedStatus {
				t.Errorf("expected status code to be %d, got %d", tc.expectedStatus, tc.responseWriter.Code)
			}

			var actual GetExampleResponse
			json.Unmarshal(tc.responseWriter.Body.Bytes(), &actual)

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("expected body to be:\n%v\ngot:\n%v", expected, actual)
			}
		})
	}
}

// MARK: GET EXAMPLES FOR USER
func TestControllerGetExamplesForUser(t *testing.T) {
	tests := []struct {
		name           string
		responseWriter *httptest.ResponseRecorder
		expectedStatus int
		numItems       int
		userId         string
		page           int
		limit          int
	}{
		{
			name:           "PassingCase-SomeItems",
			responseWriter: httptest.NewRecorder(),
			userId:         uuid.NewString(),
			expectedStatus: http.StatusOK,
			numItems:       50,
			limit:          50,
			page:           1,
		},
		{
			name:           "PassingCase-NoItems",
			responseWriter: httptest.NewRecorder(),
			userId:         uuid.NewString(),
			expectedStatus: http.StatusOK,
			numItems:       0,
			limit:          50,
			page:           1,
		},
		{
			name:           "FailingCase-InvalidLimit",
			responseWriter: httptest.NewRecorder(),
			userId:         uuid.NewString(),
			expectedStatus: http.StatusBadRequest,
			numItems:       0,
			limit:          51,
			page:           1,
		},
		{
			name:           "FailingCase-InvalidPage",
			responseWriter: httptest.NewRecorder(),
			userId:         uuid.NewString(),
			expectedStatus: http.StatusBadRequest,
			numItems:       0,
			limit:          50,
			page:           0,
		},
	}

	cache := cache.NewInMemoryCache()
	es := newInMemoryExampleStore()
	us := newInMemoryUserStore()
	as := newInMemoryAuditLogStore()
	service := Service{ExampleStore: es, UserStore: us, AuditStore: as}
	controller := Controller{Service: service, Cache: cache}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/admin/users/%s/examples?page=%d&limit=%d", tc.userId, tc.page, tc.limit), nil)
			req.SetPathValue("id", tc.userId)

			// Given
			for i := range tc.numItems {
				es.add(context.TODO(), example.Example{UserId: tc.userId, Message: fmt.Sprintf("message-%d", i)})
			}

			// When
			controller.GetExamplesForUser(tc.responseWriter, req)

			// Then
			if tc.responseWriter.Code != tc.expectedStatus {
				t.Errorf("expected status code to be %d, got %d", tc.expectedStatus, tc.responseWriter.Code)
			}

			if tc.expectedStatus != http.StatusBadRequest {
				var actual ListExampleResponse
				json.Unmarshal(tc.responseWriter.Body.Bytes(), &actual)

				if len(actual.Items) != tc.numItems {
					t.Errorf("expected %d items, found %d", tc.numItems, len(actual.Items))
				}
			}
		})
	}
}

// MARK: DELETE EXAMPLE
func TestControllerDeleteExample(t *testing.T) {
	tests := []struct {
		name           string
		responseWriter *httptest.ResponseRecorder
		expectedStatus int
		userId         string
		validate       bool
	}{
		{
			name:           "PassingCase",
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusNoContent,
			userId:         uuid.NewString(),
			validate:       true,
		},
		{
			name:           "NotFoundCase",
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusNotFound,
			userId:         uuid.NewString(),
			validate:       false,
		},
	}

	b := bus.NewFake()
	cache := cache.NewInMemoryCache()
	es := newInMemoryExampleStore()
	us := newInMemoryUserStore()
	as := newInMemoryAuditLogStore()
	service := Service{ExampleStore: es, UserStore: us, AuditStore: as, Bus: b}
	controller := Controller{Service: service, Cache: cache}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			var id string = "1234"

			if tc.expectedStatus == http.StatusNoContent {
				ex, err := es.add(context.TODO(), example.Example{UserId: tc.userId})
				if err != nil {
					t.Errorf("unexpected error adding example, %d", err)
				}

				id = ex.Id
			}

			request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/admin/examples/%s", id), nil)
			request.SetPathValue("id", id)

			ctx := middleware.ContextWithUser(request.Context(), middleware.RequestingUser{
				Id:          "123",
				Role:        middleware.AdministratorRole,
				Permissions: middleware.NewPermissionSet([]string{"example::read", "example::create", "example::delete"}),
			})
			r := request.WithContext(ctx)

			// When
			controller.DeleteExample(tc.responseWriter, r)

			// Then
			if tc.responseWriter.Code != tc.expectedStatus {
				t.Errorf("expected status code to be %d, got %d", tc.expectedStatus, tc.responseWriter.Code)
			}

			if tc.validate {
				_, err := es.Get(context.TODO(), id)

				if err == nil {
					t.Errorf("expected repository to return that item was not found, item found")
				}

				// Value is no longer in cache
				_, ok := cache.Get(context.TODO(), id)

				if ok {
					t.Errorf("expected item to no longer be in cache, item in cache")
				}
			}
		})
	}
}

// MARK: GET EVENTS FOR ITEM
func TestControllerGetEventsForItem(t *testing.T) {
	tests := []struct {
		name string
	}{
		{},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given

			// When

			// Then
		})
	}
}

// MARK: GET EVENTS FOR USER
func TestControllerGetEventsForUser(t *testing.T) {
	tests := []struct {
		name string
	}{
		{},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given

			// When

			// Then
		})
	}
}
