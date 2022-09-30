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
	UserId string `json:"id"`
	Type   string `json:"type"`
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

func (j *JWTHandler) GenerateAccessToken(id string) (JWTResponse, error) {
	expiresAt := time.Now().Add(j.atExpires).Unix()
	claims := UserClaims{
		UserId: id,
		Type:   "access",
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
func (j *JWTHandler) GenerateRefreshToken(id string) (JWTResponse, error) {
	expiresAt := time.Now().Add(j.rtExpires).Unix()
	claims := UserClaims{
		UserId: id,
		Type:   "refresh",
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

func (j *JWTHandler) Parse(tokenString string, isRefresh bool) (*UserClaims, error) {
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
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ATExpires    int64  `json:"at_expires_at"`
	RTExpires    int64  `json:"rt_expires_at"`
}

func (j *JWTHandler) CreatePair() *Pair {
	return nil
}
