package main

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestNewServerConfiguration(t *testing.T) {
	server := newServer(":8080", http.NewServeMux())
	if server.ReadHeaderTimeout != 5*time.Second ||
		server.ReadTimeout != 10*time.Second ||
		server.WriteTimeout != 10*time.Second ||
		server.IdleTimeout != 60*time.Second ||
		server.MaxHeaderBytes != 1<<16 {
		t.Fatalf("unexpected server configuration: %+v", server)
	}
}

func TestRunStopsWhenContextIsCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := run(ctx, "127.0.0.1:0", http.NewServeMux()); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}

func TestEnvOrDefault(t *testing.T) {
	t.Setenv("TEST_PORT", "9090")
	if got := envOrDefault("TEST_PORT", "8080"); got != "9090" {
		t.Fatalf("envOrDefault() = %q, want 9090", got)
	}
	t.Setenv("TEST_PORT", "")
	if got := envOrDefault("TEST_PORT", "8080"); got != "8080" {
		t.Fatalf("envOrDefault() = %q, want 8080", got)
	}
}

func TestServerConfigFromEnvironment(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("CORS_ALLOWED_ORIGIN", "https://example.test")
	addr, config := serverConfig()
	if addr != ":9090" || config.CORSAllowedOrigin != "https://example.test" {
		t.Fatalf("serverConfig() = %q, %#v", addr, config)
	}
}
