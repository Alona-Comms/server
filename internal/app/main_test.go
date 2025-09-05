package app

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func setupTestServer() *echo.Echo {
	e := echo.New()
	e.GET("/health", healthHandler)
	e.POST("/rooms/anonymous", roomsAnonymousHandler)
	return e
}

func TestHealthEndpoint(t *testing.T) {
	e := setupTestServer()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRoomsAnonymousEndpoint(t *testing.T) {
	e := setupTestServer()
	req := httptest.NewRequest(http.MethodPost, "/rooms/anonymous", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	assert.Equal(t, "application/json", rec.Header().Get(echo.HeaderContentType))
	assert.Equal(t, http.StatusCreated, rec.Code)
}
