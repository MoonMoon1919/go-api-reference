package adminservice

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moonmoon1919/go-api-reference/internal/cache"
	"github.com/moonmoon1919/go-api-reference/internal/config"
	"github.com/moonmoon1919/go-api-reference/internal/store"
	"github.com/moonmoon1919/go-api-reference/pkg/events"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
	"github.com/moonmoon1919/go-api-reference/pkg/users"
	"github.com/valkey-io/valkey-go"
	"github.com/valkey-io/valkey-go/valkeyaside"
)

var testType = os.Getenv("TEST_TYPE")

type testConfig struct {
	database store.Config
	cache    cache.Config
}

func buildConfig() testConfig {
	return testConfig{
		database: store.Config{
			Host:     config.NewEnvironmentSource("DB_HOST"),
			User:     config.NewEnvironmentSource("DB_USER"),
			Password: config.NewEnvironmentSource("DB_PASS"),
			Database: config.NewEnvironmentSource("DB_NAME"),
			Schema: config.NewFirst(
				config.NewEnvironmentSource("DB_SCHEMA"),
				config.NewDefaultValueSource("schemas"),
			),
		},
		cache: cache.Config{
			Host: config.NewEnvironmentSource("CACHE_HOST"),
		},
	}
}

func buildClients(cfg testConfig) (*pgxpool.Pool, valkeyaside.CacheAsideClient, error) {
	dbCache, err := valkeyaside.NewClient(valkeyaside.ClientOption{
		ClientOption: valkey.ClientOption{
			InitAddress: []string{cfg.cache.Host.Must()},
			SelectDB:    1,
		},
	})
	if err != nil {
		return nil, nil, err
	}

	dbConfig, err := pgxpool.ParseConfig(cfg.database.ConnectionString())
	if err != nil {
		return nil, nil, err
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, nil, err
	}

	return dbpool, dbCache, nil
}

// MARK: Users
func TestIntegrationAdminUserAdd(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name         string
		userId       string
		errorMessage string
	}{
		{
			name:         "PassingCase",
			userId:       uuid.NewString(),
			errorMessage: "",
		},
	}

	cfg := buildConfig()
	pool, _, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewUserSQLRepository(pool)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// When
			err := repository.Add(context.TODO(), users.NewUserWithId(tc.userId))

			var errorMessage string
			if err != nil {
				errorMessage = err.Error()
			}

			// Then
			if errorMessage != tc.errorMessage {
				t.Errorf("Expected error message %s, got %s", tc.errorMessage, errorMessage)
			}

			_, err = repository.Get(context.TODO(), tc.userId)
			if err != nil {
				t.Errorf("Expected to find user with id %s, got error %s", tc.userId, err.Error())
			}

			// Cleanup
			pool.Exec(context.TODO(), "DELETE FROM users where id=$1", tc.userId)
		})
	}
}

func TestIntegrationAdminUserGet(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name         string
		userId       string
		errorMessage string
	}{
		{
			name:         "PassingCase",
			userId:       uuid.NewString(),
			errorMessage: "",
		},
	}

	cfg := buildConfig()
	pool, _, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewUserSQLRepository(pool)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			err := repository.Add(context.TODO(), users.NewUserWithId(tc.userId))
			if err != nil {
				t.Errorf("Unexpected error aidding user %s", err.Error())
			}

			// When
			resp, err := repository.Get(context.TODO(), tc.userId)

			var errorMessage string
			if err != nil {
				errorMessage = err.Error()
			}

			// Then
			if errorMessage != tc.errorMessage {
				t.Errorf("Expected error message %s, got %s", tc.errorMessage, errorMessage)
			}

			if resp.Id != tc.userId {
				t.Errorf("Expected user with id %s, got user with id %s", tc.userId, resp.Id)
			}

			// Cleanup
			pool.Exec(context.TODO(), "DELETE FROM users where id=$1", tc.userId)
		})
	}
}

func TestIntegrationAdminUserList(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name         string
		numUsers     int
		errorMessage string
	}{
		{
			name:         "PassingCase",
			numUsers:     50,
			errorMessage: "",
		},
	}

	cfg := buildConfig()
	pool, _, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewUserSQLRepository(pool)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			for range tc.numUsers {
				err := repository.Add(context.TODO(), users.NewUserWithId(uuid.NewString()))

				if err != nil {
					t.Errorf("Unexpected error aidding user %s", err.Error())
				}
			}

			// When
			resp, err := repository.List(context.TODO(), tc.numUsers, 1)

			var errorMessage string
			if err != nil {
				errorMessage = err.Error()
			}

			// Then
			if errorMessage != tc.errorMessage {
				t.Errorf("Expected error message %s, got %s", tc.errorMessage, errorMessage)
			}

			if len(resp) != tc.numUsers {
				t.Errorf("Expected to find %d users, got %d", tc.numUsers, len(resp))
			}

			// Clean up
			for _, user := range resp {
				pool.Exec(context.TODO(), "DELETE FROM users where id=$1", user.Id)
			}
		})
	}
}

func TestIntegrationAdminUserDelete(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name         string
		userId       string
		errorMessage string
	}{
		{
			name:         "PassingCase",
			userId:       uuid.NewString(),
			errorMessage: "",
		},
	}

	cfg := buildConfig()
	pool, _, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewUserSQLRepository(pool)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			err := repository.Add(context.TODO(), users.NewUserWithId(tc.userId))
			if err != nil {
				t.Errorf("Unexpected error adding a user %s", err.Error())
			}

			// When
			err = repository.Delete(context.TODO(), tc.userId)

			var errorMessage string
			if err != nil {
				errorMessage = err.Error()
			}

			// Then
			if errorMessage != tc.errorMessage {
				t.Errorf("Expected errror message %s, got %s", tc.errorMessage, errorMessage)
			}

			_, err = repository.Get(context.TODO(), tc.userId)
			if err == nil {
				t.Errorf("Expected getting deleted user to raise an error")
			}

			pool.Exec(context.TODO(), "DELETE FROM users where id=$1", tc.userId)
		})
	}
}

// MARK: Examples
func TestIntegrationAdminExampleGet(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name         string
		userId       string
		message      string
		errorMessage string
	}{
		{
			name:         "PassingCase",
			userId:       uuid.NewString(),
			message:      "test",
			errorMessage: "",
		},
	}

	cfg := buildConfig()
	pool, cache, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewExampleSQLRepository(pool, cache)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// Insert the user for FK constraints
			pool.Exec(context.TODO(), "INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", tc.userId)

			// Insert the item
			var id string
			err := pool.QueryRow(context.TODO(), "INSERT INTO examples (message, uid) VALUES ($1, $2) RETURNING id", tc.message, tc.userId).Scan(&id)

			if err != nil {
				t.Errorf("Unexpected error inserting test example: %s", err.Error())
			}

			// When
			resp, err := repository.Get(context.TODO(), id)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			// Then
			if errMessage != tc.errorMessage {
				t.Errorf("Expected error message %s, got %s", tc.errorMessage, errMessage)
				t.FailNow()
			}

			// Admins don't have access to message content
			if resp.Id != id {
				t.Errorf("Expected Id %s, got %s", id, resp.Id)
			}

			if resp.UserId != tc.userId {
				t.Errorf("Expected UserId %s, got %s", tc.userId, resp.UserId)
			}

			// Cleanup - triggers cascading delete
			pool.Exec(context.TODO(), "DELETE FROM users where id=$1", tc.userId)
		})
	}
}

func TestIntegrationAdminExampleGetForUser(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name         string
		userId       string
		numItems     int
		errorMessage string
	}{
		{
			name:         "PassingCase",
			userId:       uuid.NewString(),
			numItems:     50,
			errorMessage: "",
		},
	}

	cfg := buildConfig()
	pool, cache, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewExampleSQLRepository(pool, cache)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// Insert the user for FK constraints
			pool.Exec(context.TODO(), "INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", tc.userId)

			// Insert the item
			for i := range tc.numItems {
				var id string
				err := pool.QueryRow(context.TODO(), "INSERT INTO examples (message, uid) VALUES ($1, $2) RETURNING id", fmt.Sprintf("message-%d", i), tc.userId).Scan(&id)

				if err != nil {
					t.Errorf("Unexpected error inserting test example: %s", err.Error())
				}
			}

			// When
			resp, err := repository.GetForUser(context.TODO(), tc.userId, tc.numItems, 1)

			var errorMessage string
			if err != nil {
				errorMessage = err.Error()
			}

			// Then
			if errorMessage != tc.errorMessage {
				t.Errorf("Expected error message %s, got %s", tc.errorMessage, errorMessage)
			}

			if len(resp) != tc.numItems {
				t.Errorf("Expected %d items, found %d", tc.numItems, len(resp))
			}

			for _, item := range resp {
				if item.UserId != tc.userId {
					t.Errorf("Expected all items to have user id %s, item %s has user %s", tc.userId, item.Id, item.UserId)
				}
			}

			// Cleanup - triggers cascading delete
			pool.Exec(context.TODO(), "DELETE FROM users where id=$1", tc.userId)
		})
	}
}

func TestIntegrationAdminExampleDelete(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name         string
		userId       string
		errorMessage string
	}{
		{
			name:         "PassingCase",
			userId:       uuid.NewString(),
			errorMessage: "",
		},
	}

	cfg := buildConfig()
	pool, cache, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewExampleSQLRepository(pool, cache)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// Insert the user for FK constraints
			pool.Exec(context.TODO(), "INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", tc.userId)

			// Insert the item
			var id string
			err := pool.QueryRow(context.TODO(), "INSERT INTO examples (message, uid) VALUES ($1, $2) RETURNING id", "cool message", tc.userId).Scan(&id)

			if err != nil {
				t.Errorf("Unexpected error inserting test example: %s", err.Error())
			}

			// When
			err = repository.Delete(context.TODO(), id)

			var errorMessage string
			if err != nil {
				errorMessage = err.Error()
			}

			// Then
			if errorMessage != tc.errorMessage {
				t.Errorf("Expected error message %s, got %s", tc.errorMessage, errorMessage)
			}

			_, err = repository.Get(context.TODO(), id)
			if err == nil {
				t.Errorf("Expected getting a deleted item to error, found no error")
			}

			// Cleanup - triggers cascading delete
			pool.Exec(context.TODO(), "DELETE FROM users where id=$1", tc.userId)
		})
	}
}

// MARK: Audit
func TestIntegrationAdminAuditGetEventsForItem(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name      string
		numEvents int
		userId    string
		entityId  string
	}{}

	cfg := buildConfig()
	pool, _, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewAuditLogSQLRepository(pool)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			for i := range tc.numEvents {
				var eventData events.EventData
				if i == 1 {
					eventData = example.ExampleUpdated{
						Id: tc.entityId,
					}
				} else {
					eventData = example.ExampleUpdated{
						Id: tc.entityId,
					}
				}

				event := events.NewEvent(tc.userId, eventData)
				x, _ := json.Marshal(eventData)

				pool.Exec(context.TODO(), "INSERT INTO auditlog (eventname, uid, entityid, timestamp, event) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (eventname, entityid, uid, timestamp) DO NOTHING",
					event.Name,
					event.UserId,
					event.EntityId,
					event.Timestamp,
					x,
				)
			}

			// When
			resp, err := repository.GetEventsForItem(context.TODO(), tc.entityId, 10, 1)

			if err != nil {
				t.Errorf("Received unexpected error %s", err.Error())
			}

			// Then
			if len(resp) != tc.numEvents {
				t.Errorf("Expected to find %d items, found %d", tc.numEvents, len(resp))
			}

			// Cleanup
			for _, event := range resp {
				pool.Exec(context.TODO(), "DELETE FROM auditlog WHERE eventname=$1 uid=$2 event=$3 timestamp=$4",
					event.Name,
					event.UserId,
					event.EntityId,
					event.Timestamp,
				)
			}
		})
	}
}

func TestIntegrationAdminAuditGetEventsForUser(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name       string
		userId     string
		eventsList []events.EventData
	}{
		{
			name:   "PassingCase-SomeItems",
			userId: uuid.NewString(),
			eventsList: []events.EventData{
				example.ExampleCreated{
					Id: uuid.NewString(),
				},
				example.ExampleUpdated{
					Id: uuid.NewString(),
				},
				example.ExampleDeleted{
					Id: uuid.NewString(),
				},
				example.ExampleCreated{
					Id: uuid.NewString(),
				},
			},
		},
		{
			name:       "PassingCase-NoItems",
			userId:     uuid.NewString(),
			eventsList: []events.EventData{},
		},
	}

	cfg := buildConfig()
	pool, _, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewAuditLogSQLRepository(pool)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			for _, eventData := range tc.eventsList {
				event := events.NewEvent(tc.userId, eventData)
				x, _ := json.Marshal(eventData)

				pool.Exec(context.TODO(), "INSERT INTO auditlog (eventname, uid, entityid, timestamp, event) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (eventname, entityid, uid, timestamp) DO NOTHING",
					event.Name,
					event.UserId,
					event.EntityId,
					event.Timestamp,
					x,
				)
			}

			// When
			resp, err := repository.GetEventsForUser(context.TODO(), tc.userId, 10, 1)

			if err != nil {
				t.Errorf("Received unexpected error %s", err.Error())
			}

			// Then
			if len(resp) != len(tc.eventsList) {
				t.Errorf("Expected to find %d items, found %d", len(tc.eventsList), len(resp))
			}

			// Cleanup
			for _, event := range resp {
				pool.Exec(context.TODO(), "DELETE FROM auditlog WHERE eventname=$1 uid=$2 event=$3 timestamp=$4",
					event.Name,
					event.UserId,
					event.EntityId,
					event.Timestamp,
				)
			}
		})
	}
}

func TestIntegrationAdminAuditGetByEventAndUser(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name       string
		userId     string
		eventName  example.ExampleEvent
		eventsList []events.EventData
	}{
		{
			name:      "PassingCase-SomeMatch",
			userId:    uuid.NewString(),
			eventName: example.ExampleCreatedEvent,
			eventsList: []events.EventData{
				example.ExampleCreated{
					Id: uuid.NewString(),
				},
				example.ExampleUpdated{
					Id: uuid.NewString(),
				},
				example.ExampleDeleted{
					Id: uuid.NewString(),
				},
				example.ExampleCreated{
					Id: uuid.NewString(),
				},
			},
		},
		{
			name:      "PassingCase-NoneMatch",
			userId:    uuid.NewString(),
			eventName: example.ExampleCreatedEvent,
			eventsList: []events.EventData{
				example.ExampleUpdated{
					Id: uuid.NewString(),
				},
				example.ExampleDeleted{
					Id: uuid.NewString(),
				},
			},
		},
	}

	cfg := buildConfig()
	pool, _, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewAuditLogSQLRepository(pool)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			for _, eventData := range tc.eventsList {
				event := events.NewEvent(tc.userId, eventData)
				x, _ := json.Marshal(eventData)

				pool.Exec(context.TODO(), "INSERT INTO auditlog (eventname, uid, entityid, timestamp, event) VALUES ($1, $2, $3, $4, $5) ON CONFLICT (eventname, entityid, uid, timestamp) DO NOTHING",
					event.Name,
					event.UserId,
					event.EntityId,
					event.Timestamp,
					x,
				)
			}

			// When
			resp, err := repository.GetByEventAndUser(context.TODO(), tc.userId, string(tc.eventName), 10, 1)

			if err != nil {
				t.Errorf("Received unexpected error %s", err.Error())
			}

			// Then
			// Filter down the events list to just the events the match the event name
			var filteredList []events.EventData
			for _, eventData := range tc.eventsList {
				if eventData.Name() == tc.eventName {
					filteredList = append(filteredList, eventData)
				}
			}

			if len(resp) != len(filteredList) {
				t.Errorf("Expected to find %d items, found %d", len(filteredList), len(resp))
			}

			// Cleanup
			for _, event := range resp {
				pool.Exec(context.TODO(), "DELETE FROM auditlog WHERE eventname=$1 uid=$2 event=$3 timestamp=$4",
					event.Name,
					event.UserId,
					event.EntityId,
					event.Timestamp,
				)
			}
		})
	}
}
