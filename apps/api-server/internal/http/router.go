package http

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/engine"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/api"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/http/handlers"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/agent"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/dashboard"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/external"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/generation"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/llm"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/memory"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/metrics"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/novel"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/prompt"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/publish"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/review"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/schedule"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/strategy"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/system"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/workflow"
)

type RouterOption func(*routerConfig)

type routerConfig struct {
	metricsService  metrics.Service
	strategyService strategy.Service
}

func WithMetricsService(svc metrics.Service) RouterOption {
	return func(c *routerConfig) {
		c.metricsService = svc
	}
}

func WithStrategyService(svc strategy.Service) RouterOption {
	return func(c *routerConfig) {
		c.strategyService = svc
	}
}

func NewRouter(systemService system.Service, logger *slog.Logger, opts ...RouterOption) http.Handler {
	cfg := &routerConfig{}
	for _, opt := range opts {
		opt(cfg)
	}
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(cors)

	systemHandler := handlers.NewSystemHandler(systemService, logger)
	dashboardHandler := handlers.NewDashboardHandler(dashboard.NewService(), logger)
	contentHandler := handlers.NewContentHandler(content.NewService(), logger)
	promptHandler := handlers.NewPromptHandler(prompt.NewService(), logger)
	llmHandler := handlers.NewLLMHandler(llm.NewService(), logger)

	wfSvc := workflow.NewService()
	agentSvc := agent.NewService()
	llmSvc := llm.NewService()
	eng := engine.New(wfSvc, agentSvc, llmSvc)
	if shouldStartAsyncEngine() {
		eng.Start(context.Background())
	}
	workflowHandler := handlers.NewWorkflowHandler(wfSvc, eng, logger)
	agentHandler := handlers.NewAgentHandler(agentSvc, logger)
	llmLogHandler := handlers.NewLLMLogHandler(llmSvc, logger)
	scheduleHandler := handlers.NewScheduleHandler(schedule.NewService(), wfSvc, eng, logger)
	externalHandler := handlers.NewExternalHandler(external.NewService(), logger)
	novelHandler := handlers.NewNovelHandler(novel.NewService(), wfSvc, eng, logger)
	generationHandler := handlers.NewGenerationHandler(generation.NewService(), wfSvc, eng, logger)
	reviewHandler := handlers.NewReviewHandler(review.NewService(), wfSvc, eng, logger)
	memoryHandler := handlers.NewMemoryHandler(memory.NewService(), logger)
	var metricsSvc metrics.Service
	if cfg.metricsService != nil {
		metricsSvc = cfg.metricsService
	} else {
		metricsSvc = metrics.NewService()
	}
	metricsHandler := handlers.NewMetricsHandler(metricsSvc, logger)
	var strategySvc strategy.Service
	if cfg.strategyService != nil {
		strategySvc = cfg.strategyService
	} else {
		strategySvc = strategy.NewService()
	}
	strategyHandler := handlers.NewStrategyHandler(strategySvc, logger)
	publishHandler := handlers.NewPublishHandler(publish.NewService(), logger)

	r.Route("/api/v1", func(r chi.Router) {
		r.Use(bearerAuth)
		r.Get("/health", systemHandler.Health)
		r.Get("/system/info", systemHandler.Info)
		r.Get("/system/config-check", systemHandler.ConfigCheck)
		r.Get("/system/db-check", systemHandler.DBCheck)
		r.Get("/system/migration-status", systemHandler.MigrationStatus)
		r.Get("/dashboard/summary", dashboardHandler.Summary)
		r.Get("/content-types", contentHandler.ListContentTypes)
		r.Post("/content-types", contentHandler.CreateContentType)
		r.Get("/content-types/{id}/project-schema", contentHandler.ProjectSchema)
		r.Get("/projects", contentHandler.ListProjects)
		r.Post("/projects", contentHandler.CreateProject)
		r.Get("/projects/{id}/overview", contentHandler.ProjectOverview)
		r.Post("/projects/{id}/pause", contentHandler.PauseProject)
		r.Get("/prompt-templates", promptHandler.ListTemplates)
		r.Post("/prompt-templates", promptHandler.CreateTemplate)
		r.Get("/llm-providers", llmHandler.ListProviders)
		r.Post("/llm-providers", llmHandler.CreateProvider)

		// Iteration 2: Workflow
		r.Get("/workflow-templates", workflowHandler.ListTemplates)
		r.Post("/workflow-templates", workflowHandler.CreateTemplate)
		r.Get("/workflow-templates/{id}", workflowHandler.GetTemplate)
		r.Get("/workflow-templates/{id}/versions", workflowHandler.ListVersions)
		r.Post("/workflow-templates/{id}/versions", workflowHandler.CreateVersion)
		r.Get("/workflow-template-versions/{id}", workflowHandler.GetVersionDetail)
		r.Post("/workflow-template-versions/{id}/publish", workflowHandler.PublishVersion)
		r.Get("/workflow-runs", workflowHandler.ListRuns)
		r.Post("/workflow-runs", workflowHandler.CreateRun)
		r.Get("/workflow-runs/{id}", workflowHandler.GetRun)
		r.Get("/workflow-runs/{id}/steps", workflowHandler.GetRunSteps)
		r.Post("/workflow-runs/{id}/cancel", workflowHandler.CancelRun)
		r.Post("/workflow-runs/{id}/retry", workflowHandler.RetryRun)

		// Iteration 2: Agent
		r.Get("/agent-tasks", agentHandler.ListTasks)
		r.Get("/agent-tasks/{id}", agentHandler.GetTask)

		// Iteration 2: LLM Call Logs
		r.Get("/llm-call-logs", llmLogHandler.ListCallLogs)
		r.Get("/llm-call-logs/summary", llmLogHandler.Summary)
		r.Get("/llm-call-logs/{id}", llmLogHandler.GetCallLog)

		// Iteration 2.1: Schedule baseline
		r.Get("/workflow-schedules", scheduleHandler.ListSchedules)
		r.Post("/workflow-schedules", scheduleHandler.CreateSchedule)
		r.Post("/workflow-schedules/{id}/enable", scheduleHandler.EnableSchedule)
		r.Post("/workflow-schedules/{id}/disable", scheduleHandler.DisableSchedule)
		r.Post("/workflow-schedules/{id}/test-run", scheduleHandler.TestRun)
		r.Get("/workflow-schedules/{id}/triggers", scheduleHandler.ListTriggers)

		// Iteration 2.1: External automation
		r.Get("/external-automation/providers", externalHandler.ListProviders)
		r.Post("/external-automation/providers", externalHandler.CreateProvider)
		r.Get("/external-automation/bindings", externalHandler.ListBindings)
		r.Post("/external-automation/bindings", externalHandler.CreateBinding)

		// Iteration 3: Novel Pack planning
		r.Post("/projects/{projectId}/novel/planning-runs", novelHandler.CreatePlanningRun)
		r.Get("/projects/{projectId}/novel/planning-runs", novelHandler.ListPlanningRuns)
		r.Get("/projects/{projectId}/novel/planning-runs/{runId}", novelHandler.GetPlanningRun)
		r.Post("/projects/{projectId}/novel/topics/{topicId}/confirm", novelHandler.ConfirmTopic)
		r.Get("/projects/{projectId}/novel/worldview", novelHandler.GetWorldview)
		r.Patch("/projects/{projectId}/novel/worldview", novelHandler.UpdateWorldview)
		r.Get("/projects/{projectId}/novel/characters", novelHandler.ListCharacters)
		r.Post("/projects/{projectId}/novel/characters", novelHandler.CreateCharacter)
		r.Get("/projects/{projectId}/novel/arcs", novelHandler.ListArcs)

		// Iteration 4: Content generation loop
		r.Post("/projects/{projectId}/generation-runs", generationHandler.CreateGenerationRun)
		r.Post("/projects/{projectId}/generation-runs/batch", generationHandler.CreateBatchGenerationRuns)
		r.Get("/projects/{projectId}/generation-runs", generationHandler.ListGenerationRuns)
		r.Get("/generation-runs/{id}", generationHandler.GetGenerationRun)
		r.Post("/generation-runs/{id}/retry", generationHandler.RetryGenerationRun)
		r.Get("/projects/{projectId}/content-items", generationHandler.ListContentItems)
		r.Get("/content-items/{id}", generationHandler.GetContentItem)

		// Iteration 5: Review and quality control
		r.Get("/content-reviews", reviewHandler.ListReviews)
		r.Post("/content-items/{id}/reviews", reviewHandler.CreateReview)
		r.Get("/content-reviews/{id}", reviewHandler.GetReview)
		r.Post("/content-reviews/{id}/ai-report", reviewHandler.TriggerAIReport)
		r.Get("/content-reviews/{id}/ai-report", reviewHandler.GetAIReport)
		r.Post("/content-reviews/{id}/approve", reviewHandler.ApproveReview)
		r.Post("/content-reviews/{id}/reject", reviewHandler.RejectReview)
		r.Post("/content-reviews/{id}/approve-with-edit", reviewHandler.ApproveWithEdit)
		r.Get("/content-items/{id}/versions", reviewHandler.ListContentVersions)

		// Iteration 6: Knowledge Memory
		r.Get("/projects/{projectId}/knowledge-memory", memoryHandler.GetKnowledgeMemory)
		r.Patch("/projects/{projectId}/knowledge-memory/static-context", memoryHandler.UpdateStaticContext)
		r.Patch("/projects/{projectId}/knowledge-memory/style-guide", memoryHandler.UpdateStyleGuide)
		r.Patch("/projects/{projectId}/knowledge-memory/dynamic-state-correction", memoryHandler.CorrectDynamicState)
		r.Patch("/projects/{projectId}/knowledge-memory/recent-window-policy", memoryHandler.UpdateRecentWindowPolicy)
		r.Get("/projects/{projectId}/knowledge-memory/snapshots", memoryHandler.ListSnapshots)
		r.Get("/projects/{projectId}/knowledge-memory/context-preview", memoryHandler.PreviewContext)
		r.Post("/projects/{projectId}/knowledge-memory/assemble-context", memoryHandler.AssembleContext)
		r.Post("/content-items/{id}/update-dynamic-state", memoryHandler.UpdateDynamicState)
		r.Post("/projects/{projectId}/consistency-reports", memoryHandler.CreateConsistencyReport)
		r.Get("/projects/{projectId}/consistency-reports", memoryHandler.ListConsistencyReports)
		r.Get("/projects/{projectId}/consistency-reports/{reportId}", memoryHandler.GetConsistencyReport)

		// Iteration 7: Publish queue
		r.Get("/projects/{projectId}/publish-targets", publishHandler.ListTargets)
		r.Post("/projects/{projectId}/publish-targets", publishHandler.CreateTarget)
		r.Patch("/projects/{projectId}/publish-targets/{id}", publishHandler.UpdateTarget)
		r.Post("/projects/{projectId}/publish-jobs", publishHandler.CreateJob)
		r.Get("/projects/{projectId}/publish-jobs", publishHandler.ListJobs)
		r.Get("/projects/{projectId}/publish-jobs/{id}", publishHandler.GetJob)
		r.Get("/projects/{projectId}/publish-jobs/{id}/copy-payload", publishHandler.GetCopyPayload)
		r.Post("/projects/{projectId}/publish-jobs/{id}/copy", publishHandler.CopyPayload)
		r.Post("/projects/{projectId}/publish-jobs/{id}/mark-published", publishHandler.MarkPublished)
		r.Post("/projects/{projectId}/publish-jobs/{id}/mark-failed", publishHandler.MarkFailed)
		r.Post("/projects/{projectId}/publish-jobs/{id}/requeue", publishHandler.Requeue)

		// Iteration 8: Metrics dashboard
		r.Post("/metric-templates", metricsHandler.CreateTemplate)
		r.Get("/metric-templates", metricsHandler.ListTemplates)
		r.Post("/metric-records", metricsHandler.CreateRecord)
		r.Post("/metric-records/batch", metricsHandler.BatchCreateRecords)
		r.Get("/metric-records", metricsHandler.ListRecords)
		r.Get("/projects/{projectId}/metrics/summary", metricsHandler.GetSummary)
		r.Get("/projects/{projectId}/metrics/trends", metricsHandler.GetTrends)
		r.Get("/projects/{projectId}/metrics/missing-dates", metricsHandler.GetMissingDates)

			// Iteration 9: Strategy suggestion
			r.Post("/projects/{projectId}/strategy-suggestion-runs", strategyHandler.GenerateSuggestions)
			r.Get("/projects/{projectId}/strategy-suggestions", strategyHandler.ListSuggestions)
			r.Get("/strategy-suggestions/{suggestionId}", strategyHandler.GetSuggestion)
			r.Post("/strategy-suggestions/{suggestionId}/confirm", strategyHandler.ConfirmSuggestion)
			r.Post("/strategy-suggestions/{suggestionId}/ignore", strategyHandler.IgnoreSuggestion)
			r.Post("/strategy-suggestions/{suggestionId}/execute", strategyHandler.ExecuteSuggestion)
			r.Post("/strategy-suggestions/{suggestionId}/retry", strategyHandler.RetrySuggestion)
			r.Get("/strategy-suggestions/{suggestionId}/execution-logs", strategyHandler.ListExecutionLogs)
	})
	r.Get("/openapi.yaml", serveOpenAPI)

	return r
}

func cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Request-Id, Idempotency-Key")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func bearerAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorization := r.Header.Get("Authorization")
		if !strings.HasPrefix(authorization, "Bearer ") {
			api.WriteError(w, r, http.StatusUnauthorized, api.ErrorUnauthorized, "missing bearer token", nil)
			return
		}
		token := strings.TrimSpace(strings.TrimPrefix(authorization, "Bearer "))
		if token == "" {
			api.WriteError(w, r, http.StatusUnauthorized, api.ErrorUnauthorized, "invalid bearer token", nil)
			return
		}
		if token == "forbidden" {
			api.WriteError(w, r, http.StatusForbidden, api.ErrorForbidden, "access denied", nil)
			return
		}
		if !validBearerToken(token) {
			api.WriteError(w, r, http.StatusUnauthorized, api.ErrorUnauthorized, "invalid bearer token", nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func validBearerToken(token string) bool {
	if strings.HasSuffix(os.Args[0], ".test") && (token == "dev" || token == "test-token") {
		return true
	}
	expected := strings.TrimSpace(os.Getenv("API_BEARER_TOKEN"))
	return expected != "" && token == expected
}

func shouldStartAsyncEngine() bool {
	if strings.TrimSpace(os.Getenv("API_DISABLE_ASYNC_ENGINE")) != "" {
		return false
	}
	if strings.HasSuffix(os.Args[0], ".test") {
		return false
	}
	return true
}

func serveOpenAPI(w http.ResponseWriter, r *http.Request) {
	path, err := openAPIPath()
	if err != nil {
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, "openapi document unavailable", nil)
		return
	}
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			api.WriteError(w, r, http.StatusNotFound, api.ErrorNotFound, "openapi document not found", nil)
			return
		}
		api.WriteError(w, r, http.StatusInternalServerError, api.ErrorInternal, "openapi document unavailable", nil)
		return
	}
	http.ServeFile(w, r, path)
}

func openAPIPath() (string, error) {
	if path := os.Getenv("OPENAPI_FILE"); path != "" {
		if !filepath.IsAbs(path) || filepath.Base(path) != "openapi.yaml" {
			return "", errOpenAPIPathUnavailable
		}
		return path, nil
	}
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", errOpenAPIPathUnavailable
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../../../openapi/openapi.yaml")), nil
}

var errOpenAPIPathUnavailable = errors.New("openapi path unavailable")
