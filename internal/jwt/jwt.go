package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type Token struct {
	Token     string `json:"token"`
	Provider  string `json:"provider"`
	ExpiresAt int64  `json:"expires_at"`
}

// UserClaims is used for both the access and refresh tokens.
type UserClaims struct {
	UserId   string `json:"id"`
	Type     string `json:"type"` // 'refresh' for refresh tokens and 'access' for access tokens.
	Version  int    `json:"version"`
	Provider string `json:"provider"`
	jwt.StandardClaims
}

// New creates a new JWTHandler instance.
func New(secret string, atExpires, rtExpires time.Duration) *JWTHandler {
	return &JWTHandler{secret: secret, atExpires: atExpires, rtExpires: rtExpires}
}

type JWTHandler struct {
	secret    string
	atExpires time.Duration
	rtExpires time.Duration
}

func (j *JWTHandler) GenerateAccessToken(id string, version int, provider string) (Token, error) {
	expiresAt := time.Now().Add(j.atExpires).Unix()
	claims := UserClaims{
		UserId:   id,
		Version:  version,
		Type:     "access",
		Provider: provider,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return Token{}, err
	}
	return Token{Token: t, ExpiresAt: claims.ExpiresAt, Provider: provider}, nil
}
func (j *JWTHandler) GenerateRefreshToken(id string, version int, provider string) (Token, error) {
	expiresAt := time.Now().Add(j.rtExpires).Unix()
	claims := UserClaims{
		UserId:   id,
		Version:  version,
		Type:     "refresh",
		Provider: provider,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Subject:   id,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return Token{}, err
	}
	return Token{Token: t, ExpiresAt: claims.ExpiresAt, Provider: provider}, nil
}

func (j *JWTHandler) Validate(tokenString string, isRefresh bool) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return &UserClaims{}, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		if isRefresh && claims.Type != "refresh" || !isRefresh && claims.Type != "access" {
			return nil, errors.New("token type mismatch")
		}
		return claims, nil
	}
	return &UserClaims{}, err
}

type Pair struct {
	AccessToken  Token `json:"access_token"`
	RefreshToken Token `json:"refresh_token"`
}

func (j *JWTHandler) CreatePair(userID string, version int, provider string) (*Pair, error) {
	refreshToken, err := j.GenerateRefreshToken(userID, version, provider)
	if err != nil {
		return nil, err
	}
	accessToken, err := j.GenerateAccessToken(userID, version, provider)
	if err != nil {
		return nil, err
	}
	return &Pair{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}

type PasswordResetTokenClaims struct {
	UserId  string `json:"user_id"`
	Version int    `json:"version"`
	jwt.StandardClaims
}

func (j *JWTHandler) GeneratePasswordResetToken(userId string, version int) (string, error) {
	expiresAt := time.Now().Add(5 * time.Minute).Unix()
	claims := PasswordResetTokenClaims{
		UserId:  userId,
		Version: version,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Subject:   userId,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func (j *JWTHandler) ValidatePasswordResetToken(tokenString string) (*PasswordResetTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &PasswordResetTokenClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(*PasswordResetTokenClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, err
}
