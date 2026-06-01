package app_test

import (
	"log/slog"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/app"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/config"
)

func TestNewServerRejectsMissingHTTPAddr(t *testing.T) {
	_, err := app.NewServer(config.Config{AppName: "ai-content-go"}, slog.Default())
	if err == nil {
		t.Fatalf("expected missing HTTPAddr to be rejected")
	}
}
