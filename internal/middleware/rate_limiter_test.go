package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"golang.org/x/time/rate"
)

func TestIPRateLimiter(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(2), 1)

	e := echo.New()
	e.Use(limiter.Middleware())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusTooManyRequests, rec2.Code)

	time.Sleep(500 * time.Millisecond)
	req3 := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec3 := httptest.NewRecorder()
	e.ServeHTTP(rec3, req3)
	assert.Equal(t, http.StatusOK, rec3.Code)
}

func TestIPRateLimiterDifferentIPs(t *testing.T) {
	limiter := NewIPRateLimiter(rate.Limit(1), 1)

	e := echo.New()
	e.Use(limiter.Middleware())
	e.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "success")
	})

	req1 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req1.Header.Set("X-Real-IP", "192.168.1.1")
	rec1 := httptest.NewRecorder()
	e.ServeHTTP(rec1, req1)
	assert.Equal(t, http.StatusOK, rec1.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/test", nil)
	req2.Header.Set("X-Real-IP", "192.168.1.2")
	rec2 := httptest.NewRecorder()
	e.ServeHTTP(rec2, req2)
	assert.Equal(t, http.StatusOK, rec2.Code)
}
