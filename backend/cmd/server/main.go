package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/isabelagutierrez/calculator/backend/internal/api"
)

func main() {
	ctx, cancel := signalContext()
	defer cancel()
	addr, config := serverConfig()
	if err := run(ctx, addr, api.NewHandler(config)); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, addr string, handler http.Handler) error {
	server := newServer(addr, handler)
	serverErrors := make(chan error, 1)
	go func() {
		log.Println("listening on", server.Addr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	case <-ctx.Done():
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return server.Shutdown(shutdownContext)
	}
}

func newServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 16,
	}
}

func signalContext() (context.Context, context.CancelFunc) {
	return signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func serverConfig() (string, api.Config) {
	return ":" + envOrDefault("PORT", "8080"), api.Config{
		CORSAllowedOrigin: envOrDefault("CORS_ALLOWED_ORIGIN", "*"),
	}
}
