package auditservice

import (
	"context"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moonmoon1919/go-api-reference/internal/config"
	"github.com/moonmoon1919/go-api-reference/internal/store"
	"github.com/moonmoon1919/go-api-reference/pkg/events"
	"github.com/moonmoon1919/go-api-reference/pkg/example"
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

func TestIntegrationAuditAddSQLRepository(t *testing.T) {
	if testType != "INTEGRATION" {
		t.Skip()
	}

	tests := []struct {
		name      string
		userId    string
		eventData events.EventData
	}{
		{
			name:   "PassingCase",
			userId: uuid.NewString(),
			eventData: example.ExampleCreated{
				Id: uuid.NewString(),
			},
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
			err := repository.Add(context.TODO(), events.NewEvent(tc.userId, tc.eventData))

			if err != nil {
				t.Errorf("Unexpected error adding event %s", err.Error())
			}

			// Then
			var eventname string
			var entityId string
			var userId string
			err = pool.QueryRow(context.TODO(), "SELECT eventname, entityid, uid FROM auditlog WHERE entityid=$1 AND uid=$2 AND eventname=$3", tc.eventData.EntityId(), tc.userId, tc.eventData.Name()).Scan(&eventname, &entityId, &userId)

			if err != nil {
				t.Errorf("Unexpected error %s", err.Error())
			}

			if eventname != string(tc.eventData.Name()) {
				t.Errorf("Expected e")
			}

			// Clean up
			pool.Exec(context.TODO(), "DELETE FROM auditlog WHERE entityid=$1 AND uid=$2 AND eventname=$3", tc.eventData.EntityId(), tc.userId, tc.eventData.Name())
		})
	}
}
