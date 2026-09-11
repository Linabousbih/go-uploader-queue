# Async reports

A Go API and background worker for generating reports with PostgreSQL, SQS, and S3.

## Project structure

```text
cmd/                 Executable entry points: apiserver, worker, awstest
apiserver/           HTTP server startup and shutdown
config/              Environment configuration
routes/              HTTP endpoint registration and middleware wiring
controllers/         Authentication, health, and report request handlers
middleware/          Request logging and authentication
helpers/             JSON encoding, HTTP errors, JWTs, and user context
models/              Domain records, API request/response types, queue messages
repositories/        SQL queries and persistence
database/            PostgreSQL connection setup
services/reports/    Report generation and queue worker
fixtures/            Integration-test setup
migrations/          Database schema migrations
terraform/           Infrastructure configuration
```

The API entry point creates dependencies and passes controllers to `routes.New`.
The resulting HTTP handler is passed to `apiserver.New`, which manages the server
lifecycle. Controllers use repositories for persistence and helpers for shared
HTTP and authentication functions. Repositories and report services share domain
types from `models`.

Add HTTP endpoints in `routes/routes.go`, their handlers in `controllers`, and
their request/response types in `models`. Keep SQL in `repositories` and report
processing in `services/reports`.

## Development

Build all executables with `go build ./...` and run checks with `go test ./...`.
The repository integration test needs PostgreSQL and the database environment
variables defined in `config/config.go`.

Run the API with `go run ./cmd/apiserver` or the worker with
`go run ./cmd/worker` after configuring the database and AWS/LocalStack resources.
