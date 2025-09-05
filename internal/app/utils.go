package app

import (
	"crypto/rand"
	"encoding/base64"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type GuestClaims struct {
	Slug string `json:"slug"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type GuestTokenResponse struct {
	GuestJWT  string `json:"guest_jwt"`
	ExpiresAt string `json:"expires_at"`
}

func generateSlug(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	encoded := base64.URLEncoding.EncodeToString(bytes)
	if len(encoded) > length {
		encoded = encoded[:length]
	}
	return encoded, nil
}

func generateJWT(slug string) (string, error) {
	claims := jwt.MapClaims{
		"slug": slug,
		"role": "host",
		"iat":  time.Now().Unix(),
		"exp":  time.Now().Add(tokenValidity).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

func sanitizeSlug(slug string) string {
	slug = strings.TrimSpace(slug)

	if len(slug) == 0 || len(slug) > 50 {
		return ""
	}

	matched, _ := regexp.MatchString(`^[a-zA-Z0-9\-_]+$`, slug)
	if !matched {
		return ""
	}

	return slug
}

func generateGuestJWT(slug string) (string, time.Time, error) {
	expiresAt := time.Now().Add(guestTokenValidity)

	claims := GuestClaims{
		Slug: slug,
		Role: "guest",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

func getJWTSecret() string {
	if secret := strings.TrimSpace(os.Getenv("JWT_SECRET")); secret != "" {
		return secret
	}
	return ""
}
