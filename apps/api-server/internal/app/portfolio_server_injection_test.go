package app_test

import (
	"log/slog"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/app"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/config"
)

// @Test
func TestTask04NewServerAcceptsHTTPAddrWithoutDB(t *testing.T) {
	srv, err := app.NewServer(config.Config{AppName: "ai-content-go", HTTPAddr: ":0"}, slog.Default())
	if err != nil {
		t.Fatalf("expected server creation to succeed without DB: %v", err)
	}
	if srv == nil {
		t.Fatalf("expected non-nil server")
	}
}
