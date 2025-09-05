package app

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/Kaamos-Comms/server/internal/signaling"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var defaultPort = "8080"

type App struct {
	e               *echo.Echo
	signalingServer *signaling.Server
	port            string
}

func Initialize() *App {
	app := &App{
		e:               echo.New(),
		signalingServer: signaling.NewServer(),
		port:            getPort(),
	}

	app.e.HideBanner = true
	app.e.HidePort = false

	app.e.Use(middleware.Logger())
	app.e.Use(middleware.Recover())
	app.e.Use(middleware.CORS())
	app.e.Use(middleware.RequestID())

	app.e.GET("/health", healthHandler)
	app.e.POST("/rooms/anonymous", roomsAnonymousHandler)

	app.e.GET("/ws", echo.WrapHandler(http.HandlerFunc(app.signalingServer.HandleWebSocket)))
	app.e.GET("/rooms/:slug/stats", func(c echo.Context) error {
		slug := c.Param("slug")
		stats := app.signalingServer.GetRoomStats(slug)
		if stats == nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "room not found"})
		}
		return c.JSON(http.StatusOK, stats)
	})

	app.e.GET("/rooms/:slug/guest-token", guestTokenHandler)

	return app
}

func (a *App) Start() {
	go func() {
		log.Printf("Starting server on port %s", a.port)
		if err := a.e.Start(":" + a.port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server startup failed: %v", err)
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) error {
	a.signalingServer.Shutdown()
	return a.e.Shutdown(ctx)
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return defaultPort
}
