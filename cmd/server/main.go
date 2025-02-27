package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sukryu/IV-auth-services/internal/config"
	"github.com/sukryu/IV-auth-services/internal/grpcsvc"
	"github.com/sukryu/IV-auth-services/internal/grpcsvc/interceptors"
	"github.com/sukryu/IV-auth-services/internal/infra/database"
	"github.com/sukryu/IV-auth-services/internal/infra/kafka"
	"github.com/sukryu/IV-auth-services/internal/infra/redis"
	authv1 "github.com/sukryu/IV-auth-services/internal/proto/auth/v1"
	"github.com/sukryu/IV-auth-services/internal/services"
	"google.golang.org/grpc"
)

// main은 Authentication Service의 진입점입니다.
// 환경 설정을 로드하고, 데이터베이스, Redis, Kafka 연결을 초기화하며, gRPC 서버를 실행합니다.
func main() {
	// 환경 설정 로드
	env := config.GetEnvFromArgs()
	cfg, err := config.LoadConfig(env)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err) // 초기화 전이므로 std log 유지
	}

	// 데이터베이스 연결
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Redis 연결
	redisClient := redis.NewClient(cfg)
	defer redisClient.Close()

	// Rate Limiter 초기화 (분당 10회 제한 (나중에 플랜별 제한 적용))
	rateLimiter := redis.NewRateLimiter(redisClient, 10, 60*time.Second)

	// Kafka 연결
	kafkaProducer := kafka.NewProducer(cfg)
	defer kafkaProducer.Close()

	// JWT RSA Token 가져오기
	privateKeyBytes, err := os.ReadFile(cfg.JWT.PrivateKeyPath)
	if err != nil {
		log.Fatalf("Failed to read private key from %s: %v", cfg.JWT.PrivateKeyPath, err)
	}
	publicKeyBytes, err := os.ReadFile(cfg.JWT.PublicKeyPath)
	if err != nil {
		log.Fatalf("Failed to read public key from %s: %v", cfg.JWT.PublicKeyPath, err)
	}

	privateKeyBlock, rest := pem.Decode(privateKeyBytes)
	if privateKeyBlock == nil {
		log.Fatalf("Failed to decode private key PEM: invalid format, content: %s, rest: %s", string(privateKeyBytes), string(rest))
	}
	var privateKey *rsa.PrivateKey
	switch privateKeyBlock.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			log.Fatalf("Failed to parse PKCS#8 private key: %v", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			log.Fatalf("Parsed private key is not RSA type")
		}
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			log.Fatalf("Failed to parse PKCS#1 private key: %v", err)
		}
	default:
		log.Fatalf("Unexpected PEM type '%s', expected 'PRIVATE KEY' or 'RSA PRIVATE KEY'", privateKeyBlock.Type)
	}

	publicKeyBlock, rest := pem.Decode(publicKeyBytes)
	if publicKeyBlock == nil {
		log.Fatalf("Failed to decode public key PEM: invalid format, content: %s, rest: %s", string(publicKeyBytes), string(rest))
	}
	if publicKeyBlock.Type != "PUBLIC KEY" {
		log.Fatalf("Expected PEM type 'PUBLIC KEY', got '%s'", publicKeyBlock.Type)
	}
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		log.Fatalf("Failed to parse public key: %v", err)
	}
	publicRSAKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		log.Fatalf("Public key is not RSA type")
	}

	// 도메인 서비스 초기화
	dbUserRepo := database.NewUserRepository(db)
	dbPlatformRepo := database.NewPlatformRepository(db)
	dbTokenRepo := database.NewTokenRepository(db)

	userRepo := redis.NewUserRepository(redisClient, dbUserRepo, 5*time.Minute)
	platformRepo := redis.NewPlatformRepository(redisClient, dbPlatformRepo, 5*time.Minute)
	tokenRepo := redis.NewTokenRepository(redisClient, dbTokenRepo)
	eventPub := kafka.NewEventPublisher(kafkaProducer, cfg.Kafka.TopicPrefix)

	authSvc := services.NewAuthService(services.AuthServiceConfig{
		UserRepo:   userRepo,
		TokenRepo:  tokenRepo,
		EventPub:   eventPub,
		PrivateKey: privateKey,   // 실제 키 로드 필요
		PublicKey:  publicRSAKey, // 실제 키 로드 필요
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 7 * 24 * time.Hour,
	})
	userSvc := services.NewUserService(services.UserServiceConfig{
		UserRepo: userRepo,
		EventPub: eventPub,
	})
	platformSvc := services.NewPlatformService(services.PlatformServiceConfig{
		UserRepo:     userRepo,
		PlatformRepo: platformRepo,
		EventPub:     eventPub,
		OAuthClient:  nil, // 실제 OAuth 클라이언트 구현 필요
	})

	// gRPC 서버 설정
	port := fmt.Sprintf(":%d", cfg.Server.Port)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", port, err)
	}

	// 인터셉터 적용된 gRPC 서버 생성
	grpcServer := grpc.NewServer(interceptors.ChainUnaryInterceptors(authSvc, rateLimiter))
	authv1.RegisterAuthServiceServer(grpcServer, grpcsvc.NewAuthService(authSvc))
	authv1.RegisterUserServiceServer(grpcServer, grpcsvc.NewUserService(userSvc))
	authv1.RegisterPlatformServiceServer(grpcServer, grpcsvc.NewPlatformService(platformSvc))

	// gRPC 서버 실행
	go func() {
		log.Printf("Starting gRPC server on port %s", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Failed to serve gRPC server: %v", err)
		}
	}()

	// 종료 신호 대기
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Printf("Preparing to shutdown gRPC server")
	gracefulStop(grpcServer, 5*time.Second)
	log.Printf("gRPC server stopped successfully")
}

// gracefulStop는 gRPC 서버를 안전하게 종료합니다.
// 지정된 타임아웃 내에 모든 연결을 처리하고 종료합니다.
func gracefulStop(server *grpc.Server, timeout time.Duration) {
	done := make(chan struct{})
	go func() {
		server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		// 정상 종료
	case <-time.After(timeout):
		log.Printf("Graceful shutdown timed out, forcing stop")
		server.Stop()
	}
}
