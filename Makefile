.PHONY: proto build run test lint migrate seed integration-test deps docker-build docker-run generate-keys

# proto generates Go code from Protocol Buffers definitions
proto:
	protoc --go_out=. --go_opt=paths=source_relative \
           --go-grpc_out=. --go-grpc_opt=paths=source_relative \
           api/proto/auth/v1/*.proto api/proto/platform/v1/*.proto

# build compiles the IV-auth-service binary
build:
	go build -o bin/IV-auth-service ./cmd/IV-auth-service

# run executes the IV-auth-service directly
run:
	go run ./cmd/IV-auth-service

# test runs unit tests for the internal packages
test:
	go test ./internal/... -v -cover

# lint runs golangci-lint to check code quality
lint:
	golangci-lint run

# migrate applies database migrations
migrate:
	migrate -path db/migrations -database "postgres://auth_user:auth_password@localhost:5432/auth_db?sslmode=disable" up

# rollback reverts the last database migration
rollback:
	migrate -path db/migrations -database "postgres://auth_user:auth_password@localhost:5432/auth_db?sslmode=disable" down 1

# seed populates the database with initial data
seed:
	psql -h localhost -U auth_user -d auth_db -f db/seeds/users.sql
	psql -h localhost -U auth_user -d auth_db -f db/seeds/platform_accounts.sql

# integration-test runs integration tests
integration-test:
	go test ./test/integration -v

# deps installs all Go dependencies
deps:
	go mod download
	go mod tidy
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go get github.com/segmentio/kafka-go
	go get github.com/golang-jwt/jwt/v5
	go get github.com/stretchr/testify

# docker-build builds the Docker image for IV-auth-service
docker-build:
	docker build -t iv-auth-service:latest -f deployment/docker/Dockerfile .

# docker-run runs the Docker container with environment variables
docker-run:
	docker run -p 50051:50051 --env-file .env iv-auth-service:latest

# docker-compose-up starts all services using docker-compose
docker-compose-up:
	cd deployment/docker && docker-compose up -d

# docker-compose-down stops all services using docker-compose
docker-compose-down:
	cd deployment/docker && docker-compose down

# generate-keys generates RSA key pair using OpenSSL and stores them in certs/
generate-keys:
	mkdir -p certs
	openssl genrsa -out certs/private.pem 2048
	openssl rsa -in certs/private.pem -outform PEM -pubout -out certs/public.pem