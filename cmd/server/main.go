package main

import (
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

	// Kafka 연결
	kafkaProducer := kafka.NewProducer(cfg)
	defer kafkaProducer.Close()

	// 도메인 서비스 초기화
	dbUserRepo := database.NewUserRepository(db)
	dbPlatformRepo := database.NewPlatformRepository(db)
	dbTokenRepo := database.NewTokenRepository(db)

	userRepo := redis.NewUserRepository(redisClient, dbUserRepo, 5*time.Minute)
	platformRepo := redis.NewPlatformRepository(redisClient, dbPlatformRepo, 5*time.Minute)
	tokenRepo := redis.NewTokenRepository(redisClient, dbTokenRepo)
	eventPub := kafka.NewEventPublisher(kafkaProducer)

	authSvc := services.NewAuthService(services.AuthServiceConfig{
		UserRepo:   userRepo,
		TokenRepo:  tokenRepo,
		EventPub:   eventPub,
		PrivateKey: []byte(cfg.JWT.PrivateKeyPath), // 실제 키 로드 필요
		PublicKey:  []byte(cfg.JWT.PublicKeyPath),  // 실제 키 로드 필요
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
	grpcServer := grpc.NewServer(interceptors.ChainUnaryInterceptors(authSvc))
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
