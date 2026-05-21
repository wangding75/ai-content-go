package system_test

import (
	"context"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/config"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/store"
)

func TestHealthUsesFallbackVersionWhenBuildCommitMissing(t *testing.T) {
	svc := system.NewService(config.Config{AppName: "ai-content-go", Environment: "test"}, store.NoopDBChecker{}, store.NoopMigrationReader{})

	got, err := svc.Health(context.Background())
	if err != nil {
		t.Fatalf("health returned error: %v", err)
	}
	if got.Version == "" {
		t.Fatalf("expected non-empty fallback version")
	}
}
