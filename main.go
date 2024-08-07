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
	"ytsruh.com/backmeup/views"
)

func main() {
	// Create a directory to store zip files
	err := os.MkdirAll("zips", 0755) // 0755 is the file permission (read and write permission)
	if err != nil {
		log.Fatalf("error creating temp directory: %s", err)
	}
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return Render(c, http.StatusOK, views.Home())
	})
	e.POST("/", Scan)
	e.GET("/dl/:zip", func(c echo.Context) error {
		return c.File("zips/" + c.Param("zip"))
	})
	e.GET("/time", func(c echo.Context) error {
		return Render(c, http.StatusOK, views.TimeComponent(time.Now()))
	})
	e.GET("/404", func(c echo.Context) error {
		return Render(c, http.StatusNotFound, views.NotFoundComponent())
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Start server
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
