package jwt

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt"
)

type Token struct {
	Token      string `json:"token"`
	ExpiresAt  int64  `json:"expires_at"`
	EmailLogin bool   `json:"email_login"` // will be true if the user logged in with email and password.
}

// UserClaims is used for both the access and refresh tokens.
type UserClaims struct {
	UserId     string `json:"id"`
	Type       string `json:"type"` // 'refresh' for refresh tokens and 'access' for access tokens.
	Version    int    `json:"version"`
	EmailLogin bool   `json:"email_login"` // will be true if the user logged in with email and password.
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

func (j *JWTHandler) GenerateAccessToken(id string, version int, emailLogin bool) (Token, error) {
	expiresAt := time.Now().Add(j.atExpires).Unix()
	claims := UserClaims{
		UserId:     id,
		Version:    version,
		Type:       "access",
		EmailLogin: emailLogin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expiresAt,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	t, err := token.SignedString([]byte(j.secret))
	if err != nil {
		return Token{}, err
	}
	return Token{Token: t, ExpiresAt: claims.ExpiresAt, EmailLogin: emailLogin}, nil
}
func (j *JWTHandler) GenerateRefreshToken(id string, version int, emailLogin bool) (Token, error) {
	expiresAt := time.Now().Add(j.rtExpires).Unix()
	claims := UserClaims{
		UserId:     id,
		Version:    version,
		Type:       "refresh",
		EmailLogin: emailLogin,
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
	return Token{Token: t, ExpiresAt: claims.ExpiresAt, EmailLogin: emailLogin}, nil
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

func (j *JWTHandler) CreatePair(userID string, version int, emailLogin bool) (*Pair, error) {
	refreshToken, err := j.GenerateRefreshToken(userID, version, emailLogin)
	if err != nil {
		return nil, err
	}
	accessToken, err := j.GenerateAccessToken(userID, version, emailLogin)
	if err != nil {
		return nil, err
	}
	return &Pair{
		RefreshToken: refreshToken,
		AccessToken:  accessToken,
	}, nil
}
