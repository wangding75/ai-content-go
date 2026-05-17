package novel

import (
	"context"
	"errors"
)

var (
	ErrValidation          = errors.New("validation error")
	ErrNotFound            = errors.New("not found")
	ErrForbidden           = errors.New("forbidden")
	ErrConflict            = errors.New("conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrWorkflowRunFailed   = errors.New("workflow run failed")
)

type Service interface {
	CreatePlanningRun(ctx context.Context, projectID string, req CreatePlanningRunRequest, workflowRunID string, idempotencyKey string) (CreatePlanningRunResponse, error)
	ListPlanningRuns(ctx context.Context, projectID string, req ListPlanningRunsRequest) (PagedPlanningRunsResponse, error)
	GetPlanningRun(ctx context.Context, projectID, runID string) (PlanningRunDetailResponse, error)
	ConfirmTopic(ctx context.Context, projectID, topicID string, req ConfirmTopicRequest, idempotencyKey string) (ConfirmTopicResponse, error)
	GetWorldview(ctx context.Context, projectID string) (WorldviewResponse, error)
	UpdateWorldview(ctx context.Context, projectID string, req UpdateWorldviewRequest) (UpdateWorldviewResponse, error)
	ListCharacters(ctx context.Context, projectID string, req ListCharactersRequest) (PagedCharactersResponse, error)
	CreateCharacter(ctx context.Context, projectID string, req CreateCharacterRequest) (CreateCharacterResponse, error)
	ListArcs(ctx context.Context, projectID string, req ListArcsRequest) (PagedArcsResponse, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) CreatePlanningRun(ctx context.Context, projectID string, req CreatePlanningRunRequest, workflowRunID string, idempotencyKey string) (CreatePlanningRunResponse, error) {
	return CreatePlanningRunResponse{}, ErrWorkflowRunFailed
}

func (s *service) ListPlanningRuns(ctx context.Context, projectID string, req ListPlanningRunsRequest) (PagedPlanningRunsResponse, error) {
	return PagedPlanningRunsResponse{}, ErrNotFound
}

func (s *service) GetPlanningRun(ctx context.Context, projectID, runID string) (PlanningRunDetailResponse, error) {
	return PlanningRunDetailResponse{}, ErrNotFound
}

func (s *service) ConfirmTopic(ctx context.Context, projectID, topicID string, req ConfirmTopicRequest, idempotencyKey string) (ConfirmTopicResponse, error) {
	return ConfirmTopicResponse{}, ErrConflict
}

func (s *service) GetWorldview(ctx context.Context, projectID string) (WorldviewResponse, error) {
	return WorldviewResponse{}, ErrNotFound
}

func (s *service) UpdateWorldview(ctx context.Context, projectID string, req UpdateWorldviewRequest) (UpdateWorldviewResponse, error) {
	return UpdateWorldviewResponse{}, ErrValidation
}

func (s *service) ListCharacters(ctx context.Context, projectID string, req ListCharactersRequest) (PagedCharactersResponse, error) {
	return PagedCharactersResponse{}, ErrNotFound
}

func (s *service) CreateCharacter(ctx context.Context, projectID string, req CreateCharacterRequest) (CreateCharacterResponse, error) {
	return CreateCharacterResponse{}, ErrValidation
}

func (s *service) ListArcs(ctx context.Context, projectID string, req ListArcsRequest) (PagedArcsResponse, error) {
	return PagedArcsResponse{}, ErrNotFound
}
