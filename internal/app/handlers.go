package app

import (
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

const (
	tokenValidity      = 24 * time.Hour
	slugLength         = 8
	guestTokenValidity = 2 * time.Hour
)

var jwtSecret = getJWTSecret()

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

// guestTokenHandler обрабатывает GET /rooms/:slug/guest-token
func guestTokenHandler(c echo.Context) error {
	slug := c.Param("slug")

	// Валидируем slug
	slug = sanitizeSlug(slug)
	if slug == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid room slug",
		})
	}

	// Генерируем гостевой токен
	token, expiresAt, err := generateGuestJWT(slug)
	if err != nil {
		log.Printf("Failed to generate guest JWT: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "failed to generate guest token",
		})
	}

	return c.JSON(http.StatusOK, GuestTokenResponse{
		GuestJWT:  token,
		ExpiresAt: expiresAt.UTC().Format(time.RFC3339),
	})
}
