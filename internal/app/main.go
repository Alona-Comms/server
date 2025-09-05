package app

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var defaultPort = "8080"

type App struct {
	e    *echo.Echo
	port string
}

func Initialize() *App {
	app := &App{
		e:    echo.New(),
		port: getPort(),
	}

	app.e.HideBanner = true
	app.e.HidePort = false

	app.e.Use(middleware.Logger())
	app.e.Use(middleware.Recover())
	app.e.Use(middleware.CORS())
	app.e.Use(middleware.RequestID())

	app.e.GET("/health", healthHandler)
	app.e.POST("/rooms/anonymous", roomsAnonymousHandler)

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
	return a.e.Shutdown(ctx)
}

func getPort() string {
	if port := os.Getenv("PORT"); port != "" {
		return port
	}
	return defaultPort
}
