package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/valkey-io/valkey-go"

	"github.com/moonmoon1919/go-api-reference/internal/adminservice"
	"github.com/moonmoon1919/go-api-reference/internal/auditservice"
	"github.com/moonmoon1919/go-api-reference/internal/bus"
	"github.com/moonmoon1919/go-api-reference/internal/cache"
	"github.com/moonmoon1919/go-api-reference/internal/config"
	"github.com/moonmoon1919/go-api-reference/internal/healthservice"
	"github.com/moonmoon1919/go-api-reference/internal/middleware"
	"github.com/moonmoon1919/go-api-reference/internal/server"
	"github.com/moonmoon1919/go-api-reference/internal/store"
	"github.com/moonmoon1919/go-api-reference/pkg/events"
)

var (
	// Pool middleware resources
	userMiddleware          = middleware.InsertRequestingUser
	errorHandlingMiddleware = middleware.ErrorHandlingMiddleware
	loggingMiddleware       = middleware.LoggingMiddleware

	// Pool validator middleware
	auditLogReadPermissions  = middleware.PermissionValidationMiddleware(middleware.NewHas("admin::auditlog::read"))
	exampleReadPermissions   = middleware.PermissionValidationMiddleware(middleware.NewHas("admin::example::read"))
	exampleDeletePermissions = middleware.PermissionValidationMiddleware(middleware.NewHas("admin::example::delete"))
	userReadPermissions      = middleware.PermissionValidationMiddleware(middleware.NewHas("admin::user::read"))
	userCreatePermissions    = middleware.PermissionValidationMiddleware(middleware.NewHas("admin::user::create"))
	userDeletePermissions    = middleware.PermissionValidationMiddleware(middleware.NewHas("admin::user::delete"))

	// Logging
	logger     *slog.Logger
	logContext = context.Background()
)

// END MOVE
type routerControllers struct {
	admin *adminservice.Controller
}

func buildRoutes(controllers routerControllers, profiling bool) *http.ServeMux {
	router := http.NewServeMux()

	hc := healthservice.HealthController{}
	router.HandleFunc("GET /health", hc.Get)

	// Example routes
	adminRouter := http.NewServeMux()
	adminRouter.Handle("GET /users/{id}/examples", userMiddleware(exampleReadPermissions(controllers.admin.GetExamplesForUser)))
	adminRouter.Handle("GET /examples/{id}", userMiddleware(exampleReadPermissions(controllers.admin.GetExample)))
	adminRouter.Handle("DELETE /examples/{id}", userMiddleware(exampleDeletePermissions(controllers.admin.DeleteExample)))

	// User routes
	adminRouter.Handle("POST /users", userMiddleware(userCreatePermissions(controllers.admin.AddUser)))
	adminRouter.Handle("GET /users", userMiddleware(userReadPermissions(controllers.admin.ListUsers)))
	adminRouter.Handle("GET /users/{id}", userMiddleware(userReadPermissions(controllers.admin.GetUser)))
	adminRouter.Handle("DELETE /users/{id}", userMiddleware(userDeletePermissions(controllers.admin.DeleteUser)))

	// Audit log routes
	adminRouter.Handle("GET /examples/{id}/events", userMiddleware(auditLogReadPermissions(controllers.admin.GetEventsForItem)))
	adminRouter.Handle("GET /users/{id}/events", userMiddleware(auditLogReadPermissions(controllers.admin.GetEventsForUser)))

	router.Handle("/admin/", http.StripPrefix("/admin", adminRouter))

	return router
}

/*
Factory function for creating a new server with all middleware and handlers
*/
func NewServer(config server.Config, controllers routerControllers) *http.Server {
	router := buildRoutes(controllers, len(config.Profiling.Must()) > 0)

	return &http.Server{
		Handler:      errorHandlingMiddleware(loggingMiddleware(router)),
		Addr:         config.Port.Must(),
		ReadTimeout:  config.Timeouts.Read,
		WriteTimeout: config.Timeouts.Write,
		IdleTimeout:  config.Timeouts.Idle,
	}
}

type appConfig struct {
	server   server.Config
	database store.Config
	cache    cache.Config
}

/*
Main entry point for the application
*/
func main() {
	// MARK: Config
	cfg := appConfig{
		server: server.Config{
			Port: config.NewFirst(
				config.NewEnvironmentSource("PORT"),
				config.NewDefaultValueSource(":8081"),
			),
			Profiling: config.NewFirst(
				config.NewEnvironmentSource("PROFILING_ENABLED"),
				config.NewDefaultValueSource(""),
			),
			Timeouts: server.Timeouts{
				Read:  10 * time.Second,
				Write: 10 * time.Second,
				Idle:  30 * time.Second,
			},
		},
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

	// MARK: Repository
	// TODO: Wire up DB cache to delete items
	dbConfig, err := pgxpool.ParseConfig(cfg.database.ConnectionString())
	if err != nil {
		panic(err)
	}

	dbpool, err := pgxpool.NewWithConfig(context.Background(), dbConfig)
	if err != nil {
		panic(err)
	}

	defer dbpool.Close()
	exampleRepo := adminservice.NewExampleSQLRepository(dbpool)
	userRepo := adminservice.NewUserSQLRepository(dbpool)
	auditRepo := adminservice.NewAuditLogSQLRepository(dbpool)

	// MARK: Cache
	cacheClient, err := valkey.NewClient(valkey.ClientOption{
		InitAddress: []string{cfg.cache.Host.Must()},
		SelectDB:    2,
	})

	if err != nil {
		panic(err)
	}

	cache := cache.NewValkeyCache(cacheClient)

	// MARK: Event bus
	auditlogrepo := auditservice.NewSQLRepository(dbpool)
	audotlogsvc := auditservice.Service{Store: auditlogrepo}

	subscriber := func(e events.Event) error {
		d, _ := e.String()
		slog.LogAttrs(logContext, slog.LevelInfo, "EVENT_RECEIVED", slog.String("event", d))

		audotlogsvc.Add(context.TODO(), e)

		return nil
	}

	// TODO: Wire up to service layer for deletes
	eventBus := bus.New(bus.Subscribers{subscriber})

	// MARK: Service
	service := adminservice.Service{
		UserStore:    userRepo,
		ExampleStore: exampleRepo,
		AuditStore:   auditRepo,
		Bus:          &eventBus,
	}

	// MARK: Controllers
	controllers := routerControllers{
		admin: &adminservice.Controller{Service: service, Cache: cache},
	}

	// MARK: Logging
	logger = slog.New(slog.NewJSONHandler(
		os.Stdout,
		&slog.HandlerOptions{
			Level: slog.LevelInfo,
		},
	))
	slog.SetDefault(logger)

	// MARK: Signals
	processShutdownChannel := make(chan os.Signal, server.ProcessChannelsBufferSize)
	serverShutdownChannel := make(chan struct{}, server.ProcessChannelsBufferSize)
	queueShutdownChan := make(chan struct{}, server.ProcessChannelsBufferSize)
	signal.Notify(processShutdownChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	defer close(processShutdownChannel)
	defer close(serverShutdownChannel)
	defer close(queueShutdownChan)

	go eventBus.Listen(queueShutdownChan)

	// MARK: Server
	srvr := NewServer(
		cfg.server,
		controllers,
	)

	// Start the server in a goroutine so we can handle the shutdown signal
	go func() {
		slog.LogAttrs(logContext, slog.LevelInfo, server.StartingServerMsg, slog.String(server.LogKeyAddr, srvr.Addr))

		if err := srvr.ListenAndServe(); errors.Is(err, http.ErrServerClosed) {
			slog.LogAttrs(logContext, slog.LevelInfo, server.ServerClosedMsg)
		}

		serverShutdownChannel <- struct{}{}
	}()

	// Wait for the shutdown signal
	<-processShutdownChannel
	slog.LogAttrs(logContext, slog.LevelInfo, server.ShutdownSignalMsg)

	// Create a context with a timeout for the shutdown
	// Forcing a shutdown after a timeout ensures that any pending requests are not left hanging
	// 15 seconds chosen arbitrarily
	shutdownContext, shutdownCancel := context.WithTimeout(
		context.Background(),
		15*time.Second,
	)
	defer shutdownCancel()

	// Shut down the server with the context
	if err := srvr.Shutdown(shutdownContext); err != nil {
		slog.LogAttrs(logContext, slog.LevelError, server.FailedShutdownMsg, slog.String(server.LogKeyError, err.Error()))
	} else {
		slog.LogAttrs(logContext, slog.LevelInfo, server.ServerShutdownCompleteMsg)
	}

	// Wait for the server to shutdown
	<-serverShutdownChannel

	// Inform the queue we are shutting down
	queueShutdownChan <- struct{}{}
	slog.LogAttrs(logContext, slog.LevelInfo, server.QueueShutdownSignalMsg)

	// Process any events left over in the bus so we don't leave data in an inconsistent state
	// If we don't finish all work within 30 seconds, exit the program
	// 30 chosen arbitrarily
	queueDrainCtx, queueDrainRelease := context.WithTimeout(context.Background(), 30*time.Second)
	defer queueDrainRelease()

	eventBus.CloseAndDrain(queueDrainCtx)
	slog.LogAttrs(logContext, slog.LevelInfo, server.QueueShutdownCompleteMsg)

	slog.LogAttrs(logContext, slog.LevelInfo, server.ProcessShutdownMsg)
}
