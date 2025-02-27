# Makefile for ImmersiVerse Authentication Service
# 주요 명령어: 마이그레이션, 키 생성, 서버 실행, 테스트, 빌드

# 데이터베이스 연결 문자열을 환경 변수로 정의
DB_USER ?= auth_user
DB_PASSWORD ?= test_password
DB_HOST ?= 4.206.162.3
DB_PORT ?= 5432
DB_NAME ?= auth_db_test
DB_DSN ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=disable

# 설정 파일 경로
CONFIG_PATH ?= ./configs

# JWT 키 경로
JWT_PRIVATE_KEY_PATH ?= $(CONFIG_PATH)/keys/dev/jwt-private.pem
JWT_PUBLIC_KEY_PATH ?= $(CONFIG_PATH)/keys/dev/jwt-public.pem

# 기본 변수
GO ?= go
MIGRATE ?= migrate

# 마이그레이션 명령
.PHONY: migrate migrate-create migrate-up migrate-down migrate-force migrate-version migrate-reset rollback

# 마이그레이션 실행 (전체)
migrate:
	$(MIGRATE) -path db/migrations -database "$(DB_DSN)" up

# 새 마이그레이션 파일 생성
migrate-create:
	@read -p "마이그레이션 이름 입력: " name; \
	$(MIGRATE) create -ext sql -dir db/migrations -seq $${name}

# 지정된 단계만큼 업그레이드
migrate-up:
	@read -p "적용할 마이그레이션 단계 수 (기본값: all): " steps; \
	if [ -z "$$steps" ]; then \
		$(MIGRATE) -path db/migrations -database "$(DB_DSN)" up; \
	else \
		$(MIGRATE) -path db/migrations -database "$(DB_DSN)" up $$steps; \
	fi

# 지정된 단계만큼 다운그레이드
migrate-down:
	@read -p "롤백할 마이그레이션 단계 수 (기본값: 1): " steps; \
	if [ -z "$$steps" ]; then \
		$(MIGRATE) -path db/migrations -database "$(DB_DSN)" down 1; \
	else \
		$(MIGRATE) -path db/migrations -database "$(DB_DSN)" down $$steps; \
	fi

# 특정 버전으로 강제 설정
migrate-force:
	@read -p "강제 설정할 버전 번호: " version; \
	$(MIGRATE) -path db/migrations -database "$(DB_DSN)" force $$version

# 현재 마이그레이션 버전 확인
migrate-version:
	$(MIGRATE) -path db/migrations -database "$(DB_DSN)" version

# 롤백 (migrate-down의 별칭)
rollback: migrate-down

# 마이그레이션 재설정 (모두 롤백 후 다시 적용)
migrate-reset:
	@echo "\033[31m주의: 모든 마이그레이션을 롤백하고 재적용합니다.\033[0m"
	@read -p "계속하려면 'yes'를 입력하세요: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		$(MIGRATE) -path db/migrations -database "$(DB_DSN)" down -all && \
		$(MIGRATE) -path db/migrations -database "$(DB_DSN)" up; \
	else \
		echo "취소되었습니다."; \
	fi

# JWT 키 쌍 생성
.PHONY: generate-key
generate-key:
	@mkdir -p $(CONFIG_PATH)/keys/dev
	@openssl genrsa -out $(JWT_PRIVATE_KEY_PATH) 2048
	@openssl rsa -in $(JWT_PRIVATE_KEY_PATH) -pubout -out $(JWT_PUBLIC_KEY_PATH)
	@echo "JWT 키 쌍 생성 완료: $(JWT_PRIVATE_KEY_PATH), $(JWT_PUBLIC_KEY_PATH)"

# 서버 빌드
.PHONY: build
build:
	$(GO) build -o bin/auth-service ./cmd/main.go

# 서버 실행
.PHONY: run-server
run-server:
	@CONFIG_PATH=$(CONFIG_PATH) $(GO) run ./cmd/main.go dev

# 통합 테스트 실행
.PHONY: test
test:
	@CONFIG_PATH=$(CONFIG_PATH) $(GO) test -v ./internal/grpcsvc/test/

# 의존성 설치 (migrate 도구)
.PHONY: install-deps
install-deps:
	@go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "migrate 도구 설치 완료"

# 전체 정리 및 재설정
.PHONY: clean
clean:
	@rm -rf bin/
	@echo "빌드 파일 정리 완료"