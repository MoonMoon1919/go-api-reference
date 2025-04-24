/*
Integration tests for SQL store
*/

package exampleservice

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moonmoon1919/go-api-reference/internal/cache"
	"github.com/moonmoon1919/go-api-reference/internal/config"
	"github.com/moonmoon1919/go-api-reference/internal/store"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
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

func TestIntegrationExampleAddSQLRepository(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name       string
		userId     string
		message    string
		errMessage string
	}{
		{
			name:       "PassingCase",
			userId:     uuid.NewString(),
			message:    "string",
			errMessage: "",
		},
	}

	cfg := buildConfig()
	pool, cache, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewSQLRepository(pool, cache)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Insert the user so we don't get FK errors
			pool.Exec(context.TODO(), "INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", tc.userId)

			item, _ := example.New(tc.userId, tc.message)
			res, err := repository.Add(context.TODO(), item)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if tc.errMessage != errMessage {
				t.Errorf("Got unexpected error %s, expected %s", errMessage, tc.errMessage)
			}

			// Only validate if we're not checking error cases
			if tc.errMessage == "" {
				if res.Id == "" {
					t.Errorf("Exected example to have id")
				}

				if res.Message != tc.message {
					t.Errorf("Expected message %s, got %s", tc.message, res.Message)
				}

				if res.UserId != tc.userId {
					t.Errorf("Expected message %s, got %s", tc.userId, res.UserId)
				}
			}

			// Clean up by deleting the user, triggering a cascading delete
			pool.Exec(context.TODO(), "DELETE FROM users where id=$1", tc.userId)
		})
	}
}

func TestIntegrationExampleGetSQLRepository(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name       string
		userId     string
		message    string
		errMessage string
	}{
		{
			name:       "PassingCase",
			userId:     uuid.NewString(),
			message:    "string",
			errMessage: "",
		},
	}

	cfg := buildConfig()
	pool, cache, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewSQLRepository(pool, cache)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// Insert the user so we don't get FK errors
			pool.Exec(context.TODO(), "INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", tc.userId)

			item, _ := example.New(tc.userId, tc.message)
			res, err := repository.Add(context.TODO(), item)
			if err != nil {
				t.Errorf("Unexpected error adding example %s", err.Error())
			}

			// When
			retrievedItem, err := repository.Get(context.TODO(), res.Id)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			// Then
			if tc.errMessage != errMessage {
				t.Errorf("Got unexpected error %s, expected %s", errMessage, tc.errMessage)
			}

			// Only validate if we're not checking error cases
			if tc.errMessage == "" {
				if retrievedItem.Id != res.Id {
					t.Errorf("Non-matching id retrieved. Expected %s, got %s", res.Id, retrievedItem.Id)
				}

				if retrievedItem.Message != tc.message {
					t.Errorf("Expected message %s, got %s", tc.message, retrievedItem.Message)
				}

				if retrievedItem.UserId != tc.userId {
					t.Errorf("Expected message %s, got %s", tc.userId, retrievedItem.UserId)
				}
			}

			// Clean up by deleting the user, triggering a cascading delete
			pool.Exec(context.TODO(), "DELETE FROM users where id=$1", tc.userId)
		})
	}
}

func TestIntegrationExampleListSQLRepository(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name       string
		userId     string
		message    string
		numItems   int
		limit      int
		page       int
		errMessage string
	}{
		{
			name:       "PassingCase",
			userId:     uuid.NewString(),
			message:    "string",
			numItems:   50,
			limit:      25,
			page:       1,
			errMessage: "",
		},
	}

	cfg := buildConfig()
	pool, cache, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewSQLRepository(pool, cache)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// Insert the user so we don't get FK errors
			pool.Exec(context.TODO(), "INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", tc.userId)

			for range tc.numItems {
				item, _ := example.New(tc.userId, tc.message)
				_, err := repository.Add(context.TODO(), item)
				if err != nil {
					t.Errorf("Unexpected error adding example %s", err.Error())
				}
			}

			// When
			results, err := repository.List(context.TODO(), tc.userId, tc.limit, tc.page)

			// Then
			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if tc.errMessage != errMessage {
				t.Errorf("Got unexpected error %s, expecting %s", errMessage, tc.errMessage)
			}

			if tc.errMessage == "" {
				if len(results) != tc.limit {
					t.Errorf("Expected %d items to be returned, got %d", tc.limit, len(results))
				}
			}

			// Clean up by deleting the user, triggering a cascading delete
			pool.Exec(context.TODO(), "DELETE FROM users where id=$1", tc.userId)
		})
	}
}

func TestIntegrationExampleUpdateSQLRepository(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name            string
		userId          string
		originalMessage string
		updateMessage   string
		errMessage      string
	}{
		{
			name:            "PassingCase",
			userId:          uuid.NewString(),
			originalMessage: "string",
			updateMessage:   "nu-string",
			errMessage:      "",
		},
	}

	cfg := buildConfig()
	pool, cache, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewSQLRepository(pool, cache)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// Insert the user so we don't get FK errors
			pool.Exec(context.TODO(), "INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", tc.userId)

			item, _ := example.New(tc.userId, tc.originalMessage)
			res, err := repository.Add(context.TODO(), item)
			if err != nil {
				t.Errorf("Unexpected error adding example %s", err.Error())
			}

			// When
			res.SetMessage(tc.updateMessage)
			result, err := repository.Update(context.TODO(), res)

			// Then
			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if tc.errMessage != errMessage {
				t.Errorf("Got unexpected error %s, expecting %s", errMessage, tc.errMessage)
			}

			if tc.errMessage == "" {
				if result.Id != res.Id {
					t.Errorf("Non-matching id retrieved. Expected %s, got %s", res.Id, result.Id)
				}

				if result.Message != tc.updateMessage {
					t.Errorf("Expected message %s, got %s", tc.updateMessage, result.Message)
				}

				if result.UserId != tc.userId {
					t.Errorf("Expected message %s, got %s", tc.userId, result.UserId)
				}
			}

			// Clean up by deleting the user, triggering a cascading delete
			pool.Exec(context.TODO(), "DELETE FROM users where id=$1", tc.userId)
		})
	}
}

func TestIntegrationExampleDeleteSQLRepository(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name       string
		userId     string
		message    string
		errMessage string
	}{
		{
			name:       "PassingCase",
			userId:     uuid.NewString(),
			message:    "string",
			errMessage: "",
		},
	}

	cfg := buildConfig()
	pool, cache, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewSQLRepository(pool, cache)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			// Insert the user so we don't get FK errors
			pool.Exec(context.TODO(), "INSERT INTO users (id) VALUES ($1) ON CONFLICT (id) DO NOTHING", tc.userId)

			item, _ := example.New(tc.userId, tc.message)
			res, err := repository.Add(context.TODO(), item)
			if err != nil {
				t.Errorf("Unexpected error adding example %s", err.Error())
			}

			// When
			err = repository.Delete(context.TODO(), res.Id)

			// Then
			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			if tc.errMessage != errMessage {
				t.Errorf("Got unexpected error %s, expecting %s", errMessage, tc.errMessage)
			}

			_, err = repository.Get(context.TODO(), res.Id)
			if err == nil {
				t.Errorf("Expected getting a deleted item to error, found no error")
			}

			// Clean up users and examples in the event the test fails
			pool.Exec(context.TODO(), "DELETE FROM users WHERE id=$1", tc.userId)
		})
	}
}
