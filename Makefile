# 데이터베이스 연결 문자열을 환경 변수로 정의
DB_USER ?= auth_user
DB_PASSWORD ?= dev_password
DB_HOST ?= localhost
DB_PORT ?= 5432
DB_NAME ?= auth_db_dev
DB_DSN = postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)

# 마이그레이션 명령
.PHONY: migrate migrate-create migrate-up migrate-down migrate-force migrate-version

# 마이그레이션 실행 (전체)
migrate:
	migrate -path db/migrations -database "$(DB_DSN)?sslmode=disable" up

# 새 마이그레이션 파일 생성
migrate-create:
	@read -p "마이그레이션 이름 입력: " name; \
	migrate create -ext sql -dir db/migrations -seq $${name}

# 지정된 단계만큼 업그레이드
migrate-up:
	@read -p "적용할 마이그레이션 단계 수 (기본값: all): " steps; \
	if [ -z "$$steps" ]; then \
		migrate -path db/migrations -database "$(DB_DSN)?sslmode=disable" up; \
	else \
		migrate -path db/migrations -database "$(DB_DSN)?sslmode=disable" up $$steps; \
	fi

# 지정된 단계만큼 다운그레이드
migrate-down:
	@read -p "롤백할 마이그레이션 단계 수 (기본값: 1): " steps; \
	if [ -z "$$steps" ]; then \
		migrate -path db/migrations -database "$(DB_DSN)?sslmode=disable" down 1; \
	else \
		migrate -path db/migrations -database "$(DB_DSN)?sslmode=disable" down $$steps; \
	fi

# 특정 버전으로 강제 설정
migrate-force:
	@read -p "강제 설정할 버전 번호: " version; \
	migrate -path db/migrations -database "$(DB_DSN)?sslmode=disable" force $$version

# 현재 마이그레이션 버전 확인
migrate-version:
	migrate -path db/migrations -database "$(DB_DSN)?sslmode=disable" version

# 롤백 (migrate-down의 별칭)
rollback: migrate-down

# 마이그레이션 재설정 (모두 롤백 후 다시 적용)
migrate-reset:
	@echo "주의: 모든 마이그레이션을 롤백하고 재적용합니다."
	@read -p "계속하려면 'yes'를 입력하세요: " confirm; \
	if [ "$$confirm" = "yes" ]; then \
		migrate -path db/migrations -database "$(DB_DSN)?sslmode=disable" down -all && \
		migrate -path db/migrations -database "$(DB_DSN)?sslmode=disable" up; \
	else \
		echo "취소되었습니다."; \
	fi