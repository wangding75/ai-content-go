package system_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/config"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/store"
)

type failingDBChecker struct{}

func (failingDBChecker) Check(ctx context.Context) (store.DBCheckResult, error) {
	return store.DBCheckResult{}, errors.New("connection refused")
}

func TestDBCheckWrapsDependencyErrorWithContext(t *testing.T) {
	svc := system.NewService(config.Config{AppName: "ai-content-go", HTTPAddr: ":8080"}, failingDBChecker{}, store.NoopMigrationReader{})

	_, err := svc.DBCheck(context.Background())
	if err == nil {
		t.Fatalf("expected db check error")
	}
	if !strings.Contains(err.Error(), "db check") {
		t.Fatalf("expected wrapped db check context, got %q", err.Error())
	}
}
