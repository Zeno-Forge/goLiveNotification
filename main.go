package main

import (
	"context"
	"log/slog"

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
	e.Static("/static", "./static")
	e.Static("/data/uploads", "./data/uploads")

	// Routes
	e.GET("/", handlers.MainPageHandler)
	e.GET("/post/:id", handlers.GetPostHandler)
	e.GET("/post/create", handlers.CreatePostHandler)
	e.PUT("/post/:id", handlers.EditPostHandler)
	e.DELETE("/post/:id", handlers.DeletePostHandler)
	e.GET("/events", handlers.EventsHandler)
	e.GET("/get-posts", handlers.GetPostList)
	e.GET("/template", handlers.GetTemplateHandler)
	e.PUT("/template", handlers.EditTemplateHandler)
	e.POST("/webhook", handlers.WebhookHandler)

	return e
}

func main() {
	// Get port from ENV
	port, ok := utils.GetEnv("GOLIVE_PORT")
	if !ok {
		port = ":8080"
	}

	server := NewServer(log.Log)

	handlers.ManageScheduledPosts()

	// Start server
	server.Logger.Fatal(server.Start(port))
}
