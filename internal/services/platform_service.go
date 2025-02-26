package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/sukryu/IV-auth-services/internal/domain"
	"github.com/sukryu/IV-auth-services/internal/domain/repo"
)

// PlatformService는 외부 플랫폼 계정 연동과 관련된 핵심 비즈니스 로직을 처리하는 도메인 서비스입니다.
// 플랫폼 계정 연결, 토큰 갱신, 계정 해제 기능을 제공하며, 외부 의존성은 인터페이스를 통해 주입받습니다.
type PlatformService struct {
	userRepo     repo.UserRepository     // 사용자 데이터 접근 인터페이스
	platformRepo repo.PlatformRepository // 플랫폼 계정 데이터 접근 인터페이스
	eventPub     repo.EventPublisher     // Kafka 이벤트 발행 인터페이스
	oauthClient  OAuthClient             // OAuth API 호출 인터페이스
}

// PlatformServiceConfig는 PlatformService 초기화를 위한 설정을 정의합니다.
type PlatformServiceConfig struct {
	UserRepo     repo.UserRepository
	PlatformRepo repo.PlatformRepository
	EventPub     repo.EventPublisher
	OAuthClient  OAuthClient
}

// NewPlatformService는 새로운 PlatformService 인스턴스를 생성하여 반환합니다.
// 모든 의존성은 PlatformServiceConfig를 통해 주입되며, 필수 필드 검증 후 초기화합니다.
func NewPlatformService(cfg PlatformServiceConfig) *PlatformService {
	if cfg.UserRepo == nil || cfg.PlatformRepo == nil || cfg.EventPub == nil || cfg.OAuthClient == nil {
		log.Fatal("NewPlatformService: 모든 저장소, 이벤트 발행기, OAuth 클라이언트는 필수입니다")
	}

	return &PlatformService{
		userRepo:     cfg.UserRepo,
		platformRepo: cfg.PlatformRepo,
		eventPub:     cfg.EventPub,
		oauthClient:  cfg.OAuthClient,
	}
}

// OAuthClient는 외부 플랫폼 OAuth API와의 상호작용을 추상화하는 인터페이스입니다.
// 토큰 교환 및 갱신 요청을 처리합니다.
type OAuthClient interface {
	// ExchangeCode는 OAuth 인증 코드를 액세스/리프레시 토큰으로 교환합니다.
	// 플랫폼별 엔드포인트로 요청을 보내고 결과를 반환합니다.
	ExchangeCode(ctx context.Context, platform domain.PlatformType, code string) (accessToken, refreshToken string, expiresAt time.Time, err error)

	// RefreshAccessToken은 리프레시 토큰을 사용하여 새로운 액세스 토큰을 발급받습니다.
	// 만료된 토큰을 갱신하며, 실패 시 에러를 반환합니다.
	RefreshAccessToken(ctx context.Context, platform domain.PlatformType, refreshToken string) (accessToken, newRefreshToken string, expiresAt time.Time, err error)
}

// ConnectPlatform은 사용자의 외부 플랫폼 계정을 연동합니다.
// OAuth 인증 코드를 사용하여 토큰을 교환하고, 플랫폼 계정을 생성한 후 이벤트를 발행합니다.
// 목표: 실시간 처리 (100ms 이내), 보안성 유지 (OAuth 2.0).
func (s *PlatformService) ConnectPlatform(ctx context.Context, userID string, platform domain.PlatformType, authCode string) (*domain.PlatformAccount, error) {
	// 사용자 존재 여부 확인
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user for platform connection: %w", err)
	}
	if user == nil {
		return nil, domain.ErrUserNotFound
	}
	if !user.IsActive() {
		return nil, domain.ErrUserNotActive
	}

	// OAuth 토큰 교환
	accessToken, refreshToken, expiresAt, err := s.oauthClient.ExchangeCode(ctx, platform, authCode)
	if err != nil {
		s.eventPub.Publish(ctx, domain.PlatformConnectionFailed{
			UserID:    userID,
			Platform:  platform,
			Reason:    "oauth code exchange failed",
			Timestamp: time.Now().UTC(),
		})
		return nil, fmt.Errorf("failed to exchange oauth code: %w", err)
	}

	// 플랫폼 사용자 정보 조회 (가정된 API 호출)
	platformUserID, platformUsername, err := s.fetchPlatformUserInfo(ctx, platform, accessToken)
	if err != nil {
		s.eventPub.Publish(ctx, domain.PlatformConnectionFailed{
			UserID:    userID,
			Platform:  platform,
			Reason:    "platform user info fetch failed",
			Timestamp: time.Now().UTC(),
		})
		return nil, fmt.Errorf("failed to fetch platform user info: %w", err)
	}

	// 플랫폼 계정 엔티티 생성
	account, err := domain.NewPlatformAccount(
		uuid.New().String(),
		userID,
		platform,
		platformUserID,
		platformUsername,
		accessToken,
		refreshToken,
		expiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create platform account entity: %w", err)
	}

	// 데이터베이스에 저장
	if err := s.platformRepo.Insert(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to insert platform account: %w", err)
	}

	// PlatformConnected 이벤트 발행
	s.eventPub.Publish(ctx, domain.PlatformConnected{
		UserID:            userID,
		PlatformAccountID: account.ID,
		Platform:          platform,
		Timestamp:         time.Now().UTC(),
	})
	return account, nil
}

// RefreshPlatformToken은 플랫폼 계정의 만료된 토큰을 갱신합니다.
// 리프레시 토큰을 사용하여 새로운 토큰을 발급받고, 계정 정보를 업데이트합니다.
func (s *PlatformService) RefreshPlatformToken(ctx context.Context, platformAccountID string) (*domain.PlatformAccount, error) {
	// 플랫폼 계정 조회
	account, err := s.platformRepo.FindByID(ctx, platformAccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to find platform account for refresh: %w", err)
	}
	if account == nil {
		return nil, domain.ErrPlatformAccountNotFound
	}

	// 사용자 상태 확인
	user, err := s.userRepo.FindByID(ctx, account.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to find user for platform refresh: %w", err)
	}
	if user == nil || !user.IsActive() {
		return nil, domain.ErrUserNotActive
	}

	// 토큰 갱신
	accessToken, newRefreshToken, expiresAt, err := s.oauthClient.RefreshAccessToken(ctx, account.Platform, account.RefreshToken)
	if err != nil {
		s.eventPub.Publish(ctx, domain.PlatformTokenRefreshFailed{
			UserID:            account.UserID,
			PlatformAccountID: account.ID,
			Reason:            "token refresh failed",
			Timestamp:         time.Now().UTC(),
		})
		return nil, fmt.Errorf("failed to refresh platform token: %w", err)
	}

	// 계정 정보 업데이트
	if err := account.UpdateTokens(accessToken, newRefreshToken, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to update platform tokens: %w", err)
	}

	// 데이터베이스에 반영
	if err := s.platformRepo.Update(ctx, account); err != nil {
		return nil, fmt.Errorf("failed to update platform account: %w", err)
	}

	// PlatformTokenRefreshed 이벤트 발행
	s.eventPub.Publish(ctx, domain.PlatformTokenRefreshed{
		UserID:            account.UserID,
		PlatformAccountID: account.ID,
		Platform:          account.Platform,
		Timestamp:         time.Now().UTC(),
	})
	return account, nil
}

// DisconnectPlatform은 플랫폼 계정을 사용자로부터 해제합니다.
// 계정 삭제 후 이벤트를 발행하며, 논리적 삭제(상태 변경) 대신 물리적 삭제로 구현합니다.
func (s *PlatformService) DisconnectPlatform(ctx context.Context, platformAccountID string) error {
	// 플랫폼 계정 조회
	account, err := s.platformRepo.FindByID(ctx, platformAccountID)
	if err != nil {
		return fmt.Errorf("failed to find platform account for disconnection: %w", err)
	}
	if account == nil {
		return domain.ErrPlatformAccountNotFound
	}

	// 사용자 상태 확인
	user, err := s.userRepo.FindByID(ctx, account.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user for platform disconnection: %w", err)
	}
	if user == nil || !user.IsActive() {
		return domain.ErrUserNotActive
	}

	// 데이터베이스에서 삭제
	if err := s.platformRepo.Delete(ctx, platformAccountID); err != nil {
		return fmt.Errorf("failed to delete platform account: %w", err)
	}

	// PlatformDisconnected 이벤트 발행
	s.eventPub.Publish(ctx, domain.PlatformDisconnected{
		UserID:            account.UserID,
		PlatformAccountID: account.ID,
		Platform:          account.Platform,
		Timestamp:         time.Now().UTC(),
	})
	return nil
}

// fetchPlatformUserInfo는 OAuth 액세스 토큰을 사용하여 플랫폼 사용자 정보를 조회합니다.
// 실제 구현에서는 플랫폼별 API 호출 로직이 필요하며, 여기서는 가정된 결과만 반환합니다.
func (s *PlatformService) fetchPlatformUserInfo(ctx context.Context, platform domain.PlatformType, accessToken string) (userID, username string, err error) {
	// TODO: 플랫폼별 API 호출 구현 필요 (예: Twitch Helix API, YouTube Data API)
	// 예시로 더미 데이터 반환
	switch platform {
	case domain.PlatformTypeTwitch:
		return "twitch-123", "TwitchUser", nil
	case domain.PlatformTypeYouTube:
		return "youtube-456", "YouTubeUser", nil
	default:
		return "", "", fmt.Errorf("unsupported platform: %v", platform)
	}
}
