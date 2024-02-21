package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"zenoforge.com/goLiveNotif/handlers"
	"zenoforge.com/goLiveNotif/log"
	"zenoforge.com/goLiveNotif/utils"
)

func NewServer(logger *slog.Logger) *echo.Echo {
	e := echo.New()

	// Middleware
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		LogMethod:   true,
		HandleError: true, // forwards error to the global error handler, so it can decide appropriate status code
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.Group(
						v.Method,
						slog.String("uri", v.URI),
						slog.Int("status", v.Status),
					),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.Group(
						v.Method,
						slog.String("uri", v.URI),
						slog.Int("status", v.Status),
						slog.String("err", v.Error.Error()),
					),
				)
			}
			return nil
		},
	}))
	e.Use(middleware.Recover())

	// Static files
	e.Static("/data/uploads", "./data/uploads")

	// Routes
	e.GET("/", handlers.MainPage)
	e.GET("/post/:id", handlers.GetPost)
	e.GET("/post/create", handlers.CreatePost)
	e.PUT("/post/:id", handlers.EditPost)
	e.DELETE("/post/:id", handlers.DeletePost)

	return e
}

func main() {
	// Load posts from file
	if err := utils.LoadDataFromFile(&handlers.Posts, "data", "postStorage.json"); err != nil && !os.IsNotExist(err) {
		log.LogIt(context.Background(), slog.LevelError, "Failed to load posts", slog.String("Error:", err.Error()))
	}

	server := NewServer(log.Log)

	// Start server
	server.Logger.Fatal(server.Start(":8989"))
}
