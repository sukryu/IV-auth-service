package tokens

import (
	"crypto/rsa"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sukryu/IV-auth-services/internal/config"
	"github.com/sukryu/IV-auth-services/internal/core/domain"
	"github.com/sukryu/IV-auth-services/pkg/logger"

	"go.uber.org/zap"
)

// JWTTokenGenerator implements domain.TokenGenerator for JWT with RSA256.
type JWTTokenGenerator struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	logger     *logger.Logger
}

// NewJWTTokenGenerator creates a new JWTTokenGenerator instance using RSA keys.
func NewJWTTokenGenerator(cfg *config.Config, log *logger.Logger) (domain.TokenGenerator, error) {
	// 개인 키 로드
	privateKeyData, err := os.ReadFile(cfg.JWT.PrivateKeyPath)
	if err != nil {
		log.Error("Failed to read private key file", zap.Error(err))
		return nil, errors.New("failed to read private key: " + err.Error())
	}
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(privateKeyData)
	if err != nil {
		log.Error("Failed to parse private key", zap.Error(err))
		return nil, errors.New("failed to parse private key: " + err.Error())
	}

	// 공개 키 로드
	publicKeyData, err := os.ReadFile(cfg.JWT.PublicKeyPath)
	if err != nil {
		log.Error("Failed to read public key file", zap.Error(err))
		return nil, errors.New("failed to read public key: " + err.Error())
	}
	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(publicKeyData)
	if err != nil {
		log.Error("Failed to parse public key", zap.Error(err))
		return nil, errors.New("failed to parse public key: " + err.Error())
	}

	return &JWTTokenGenerator{
		privateKey: privateKey,
		publicKey:  publicKey,
		logger:     log.With(zap.String("component", "jwt_token_generator")),
	}, nil
}

// GenerateAccessToken generates an access token for the given user with expiry.
func (g *JWTTokenGenerator) GenerateAccessToken(userID string, expiry time.Time) (string, error) {
	if userID == "" {
		return "", errors.New("user id must not be empty")
	}

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": expiry.Unix(),
		"iat": time.Now().Unix(),
		"typ": "access",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(g.privateKey)
	if err != nil {
		g.logger.Error("Failed to sign access token", zap.Error(err), zap.String("user_id", userID))
		return "", errors.New("failed to generate access token: " + err.Error())
	}

	g.logger.Debug("Access token generated successfully", zap.String("user_id", userID))
	return tokenString, nil
}

// GenerateRefreshToken generates a refresh token for the given user with expiry.
func (g *JWTTokenGenerator) GenerateRefreshToken(userID string, expiry time.Time) (string, error) {
	if userID == "" {
		return "", errors.New("user id must not be empty")
	}

	claims := jwt.MapClaims{
		"sub": userID,
		"exp": expiry.Unix(),
		"iat": time.Now().Unix(),
		"typ": "refresh",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tokenString, err := token.SignedString(g.privateKey)
	if err != nil {
		g.logger.Error("Failed to sign refresh token", zap.Error(err), zap.String("user_id", userID))
		return "", errors.New("failed to generate refresh token: " + err.Error())
	}

	g.logger.Debug("Refresh token generated successfully", zap.String("user_id", userID))
	return tokenString, nil
}

// ValidateToken verifies the token and returns the user ID if valid.
func (g *JWTTokenGenerator) ValidateToken(tokenStr string) (string, error) {
	if tokenStr == "" {
		return "", errors.New("token must not be empty")
	}

	token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, errors.New("unexpected signing method: " + token.Header["alg"].(string))
		}
		return g.publicKey, nil
	})
	if err != nil {
		g.logger.Error("Failed to parse token", zap.Error(err))
		return "", errors.New("invalid token: " + err.Error())
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		userID, ok := claims["sub"].(string)
		if !ok || userID == "" {
			return "", errors.New("invalid user id in token claims")
		}
		g.logger.Debug("Token validated successfully", zap.String("user_id", userID))
		return userID, nil
	}

	return "", errors.New("token is not valid")
}
