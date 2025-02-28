.PHONY: proto build run test lint migrate seed

proto:
    protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           api/proto/auth/v1/*.proto api/proto/platform/v1/*.proto

build:
    go build -o bin/IV-auth-service ./cmd/IV-auth-service

run:
    go run ./cmd/IV-auth-service

test:
    go test ./internal/... -v

lint:
    golangci-lint run

migrate:
    migrate -path db/migrations -database "postgres://auth_user:auth_password@localhost:5432/auth_db?sslmode=disable" up

seed:
    psql -h localhost -U auth_user -d auth_db -f db/seeds/users.sql
    psql -h localhost -U auth_user -d auth_db -f db/seeds/platform_accounts.sql