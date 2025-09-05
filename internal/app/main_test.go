package app

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

// func TestGuestTokenHandler(t *testing.T) {
// 	e := echo.New()
// 	e.GET("/rooms/:slug/guest-token", guestTokenHandler)

// 	tests := []struct {
// 		name           string
// 		slug           string
// 		expectedStatus int
// 		checkToken     bool
// 	}{
// 		{
// 			name:           "valid slug",
// 			slug:           "test-room-123",
// 			expectedStatus: http.StatusOK,
// 			checkToken:     true,
// 		},
// 		{
// 			name:           "invalid slug with spaces",
// 			slug:           "test room",
// 			expectedStatus: http.StatusBadRequest,
// 			checkToken:     false,
// 		},
// 		{
// 			name:           "empty slug",
// 			slug:           "",
// 			expectedStatus: http.StatusBadRequest,
// 			checkToken:     false,
// 		},
// 		{
// 			name:           "invalid characters",
// 			slug:           "test@room!",
// 			expectedStatus: http.StatusBadRequest,
// 			checkToken:     false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			req := httptest.NewRequest(http.MethodGet, "/rooms/"+tt.slug+"/guest-token", nil)
// 			rec := httptest.NewRecorder()

// 			e.ServeHTTP(rec, req)

// 			assert.Equal(t, tt.expectedStatus, rec.Code)

// 			if tt.checkToken {
// 				var response GuestTokenResponse
// 				err := json.Unmarshal(rec.Body.Bytes(), &response)
// 				assert.NoError(t, err)

// 				// Проверяем, что токен не пустой
// 				assert.NotEmpty(t, response.GuestJWT)
// 				assert.NotEmpty(t, response.ExpiresAt)

// 				// Проверяем, что можем распарсить токен
// 				token, err := jwt.ParseWithClaims(response.GuestJWT, &GuestClaims{}, func(token *jwt.Token) (interface{}, error) {
// 					return []byte(jwtSecret), nil
// 				})
// 				assert.NoError(t, err)

// 				claims, ok := token.Claims.(*GuestClaims)
// 				assert.True(t, ok)
// 				assert.Equal(t, tt.slug, claims.Slug)
// 				assert.Equal(t, "guest", claims.Role)

// 				// Проверяем время истечения
// 				expiresAt, err := time.Parse(time.RFC3339, response.ExpiresAt)
// 				assert.NoError(t, err)
// 				assert.True(t, expiresAt.After(time.Now()))
// 			}
// 		})
// 	}
// }

func TestGuestTokenHandlerWithContext(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/rooms/:slug/guest-token")
	c.SetParamNames("slug")
	c.SetParamValues("valid-room-123")

	err := guestTokenHandler(c)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.SetPath("/rooms/:slug/guest-token")
	c2.SetParamNames("slug")
	c2.SetParamValues("invalid slug with spaces")

	err2 := guestTokenHandler(c2)
	assert.NoError(t, err2) // Handler не возвращает ошибку, но устанавливает статус
	assert.Equal(t, http.StatusBadRequest, rec2.Code)
}

func TestSanitizeSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"valid-slug", "valid-slug"},
		{"valid_slug_123", "valid_slug_123"},
		{"  spaced  ", "spaced"},
		{"invalid@chars", ""},
		{"", ""},
		{"a", "a"},
		{strings.Repeat("a", 51), ""}, // Слишком длинный
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := sanitizeSlug(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
