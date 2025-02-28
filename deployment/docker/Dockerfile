# Stage 1: 빌드 단계
FROM golang:1.23 AS builder

# 작업 디렉토리 설정
WORKDIR /app

# Go 모듈 파일 복사 및 의존성 다운로드
COPY go.mod go.sum ./
RUN go mod download

# 소스 코드 전체 복사
COPY . .

# Protocol Buffers 컴파일 (필요 시)
RUN apt-get update && apt-get install -y protobuf-compiler \
    && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest \
    && go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest \
    && make proto

# 애플리케이션 빌드
RUN CGO_ENABLED=0 GOOS=linux go build -o /IV-auth-service ./cmd/IV-auth-service

# Stage 2: 실행 단계
FROM alpine:3.19

# 필요한 런타임 의존성 설치 (예: CA 인증서)
RUN apk --no-cache add ca-certificates

# 작업 디렉토리 설정
WORKDIR /app

# 빌드된 바이너리 복사
COPY --from=builder /IV-auth-service /app/IV-auth-service

# 환경 변수 기본값 설정 (필요 시 외부에서 오버라이드 가능)
ENV PORT=50051

# 컨테이너 실행 시 실행할 명령어
ENTRYPOINT ["/app/IV-auth-service"]

# gRPC 포트 노출
EXPOSE 50051