package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type JWTResponse struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

type UserClaims struct {
	UserId        string `json:"id"`
	Type          string `json:"type"`
	Version       int    `json:"version"`
	EmailVerified bool   `json:"email_verified"`
	jwt.StandardClaims
}

func New(secret string, atExpires, rtExpires time.Duration) *JWTHandler {
	return &JWTHandler{secret: secret, atExpires: atExpires, rtExpires: rtExpires}
}

type JWTHandler struct {
	secret    string
	atExpires time.Duration
	rtExpires time.Duration
}

func (j *JWTHandler) GenerateAccessToken(id string, version int, emailVerified bool) (JWTResponse, error) {
	expiresAt := time.Now().Add(j.atExpires).Unix()
	claims := UserClaims{
		UserId:        id,
		Version:       version,
		Type:          "access",
		EmailVerified: emailVerified,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return JWTResponse{}, err
	}
	return JWTResponse{Token: t, ExpiresAt: claims.ExpiresAt}, nil
}
func (j *JWTHandler) GenerateRefreshToken(id string, version int, emailVerified bool) (JWTResponse, error) {
	expiresAt := time.Now().Add(j.rtExpires).Unix()
	claims := UserClaims{
		UserId:        id,
		Version:       version,
		Type:          "refresh",
		EmailVerified: emailVerified,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
			Subject:   id,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return JWTResponse{}, err
	}
	return JWTResponse{Token: t, ExpiresAt: claims.ExpiresAt}, nil
}

func (j *JWTHandler) Validate(tokenString string, isRefresh bool) (*UserClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(j.secret), nil
	})
	if err != nil {
		return &UserClaims{}, err
	}
	if claims, ok := token.Claims.(*UserClaims); ok && token.Valid {
		token.Claims.Valid()
		if isRefresh && claims.Type != "refresh" || !isRefresh && claims.Type != "access" {
			return nil, errors.New("token type mismatch")
		}
		return claims, nil
	}
	return &UserClaims{}, err
}

type Pair struct {
	AccessToken  JWTResponse `json:"access_token"`
	RefreshToken JWTResponse `json:"refresh_token"`
}

func (j *JWTHandler) CreatePair(userID string, version int, emailVerified bool) (*Pair, error) {
	refreshToken, err := j.GenerateRefreshToken(userID, version, emailVerified)
	if err != nil {
		return nil, err
	}
	accessToken, err := j.GenerateAccessToken(userID, version, emailVerified)
	if err != nil {
		return nil, err
	}
	return &Pair{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}
