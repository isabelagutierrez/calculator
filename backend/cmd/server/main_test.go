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
	if err := run(ctx, http.NewServeMux()); err != nil {
		t.Fatalf("run() error = %v", err)
	}
}
