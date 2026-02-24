package echo

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	echofw "github.com/labstack/echo/v4"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/application/use_cases"
	"github.com/mirola777/Yuno-Idempotency-Challenge/internal/utils/config"
)

type Server struct {
	echo   *echofw.Echo
	config *config.Config
}

func NewServer(cfg *config.Config, container *use_cases.Container) *Server {
	e := echofw.New()
	e.HideBanner = true
	e.HTTPErrorHandler = CustomHTTPErrorHandler

	ConfigureRoutes(e, container)

	return &Server{
		echo:   e,
		config: cfg,
	}
}

func (s *Server) Start() <-chan error {
	errC := make(chan error, 1)

	go func() {
		if err := s.echo.Start(":" + s.config.AppPort); err != nil && err != http.ErrServerClosed {
			errC <- err
		}
	}()

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		<-quit

		log.Println("shutting down server")

		ctx, cancel := context.WithTimeout(context.Background(), s.config.GracefulTimeout)
		defer cancel()

		if err := s.echo.Shutdown(ctx); err != nil {
			errC <- err
		}
		close(errC)
	}()

	log.Printf("server started on port %s", s.config.AppPort)
	return errC
}
