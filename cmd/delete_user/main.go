package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/moonmoon1919/go-api-reference/internal/config"
	"github.com/moonmoon1919/go-api-reference/internal/store"
	"github.com/moonmoon1919/go-api-reference/internal/userservice"
)

const (
	deleteUserError = "DELETE_USER_ERROR"
	keyError        = "error"
)

var (
	logger     *slog.Logger
	logContext = context.Background()
	Usage      = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}
)

type appConfig struct {
	database store.Config
}

func realMain(service userservice.Service, ctx context.Context, id string) error {
	err := service.Delete(ctx, id)

	if err != nil {
		slog.LogAttrs(
			ctx,
			slog.LevelError,
			deleteUserError,
			slog.String(keyError, err.Error()),
		)
		return err
	}

	return nil
}

func main() {
	// flags implementation for the sake of simplicity
	var uid string
	flag.StringVar(&uid, "user-id", "", "the id for the user you are adding")

	flag.Parse()

	if len(uid) == 0 {
		fmt.Println("Missing required field 'user-id'")
		fmt.Println("")

		Usage()
		return
	}

	cfg := appConfig{
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

	// MARK: Repository
	dbConfig, err := pgxpool.ParseConfig(cfg.database.ConnectionString())
	if err != nil {
		panic(err)
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		panic(err)
	}

	repo := userservice.NewSQLRepository(dbpool)

	// MARK: Service
	service := userservice.Service{Store: repo}

	// MARK: Logging
	logger = slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,
		},
	))
	slog.SetDefault(logger)

	err = realMain(service, context.TODO(), uid)
	if err != nil {
		panic(err)
	}
}
