package domain

import (
	"context"
	"errors"
	"time"
)

// PlatformService defines operations for managing platform accounts.
type PlatformService interface {
	LinkAccount(ctx context.Context, userID, platform string, authCode string) (*PlatformAccount, error)
	RevokeAccount(ctx context.Context, userID, platformID string) error
}

// platformService implements PlatformService with domain logic.
type platformService struct {
	platformRepo PlatformAccountRepository
	eventPub     EventPublisher
}

// NewPlatformService creates a new instance of platformService.
func NewPlatformService(platformRepo PlatformAccountRepository, eventPub EventPublisher) PlatformService {
	return &platformService{
		platformRepo: platformRepo,
		eventPub:     eventPub,
	}
}

// LinkAccount links an external platform account to a user.
func (s *platformService) LinkAccount(ctx context.Context, userID, platformStr, authCode string) (*PlatformAccount, error) {
	if userID == "" || platformStr == "" || authCode == "" {
		return nil, errors.New("user id, platform, and auth code must not be empty")
	}

	platform := PlatformType(platformStr)
	if !isValidPlatform(platform) {
		return nil, errors.New("invalid platform type")
	}

	// 임시로 OAuth 교환 결과 가정
	platformID := generateRandomString(36)
	platformUserID := "platform-" + generateRandomString(8)
	platformUsername := "user-" + platformStr
	accessToken := "mock_access_" + generateRandomString(8)
	refreshToken := "mock_refresh_" + generateRandomString(8)
	expiresAt := time.Now().Add(time.Hour)

	account, err := NewPlatformAccount(platformID, userID, platform, platformUserID, platformUsername, accessToken, refreshToken, &expiresAt)
	if err != nil {
		return nil, err
	}

	if err := s.platformRepo.Save(ctx, account); err != nil {
		return nil, errors.New("failed to save platform account: " + err.Error())
	}

	_ = s.eventPub.Publish(&PlatformConnected{userID: userID, platformID: platformID, timestamp: time.Now()})
	return account, nil
}

// RevokeAccount removes a platform account linkage.
func (s *platformService) RevokeAccount(ctx context.Context, userID, platformID string) error {
	if userID == "" || platformID == "" {
		return errors.New("user id and platform id must not be empty")
	}

	account, err := s.platformRepo.FindByID(ctx, platformID)
	if err != nil {
		return errors.New("failed to find platform account: " + err.Error())
	}
	if account == nil || account.UserID() != userID {
		return errors.New("platform account not found or not owned by user")
	}

	if err := s.platformRepo.Delete(ctx, platformID); err != nil {
		return errors.New("failed to delete platform account: " + err.Error())
	}

	_ = s.eventPub.Publish(&PlatformDisconnected{userID: userID, platformID: platformID, timestamp: time.Now()})
	return nil
}
