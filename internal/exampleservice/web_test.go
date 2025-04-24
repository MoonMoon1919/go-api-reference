package exampleservice

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
)

type path struct {
	key   string
	value string
}

// MARK: GET
func TestControllerGet(t *testing.T) {
	tests := []struct {
		name               string
		request            *http.Request
		responseWriter     *httptest.ResponseRecorder
		expectedStatus     int
		message            string
		userId             string
		checkCacheResponse bool
	}{
		{
			name:               "PassingCase",
			request:            httptest.NewRequest(http.MethodGet, "/examples/123", nil),
			responseWriter:     httptest.NewRecorder(),
			expectedStatus:     http.StatusOK,
			message:            "HELLO_EXAMPLE_SERVICE",
			userId:             uuid.NewString(),
			checkCacheResponse: false,
		},
		{
			name:               "CacheCase",
			request:            httptest.NewRequest(http.MethodGet, "/examples/123", nil),
			responseWriter:     httptest.NewRecorder(),
			expectedStatus:     http.StatusOK,
			message:            "HELLO_EXAMPLE_SERVICE",
			userId:             uuid.NewString(),
			checkCacheResponse: true,
		},
		{
			name:               "FailingCase",
			request:            httptest.NewRequest(http.MethodGet, "/examples/123", nil),
			responseWriter:     httptest.NewRecorder(),
			expectedStatus:     http.StatusNotFound,
			userId:             uuid.NewString(),
			message:            "",
			checkCacheResponse: false,
		},
	}

	cache := cache.NewInMemoryCache()
	repo := NewInMemoryExampleRepository()
	b := bus.NewFake()
	service := Service{Store: repo, Bus: b}
	controller := Controller{Service: service, Cache: cache}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var expected GetExampleResponse
			id := "1234"

			// If the expected status is not found, dont precreate the object
			if tc.expectedStatus != http.StatusNotFound {
				d, _ := service.Add(context.TODO(), tc.userId, tc.message)
				expected = NewGetExampleResponseFromExample(d)
				id = d.Id
			}

			tc.request.SetPathValue("id", id)

			controller.Get(tc.responseWriter, tc.request)

			if tc.responseWriter.Code != tc.expectedStatus {
				t.Errorf("expected status code to be %d, got %d", tc.expectedStatus, tc.responseWriter.Code)
			}

			var actual GetExampleResponse
			json.Unmarshal(tc.responseWriter.Body.Bytes(), &actual)

			if !reflect.DeepEqual(expected, actual) {
				t.Errorf("expected body to be:\n%v\ngot:\n%v", expected, actual)
			}

			if tc.checkCacheResponse {
				etag := tc.responseWriter.Header().Get("etag")

				nuRequest := httptest.NewRequest(http.MethodGet, "/examples/123", nil)
				nuWriter := httptest.NewRecorder()
				nuRequest.Header.Set("If-None-Match", etag)
				nuRequest.SetPathValue("id", id)

				controller.Get(nuWriter, nuRequest)

				if nuWriter.Code != http.StatusNotModified {
					t.Errorf("expected not modified status code, got %d", nuWriter.Code)
				}
			}
		})
	}
}

// MARK: CREATE
func TestControllerCreate(t *testing.T) {
	tests := []struct {
		name           string
		request        *http.Request
		responseWriter *httptest.ResponseRecorder
		expectedStatus int
	}{
		{
			name:           "PassingCase",
			request:        httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(`{"message": "yay"}`)),
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "EmptyBody",
			request:        httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(`{}`)),
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "InvalidRequestBody",
			request:        httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(`{"invalid": "request"}`)),
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusBadRequest,
		},
	}

	cache := cache.NewInMemoryCache()
	repo := NewInMemoryExampleRepository()
	b := bus.NewFake()
	service := Service{Store: repo, Bus: b}
	controller := Controller{Service: service, Cache: cache}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx := middleware.ContextWithUser(tc.request.Context(), middleware.RequestingUser{
				Id:          "123",
				Role:        middleware.AdministratorRole,
				Permissions: middleware.NewPermissionSet([]string{"example::read", "example::create", "example::delete"}),
			})
			r := tc.request.WithContext(ctx)

			controller.Create(tc.responseWriter, r)

			if tc.responseWriter.Code != tc.expectedStatus {
				t.Errorf("expected status code to be %d, got %d", tc.expectedStatus, tc.responseWriter.Code)
			}
		})
	}
}

// MARK: PATCH
func TestControllerPatch(t *testing.T) {
	tests := []struct {
		name            string
		responseWriter  *httptest.ResponseRecorder
		initialValue    string
		updatedValue    string
		expectedStatus  int
		etagVal         string
		validateThruGet bool
	}{
		{
			name:            "PassingCase",
			responseWriter:  httptest.NewRecorder(),
			initialValue:    "dude",
			updatedValue:    "sweet",
			expectedStatus:  http.StatusOK,
			etagVal:         "",
			validateThruGet: true,
		},
		{
			name:            "Precondition failed",
			responseWriter:  httptest.NewRecorder(),
			initialValue:    "dude",
			updatedValue:    "sweet",
			etagVal:         "fakelolz",
			expectedStatus:  http.StatusPreconditionFailed,
			validateThruGet: false,
		},
	}

	cache := cache.NewInMemoryCache()
	repo := NewInMemoryExampleRepository()
	b := bus.NewFake()
	service := Service{Store: repo, Bus: b}
	controller := Controller{Service: service, Cache: cache}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			initialResponseWriter := httptest.NewRecorder()
			initialBody := fmt.Sprintf(`{"message": "%s"}`, tc.initialValue)
			initialRequest := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(initialBody))

			ctx := middleware.ContextWithUser(initialRequest.Context(), middleware.RequestingUser{
				Id:          "123",
				Role:        middleware.AdministratorRole,
				Permissions: middleware.NewPermissionSet([]string{"example::read", "example::create", "example::delete"}),
			})
			ir := initialRequest.WithContext(ctx)

			controller.Create(initialResponseWriter, ir)

			if initialResponseWriter.Code != 201 {
				t.Errorf("expected status 201, got %d when putting initial data", initialResponseWriter.Code)
			}

			etag := initialResponseWriter.Header().Get("Etag")

			var response CreateExampleResponse
			if err := json.NewDecoder(initialResponseWriter.Body).Decode(&response); err != nil {
				t.Errorf("error decoding create example response, %d", err)
			}

			// The actual test
			body := fmt.Sprintf(`{"message": "%s"}`, tc.updatedValue)
			request := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/examples/%s", response.Id), strings.NewReader(body))
			request.SetPathValue("id", response.Id)

			nuCtx := middleware.ContextWithUser(request.Context(), middleware.RequestingUser{
				Id:          "123",
				Role:        middleware.AdministratorRole,
				Permissions: middleware.NewPermissionSet([]string{"example::read", "example::create", "example::delete"}),
			})
			patchRequest := request.WithContext(nuCtx)

			if len(tc.etagVal) != 0 {
				request.Header.Set("If-Match", tc.etagVal)
			} else {
				request.Header.Set("If-Match", etag)
			}

			controller.Patch(tc.responseWriter, patchRequest)

			if tc.responseWriter.Code != tc.expectedStatus {
				t.Errorf("expected status code to be %d, got %d", tc.expectedStatus, tc.responseWriter.Code)
			}

			// Validate that the data was set
			if tc.validateThruGet {
				retrievedData, _ := service.Get(context.TODO(), response.Id)

				if retrievedData.Message != tc.updatedValue {
					t.Errorf("expected updated data to be %s, got %s", tc.updatedValue, retrievedData.Message)
				}
			}
		})
	}
}

// MARK: DELETE
func TestControllerDelete(t *testing.T) {
	tests := []struct {
		name           string
		body           string
		responseWriter *httptest.ResponseRecorder
		expectedStatus int
		validate       bool
	}{
		{
			name:           "PassingCase",
			body:           `{"message": "dude, sweet"}`,
			responseWriter: httptest.NewRecorder(),
			expectedStatus: http.StatusNoContent,
			validate:       true,
		},
	}

	cache := cache.NewInMemoryCache()
	repo := NewInMemoryExampleRepository()
	b := bus.NewFake()
	service := Service{Store: repo, Bus: b}
	controller := Controller{Service: service, Cache: cache}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			initialResponseWriter := httptest.NewRecorder()
			initialRequest := httptest.NewRequest(http.MethodPost, "/examples", strings.NewReader(tc.body))

			ctx := middleware.ContextWithUser(initialRequest.Context(), middleware.RequestingUser{
				Id:          "123",
				Role:        middleware.AdministratorRole,
				Permissions: middleware.NewPermissionSet([]string{"example::read", "example::create", "example::delete"}),
			})
			ir := initialRequest.WithContext(ctx)

			controller.Create(initialResponseWriter, ir)

			if initialResponseWriter.Code != http.StatusCreated {
				t.Errorf("expected status 201, got %d when putting initial data", initialResponseWriter.Code)
			}

			var response CreateExampleResponse
			if err := json.NewDecoder(initialResponseWriter.Body).Decode(&response); err != nil {
				t.Errorf("error decoding create example response, %d", err)
			}

			// When
			request := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/examples/%s", response.Id), nil)
			request.SetPathValue("id", response.Id)
			nuCtx := middleware.ContextWithUser(request.Context(), middleware.RequestingUser{
				Id:          "123",
				Role:        middleware.AdministratorRole,
				Permissions: middleware.NewPermissionSet([]string{"example::read", "example::create", "example::delete"}),
			})
			deleteRequest := request.WithContext(nuCtx)

			controller.Delete(tc.responseWriter, deleteRequest)

			// Then
			if tc.responseWriter.Code != tc.expectedStatus {
				t.Errorf("expected status code to be %d, got %d", tc.expectedStatus, tc.responseWriter.Code)
			}

			if tc.validate {
				// Value is no longer in repo
				_, err := repo.Get(context.TODO(), response.Id)

				if err == nil {
					t.Errorf("expected repository to return that item was not found, item found")
				}

				// Value is no longer in cache
				_, ok := cache.Get(context.TODO(), response.Id)

				if ok {
					t.Errorf("expected item to no longer be in cache, item in cache")
				}
			}
		})
	}
}

// MARK: LIST
func TestControllerList(t *testing.T) {
	tests := []struct {
		name           string
		responseWriter *httptest.ResponseRecorder
		request        *http.Request
		expectedStatus int
		userId         string
		numItems       int
	}{
		{
			name:           "PassingCase-50-items",
			responseWriter: httptest.NewRecorder(),
			request:        httptest.NewRequest(http.MethodGet, "/examples?limit=50&page=1", nil),
			expectedStatus: http.StatusOK,
			userId:         uuid.NewString(),
			numItems:       50,
		},
		{
			name:           "PassingCase-0-items",
			responseWriter: httptest.NewRecorder(),
			request:        httptest.NewRequest(http.MethodGet, "/examples?limit=50&page=1", nil),
			expectedStatus: http.StatusOK,
			userId:         uuid.NewString(),
			numItems:       0,
		},
		{
			name:           "FailingCase-InvalidLimit",
			responseWriter: httptest.NewRecorder(),
			request:        httptest.NewRequest(http.MethodGet, "/examples?limit=51&page=1", nil),
			expectedStatus: http.StatusBadRequest,
			userId:         uuid.NewString(),
			numItems:       0,
		},
		{
			name:           "FailingCase-InvalidPage",
			responseWriter: httptest.NewRecorder(),
			request:        httptest.NewRequest(http.MethodGet, "/examples?limit=50&page=0", nil),
			expectedStatus: http.StatusBadRequest,
			userId:         uuid.NewString(),
			numItems:       0,
		},
	}

	for _, tc := range tests {
		// Create these here so we don't have to manage resetting the stores
		// Between test cases
		cache := cache.NewInMemoryCache()
		repo := NewInMemoryExampleRepository()
		b := bus.NewFake()
		service := Service{Store: repo, Bus: b}
		controller := Controller{Service: service, Cache: cache}

		t.Run(tc.name, func(t *testing.T) {
			// Given
			ctx := middleware.ContextWithUser(tc.request.Context(), middleware.RequestingUser{
				Id:          tc.userId,
				Role:        middleware.AdministratorRole,
				Permissions: middleware.NewPermissionSet([]string{"example::read", "example::create", "example::delete"}),
			})
			requestWithContext := tc.request.WithContext(ctx)

			for i := range tc.numItems {
				service.Add(context.TODO(), tc.userId, fmt.Sprintf("message-%d", i))
			}

			// When
			controller.List(tc.responseWriter, requestWithContext)

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
