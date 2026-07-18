package service

import (
	"errors"
	"fmt"
	"time"

	"clinic-mgmt/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID      uint     `json:"user_id"`
	Username    string   `json:"username"`
	RealName    string   `json:"real_name"`
	RoleID      uint     `json:"role_id"`
	Permissions []string `json:"permissions"`
	jwt.RegisteredClaims
}

var jwtSecret []byte

func InitJWT(cfg *config.Config) {
	jwtSecret = []byte(cfg.JWTSecret)
}

func GenerateToken(userID uint, username, realName string, roleID uint, permissions []string) (string, error) {
	claims := &Claims{
		UserID:      userID,
		Username:    username,
		RealName:    realName,
		RoleID:      roleID,
		Permissions: permissions,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ParseToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}
	return claims, nil
}
