# Go HTTP Server Reference Architecture

A reference architecture for APIs written in Go using the standard library's `net/http` package.

This project aims to demonstrate best practices for building APIs in Go with minimal external dependencies. There are currently three dependencies outside of the standard library:
- google/uuid: Used for generating IDs in tests only
- jackc/pgx: Postgres
- valkey-io/valkey-go: Cache

## Why this project exists

**I wanted to build an API without the use of an API framework**

Most of the blogs and books I read referenced the capabilities of `ServeMux` pre Go 1.22, which included [routing enhancements](https://go.dev/blog/routing-enhancements) that simplify building APIs using only the standard library.

**I wanted to centralize information**

I wanted one place to look for examples. This work was previously scattered across many places on my machine as I worked through individual portions of what ultimately became this project (e.g., middleware, event bus, etc.)

**I wanted to give back to the community**

Open source code has been crucial in my professional development. I wanted to give a comprehensive example of an API in Go so developers of any level of experience could have something to reference in their projects.

## Features

- **Standard Library Based**: Built using only Go's standard library packages (except pgx for db driver and valkey for cache)
- **Background processing of domain events**: Automatically creates an audit log of domain events
- **Admin endpoints**: A comprehensive, separate, API for administrators.
- **Structured Logging**: JSON-formatted logging using the new `log/slog` package
- **Middleware Stack**:
  - Request/Response logging with timing information
  - Panic recovery with proper error responses
  - User context injection
- **Graceful Shutdown**: Handles shutdown signals (SIGTERM/SIGINT) properly
- **Audit Log**: Automatically stores events performed on domain objects
- **Database & Cache**: Postgres and Valkey
- **Configuration**: Sane defaults
- **Health Check Endpoint**: Built-in health check at `/health`

## Getting Started

### Prerequisites

- Go 1.23 or higher
- Docker (for running dependencies locally)
- Make

### Development

The project includes a Makefile with common development tasks. To see all available commands:

```bash
make help
```

### Running the application

There are two APIs - one for external users and one for admins. It is recommended to run both at the same time as the admin API does not expose full functionality.

You can find an example `.env` in `.env.example` (hint, just rename the file to `.env`).

> [!NOTE]
> For the sake of simplicity, the application uses a static user hardcoded in the middleware.
> AuthN/AuthZ are out of scope for this architecture as there are plenty of great tools and articles
> on adding that to Go APIs.
>
>
> Add the known user by running the following command:
> ```bash
> go run cmd/add_user/main.go --user-id f697115f-f723-4c45-8301-e482a21dfd89
> ```

#### Dependencies

This project is dependent on Postgres and ValKey. Both dependencies can be run in Docker compose.
```bash
docker compose up db cache -d
```

### Database

Create the database schema and tables

```bash
psql -U root -h localhost -d example -f sql/create_examples.sql
```

#### User API

Without Docker:
```bash
go run cmd/api/main.go
```

With Docker:
```bash
docker compose up --build api
```

The server will start on port 8080.

#### Admin API

Without Docker:
```bash
go run cmd/admin_api/main.go
```

With Docker:
```bash
docker compose up --build admin
```

The server will start on port 8081.

## Tests

### Unit

Use the following command to run unit tests:
```bash
make test/unit
```

### Integration

Use the following command to run integration tests (requires postgres & valkey):
```bash
make test/integration
```

### Smoke

A simple smoke test for the API can be found in `scripts/api_smoke/main.go`.

> [!NOTE]
> Smoke tests require
> 1. The server and dependencies to be running
> 1. The test user to exist in the db. To create the test user see note in [running the application](#running-the-application).


Run the following command to run smoke tests.
```bash
go run scripts/api_smoke/main.go
```

## Logging

The application uses structured logging with `log/slog`, outputting JSON-formatted logs.

## Error Handling

The server includes middleware for handling panics and converting them to proper HTTP 500 responses. All errors are logged.

## Shutdown Process

The server implements a graceful shutdown process:
1. Captures shutdown signals (SIGTERM/SIGINT)
1. Stops accepting new connections
1. Allows in-flight requests to complete (with a 15-second timeout)
1. Closes all idle connections
1. Stops allowing new messages to event bus
1. Processes any remaining messages in event bus (with a 30-second timeout)
1. Exits cleanly

## Project Structure

The server implementation includes:
- Middleware for logging, error handling, and user context
- Custom response writer to simplify status code tracking
- Server configuration with timeouts
- Graceful shutdown handling
- API handler and service implementations for a public API
- API handler and service implementations for an administrative API
- Event listeners for "user created" and "user deleted"

## Contributing

See [CONTRIBUTING](./CONTRIBUTING.md)
