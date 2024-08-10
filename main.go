package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"ytsruh.com/backmeup/utils"
	"ytsruh.com/backmeup/views"
)

func main() {
	// Create a directory to store zip files
	err := os.MkdirAll("zips", 0755) // 0755 is the file permission (read and write permission)
	if err != nil {
		log.Fatalf("error creating temp directory: %s", err)
	}
	// Start server and register middlewares
	e := echo.New()
	e.Use(middleware.Recover())
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method} status=${status} uri=${uri}\n",
	}))
	e.Use(middleware.StaticWithConfig(middleware.StaticConfig{
		Root: "public",
	}))
	e.Use(middleware.Secure())

	// Register routes
	e.GET("/", func(c echo.Context) error {
		return Render(c, http.StatusOK, views.Home())
	})
	e.POST("/", Scan)
	e.GET("/dl/:zip", func(c echo.Context) error {
		return c.File("zips/" + c.Param("zip"))
	})
	e.GET("/bulk", func(c echo.Context) error {
		return Render(c, http.StatusOK, views.Bulk())
	})
	e.POST("/bulk", Bulk)

	// Start cron jobs
	utils.CleanUpZips()
	utils.StartCronJobs()

	// Start server
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	go func() {
		e.Logger.Info("Starting server on port 1323")
		if err := e.Start(":1323"); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server with a timeout of 10 seconds.
	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

// This custom Render replaces Echo's echo.Context.Render() with templ's templ.Component.Render().
func Render(ctx echo.Context, statusCode int, t templ.Component) error {
	buf := templ.GetBuffer()
	defer templ.ReleaseBuffer(buf)

	if err := t.Render(ctx.Request().Context(), buf); err != nil {
		return err
	}

	return ctx.HTML(statusCode, buf.String())
}
