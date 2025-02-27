.PHONY: proto migrate run test lint build

CONFIG_FILE ?= configs/config.yaml

# Protocol Buffers 컴파일
proto:
	protoc -I proto proto/auth/v1/*.proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative
	@echo "Protocol Buffers compiled successfully!"

# 데이터베이스 마이그레이션
migrate:
	go run cmd/migrate/main.go $(CONFIG_FILE)

# 서비스 실행
run:
	go run -ldflags "-X main.configPath=$(CONFIG_FILE)" cmd/server/main.go

# 단위 테스트 실행
test:
	go test ./internal/... -v -cover

# 통합 테스트 실행
integration-test:
	go test ./test/integration -v

# 린트 실행
lint:
	golangci-lint run ./...

# 빌드
build:
	go build -o bin/auth-service cmd/server/main.go