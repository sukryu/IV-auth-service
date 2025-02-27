package test

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"testing"
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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// TestIntegration은 gRPC 서비스와 도메인 서비스의 통합 흐름을 테스트합니다.
// CreateUser → Login 요청 → 도메인 처리 → Kafka 이벤트 발행까지 점검합니다.
func TestIntegration(t *testing.T) {
	log.Println("Starting TestIntegration...")
	cfg, err := config.LoadConfig("dev")
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	log.Println("Config loaded successfully")

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to database at %s:%d: %v", cfg.Database.Host, cfg.Database.Port, err)
	}
	defer db.Close()

	redisClient := redis.NewClient(cfg)
	defer redisClient.Close()

	// Rate Limiter 초기화 (분당 10회 제한 (나중에 플랜별 제한 적용))
	rateLimiter := redis.NewRateLimiter(redisClient, 10, 60*time.Second)

	kafkaProducer := kafka.NewProducer(cfg)
	defer kafkaProducer.Close()

	userRepo := redis.NewUserRepository(redisClient, database.NewUserRepository(db), 5*time.Minute)
	tokenRepo := redis.NewTokenRepository(redisClient, database.NewTokenRepository(db))
	platformRepo := redis.NewPlatformRepository(redisClient, database.NewPlatformRepository(db), 5*time.Minute)
	eventPub := kafka.NewEventPublisher(kafkaProducer, cfg.Kafka.TopicPrefix)

	privateKeyBytes, err := ioutil.ReadFile(cfg.JWT.PrivateKeyPath)
	if err != nil {
		t.Fatalf("Failed to read private key from %s: %v", cfg.JWT.PrivateKeyPath, err)
	}
	publicKeyBytes, err := ioutil.ReadFile(cfg.JWT.PublicKeyPath)
	if err != nil {
		t.Fatalf("Failed to read public key from %s: %v", cfg.JWT.PublicKeyPath, err)
	}

	privateKeyBlock, rest := pem.Decode(privateKeyBytes)
	if privateKeyBlock == nil {
		t.Fatalf("Failed to decode private key PEM: invalid format, content: %s, rest: %s", string(privateKeyBytes), string(rest))
	}
	var privateKey *rsa.PrivateKey
	switch privateKeyBlock.Type {
	case "PRIVATE KEY":
		key, err := x509.ParsePKCS8PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			t.Fatalf("Failed to parse PKCS#8 private key: %v", err)
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			t.Fatalf("Parsed private key is not RSA type")
		}
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(privateKeyBlock.Bytes)
		if err != nil {
			t.Fatalf("Failed to parse PKCS#1 private key: %v", err)
		}
	default:
		t.Fatalf("Unexpected PEM type '%s', expected 'PRIVATE KEY' or 'RSA PRIVATE KEY'", privateKeyBlock.Type)
	}

	publicKeyBlock, rest := pem.Decode(publicKeyBytes)
	if publicKeyBlock == nil {
		t.Fatalf("Failed to decode public key PEM: invalid format, content: %s, rest: %s", string(publicKeyBytes), string(rest))
	}
	if publicKeyBlock.Type != "PUBLIC KEY" {
		t.Fatalf("Expected PEM type 'PUBLIC KEY', got '%s'", publicKeyBlock.Type)
	}
	publicKey, err := x509.ParsePKIXPublicKey(publicKeyBlock.Bytes)
	if err != nil {
		t.Fatalf("Failed to parse public key: %v", err)
	}
	publicRSAKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		t.Fatalf("Public key is not RSA type")
	}

	authSvc := services.NewAuthService(services.AuthServiceConfig{
		UserRepo:   userRepo,
		TokenRepo:  tokenRepo,
		EventPub:   eventPub,
		PrivateKey: privateKey,
		PublicKey:  publicRSAKey,
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
		OAuthClient:  nil,
	})

	grpcServer := grpc.NewServer(interceptors.ChainUnaryInterceptors(authSvc, rateLimiter))
	authv1.RegisterAuthServiceServer(grpcServer, grpcsvc.NewAuthService(authSvc))
	authv1.RegisterUserServiceServer(grpcServer, grpcsvc.NewUserService(userSvc))
	authv1.RegisterPlatformServiceServer(grpcServer, grpcsvc.NewPlatformService(platformSvc))

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Server.Port))
	if err != nil {
		t.Fatalf("Failed to listen on port %d: %v", cfg.Server.Port, err)
	}
	go func() {
		log.Printf("Starting test gRPC server on port %d", cfg.Server.Port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("Server failed: %v", err)
		}
	}()
	defer grpcServer.Stop()

	conn, err := grpc.Dial(
		fmt.Sprintf("localhost:%d", cfg.Server.Port),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("Failed to connect to gRPC server: %v", err)
	}
	defer conn.Close()

	Userclient := authv1.NewUserServiceClient(conn)
	ctx := metadata.NewOutgoingContext(
		context.Background(),
		metadata.Pairs("authorization", "Bearer dummy-token"),
	)
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	// 고유 사용자명으로 CreateUser
	uniqueUsername := fmt.Sprintf("newuser-%d", time.Now().UnixNano())
	res, err := Userclient.CreateUser(ctx, &authv1.CreateUserRequest{
		Username: uniqueUsername,
		Email:    fmt.Sprintf("%s@example.com", uniqueUsername),
		Password: "newpass",
	})
	if err != nil {
		t.Errorf("CreateUser failed: %v", err)
	} else {
		t.Logf("CreateUser succeeded: UserID=%s", res.User.Id)
	}

	client := authv1.NewAuthServiceClient(conn)
	resp, err := client.Login(ctx, &authv1.LoginRequest{
		Username: uniqueUsername,
		Password: "newpass",
	})
	if err != nil {
		t.Errorf("Login failed: %v", err)
	} else {
		t.Logf("Login succeeded: AccessToken=%s, RefreshToken=%s", resp.AccessToken, resp.RefreshToken)
	}

	t.Log("Check Kafka topic 'auth.events.user_created' and 'auth.events.login_succeeded' at 4.206.162.3:29092")
}
