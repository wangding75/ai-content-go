package system_test

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/config"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/store"
)

func TestConfigCheckIncludesAllRequiredRuntimeKeys(t *testing.T) {
	svc := system.NewService(config.Config{AppName: "ai-content-go", Environment: "test"}, store.NoopDBChecker{}, store.NoopMigrationReader{})

	got, err := svc.ConfigCheck(context.Background())
	if err != nil {
		t.Fatalf("config check returned error: %v", err)
	}
	if got.Summary.Total < 3 {
		t.Fatalf("expected APP_NAME, APP_ENV, and DATABASE_URL checks, got %d", got.Summary.Total)
	}
}
