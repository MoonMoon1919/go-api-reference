package userservice

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moonmoon1919/go-api-reference/internal/config"
	"github.com/moonmoon1919/go-api-reference/internal/store"
	"github.com/moonmoon1919/go-api-reference/pkg/users"
)

var testType = os.Getenv("TEST_TYPE")

type testConfig struct {
	database store.Config
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
	}
}

func buildClients(cfg testConfig) (*pgxpool.Pool, error) {

	dbConfig, err := pgxpool.ParseConfig(cfg.database.ConnectionString())
	if err != nil {
		return nil, err
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		return nil, err
	}

	return dbpool, nil
}

func TestIntegrationUserAddSQLRepository(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name       string
		userId     string
		errMessage string
	}{
		{
			name:       "PassingCase",
			userId:     uuid.NewString(),
			errMessage: "",
		},
	}

	cfg := buildConfig()
	pool, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewSQLRepository(pool)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// When
			_, err := repository.Add(context.TODO(), users.NewUserWithId(tc.userId))

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			// Then
			if errMessage != tc.errMessage {
				t.Errorf("Expected error message %s, got %s", tc.errMessage, errMessage)
			}

			var id string
			err = pool.QueryRow(context.TODO(), "SELECT id FROM users WHERE id=$1", tc.userId).Scan(&id)

			if err != nil {
				t.Errorf("Unexpected error checking for inserted user %s", err.Error())
			}

			if id != tc.userId {
				t.Errorf("Expected user id %s, got id %s", tc.userId, id)
			}

			// Clean up
			pool.Exec(context.TODO(), "DELETE FROM users WHERE id=$1", tc.userId)
		})
	}
}
func TestIntegrationUserDeleteSQLRepository(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name       string
		userId     string
		create     bool
		errMessage string
	}{
		{
			name:       "PassingCase",
			userId:     uuid.NewString(),
			create:     true,
			errMessage: "",
		},
		{
			name:       "UserNotFound",
			userId:     uuid.NewString(),
			create:     false,
			errMessage: "user not found",
		},
	}

	cfg := buildConfig()
	pool, err := buildClients(cfg)
	if err != nil {
		t.Errorf("Unexpected error building clients %s", err.Error())
	}
	defer pool.Close()

	repository := NewSQLRepository(pool)

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Given
			if tc.create {
				_, err := repository.Add(context.TODO(), users.NewUserWithId(tc.userId))
				if err != nil {
					t.Errorf("Error adding user %s", err.Error())
				}
			}

			// When
			err = repository.Delete(context.TODO(), tc.userId)

			var errMessage string
			if err != nil {
				errMessage = err.Error()
			}

			// Then
			if errMessage != tc.errMessage {
				t.Errorf("Expected error %s, got %s", tc.errMessage, errMessage)
			}

			if tc.create {
				rows, _ := pool.Query(context.TODO(), "SELECT id FROM users WHERE id=$1", tc.userId)
				// We should have no rows returned
				if rows.Next() != false {
					t.Errorf("Expected to found 0 users with id %s, found at least 1", tc.userId)
				}
			}
		})
	}
}
