package app

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

const (
	jwtSecret     = "your-secret-key-change-in-production" // Заменить на ENV переменную
	tokenValidity = 24 * time.Hour
	slugLength    = 8
)

type RoomResponse struct {
	Slug string `json:"slug"`
	JWT  string `json:"jwt"`
}

type HealthResponse struct {
	Status string `json:"status"`
	Time   string `json:"time"`
}

func healthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Status: "ok",
		Time:   time.Now().UTC().Format(time.RFC3339),
	})
}

func roomsAnonymousHandler(c echo.Context) error {
	slug, err := generateSlug(slugLength)
	if err != nil {
		log.Printf("Failed to generate slug: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to generate room identifier",
		})
	}

	token, err := generateJWT(slug)
	if err != nil {
		log.Printf("Failed to generate JWT: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to generate access token",
		})
	}

	return c.JSON(http.StatusCreated, RoomResponse{
		Slug: slug,
		JWT:  token,
	})
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
