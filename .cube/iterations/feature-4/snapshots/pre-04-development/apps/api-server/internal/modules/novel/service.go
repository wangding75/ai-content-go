package novel

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"sync"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
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

type service struct {
	mu sync.RWMutex

	planningRuns []PlanningRunDetailResponse
	worldviews   map[string]WorldviewResponse
	characters   []CharacterResponse
	arcs         []ArcResponse
	idempotency  map[string]idempotentOperation

	runNext       int
	snapshotNext  int
	topicNext     int
	worldviewNext int
	characterNext int
	arcNext       int
	oplogNext     int
}

type idempotentOperation struct {
	projectID string
	kind      string
	payload   any
	response  any
}

func NewService() Service {
	return &service{
		worldviews:  make(map[string]WorldviewResponse),
		idempotency: make(map[string]idempotentOperation),
	}
}

func (s *service) CreatePlanningRun(ctx context.Context, projectID string, req CreatePlanningRunRequest, workflowRunID string, idempotencyKey string) (CreatePlanningRunResponse, error) {
	if projectID == "" || req.Genre == "" || req.Audience == "" || req.Count < 1 || req.TemplateVersionID == "" || workflowRunID == "" || idempotencyKey == "" {
		return CreatePlanningRunResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	payload := struct {
		Request       CreatePlanningRunRequest
		WorkflowRunID string
	}{Request: req, WorkflowRunID: workflowRunID}
	if op, ok := s.idempotency[idempotencyKey]; ok {
		if op.projectID != projectID || op.kind != "create_planning_run" || !reflect.DeepEqual(op.payload, payload) {
			return CreatePlanningRunResponse{}, ErrIdempotencyConflict
		}
		return op.response.(CreatePlanningRunResponse), nil
	}

	now := time.Now()
	s.runNext++
	s.snapshotNext++
	runID := fmt.Sprintf("planning-run-%d", s.runNext)
	snapshotID := fmt.Sprintf("planning-snapshot-%d", s.snapshotNext)
	topics := make([]TopicCandidateResponse, 0, req.Count)
	for i := 1; i <= req.Count; i++ {
		s.topicNext++
		topics = append(topics, TopicCandidateResponse{
			CandidateID:   fmt.Sprintf("topic-candidate-%d", s.topicNext),
			PlanningRunID: runID,
			SnapshotID:    snapshotID,
			Title:         fmt.Sprintf("%s candidate %d", req.Genre, i),
			Logline:       fmt.Sprintf("A %s story for %s readers", req.Genre, req.Audience),
			Status:        "candidate",
			Score:         90 - float64(i),
			Reason:        "Generated from Novel Pack planning input",
		})
	}

	detail := PlanningRunDetailResponse{
		PlanningRunResponse: PlanningRunResponse{
			ID:                runID,
			ProjectID:         projectID,
			WorkflowRunID:     workflowRunID,
			TemplateVersionID: req.TemplateVersionID,
			Status:            "pending",
			Genre:             req.Genre,
			Audience:          req.Audience,
			CandidateCount:    req.Count,
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		Topics:      topics,
		StepRuns:    []StepRunSummary{},
		AgentTasks:  []AgentTaskSummary{},
		LLMCallLogs: []LLMCallLogSummary{},
	}
	s.planningRuns = append(s.planningRuns, detail)
	resp := CreatePlanningRunResponse{PlanningRunID: runID, WorkflowRunID: workflowRunID, Status: "pending"}
	s.idempotency[idempotencyKey] = idempotentOperation{projectID: projectID, kind: "create_planning_run", payload: payload, response: resp}
	return resp, nil
}

func (s *service) ListPlanningRuns(ctx context.Context, projectID string, req ListPlanningRunsRequest) (PagedPlanningRunsResponse, error) {
	if projectID == "" {
		return PagedPlanningRunsResponse{}, ErrValidation
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []PlanningRunResponse{}
	for _, run := range s.planningRuns {
		if run.ProjectID != projectID {
			continue
		}
		if req.Status != "" && run.Status != req.Status {
			continue
		}
		items = append(items, run.PlanningRunResponse)
	}
	if req.Sort == "created_at" && req.Order == "desc" {
		sort.Slice(items, func(i, j int) bool { return items[i].CreatedAt.After(items[j].CreatedAt) })
	}
	page, pageSize, start, end := paginate(len(items), req.Page, req.PageSize)
	return PagedPlanningRunsResponse{Items: items[start:end], Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: end < len(items)}}, nil
}

func (s *service) GetPlanningRun(ctx context.Context, projectID, runID string) (PlanningRunDetailResponse, error) {
	if projectID == "" || runID == "" {
		return PlanningRunDetailResponse{}, ErrValidation
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, run := range s.planningRuns {
		if run.ID != runID {
			continue
		}
		if run.ProjectID != projectID {
			return PlanningRunDetailResponse{}, ErrForbidden
		}
		return run, nil
	}
	return PlanningRunDetailResponse{}, ErrNotFound
}

func (s *service) ConfirmTopic(ctx context.Context, projectID, topicID string, req ConfirmTopicRequest, idempotencyKey string) (ConfirmTopicResponse, error) {
	if projectID == "" || topicID == "" || idempotencyKey == "" {
		return ConfirmTopicResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	payload := struct {
		TopicID string
		Note    string
	}{TopicID: topicID, Note: req.Note}
	if op, ok := s.idempotency[idempotencyKey]; ok {
		if op.projectID != projectID || op.kind != "confirm_topic" || !reflect.DeepEqual(op.payload, payload) {
			return ConfirmTopicResponse{}, ErrIdempotencyConflict
		}
		return op.response.(ConfirmTopicResponse), nil
	}

	for runIndex := range s.planningRuns {
		if s.planningRuns[runIndex].ProjectID != projectID {
			continue
		}
		for topicIndex := range s.planningRuns[runIndex].Topics {
			topic := &s.planningRuns[runIndex].Topics[topicIndex]
			if topic.CandidateID != topicID {
				continue
			}
			if topic.Status != "candidate" {
				return ConfirmTopicResponse{}, ErrConflict
			}
			s.oplogNext++
			confirmedID := fmt.Sprintf("confirmed-topic-%d", s.oplogNext)
			topic.Status = "confirmed"
			topic.ConfirmedTopicID = confirmedID
			resp := ConfirmTopicResponse{ConfirmedTopicID: confirmedID, PreviousStatus: "candidate", CurrentStatus: "confirmed", OperationLogID: fmt.Sprintf("oplog-%d", s.oplogNext)}
			s.idempotency[idempotencyKey] = idempotentOperation{projectID: projectID, kind: "confirm_topic", payload: payload, response: resp}
			return resp, nil
		}
	}
	return ConfirmTopicResponse{}, ErrNotFound
}

func (s *service) GetWorldview(ctx context.Context, projectID string) (WorldviewResponse, error) {
	if projectID == "" {
		return WorldviewResponse{}, ErrValidation
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	if worldview, ok := s.worldviews[projectID]; ok {
		return worldview, nil
	}
	return WorldviewResponse{ProjectID: projectID, VersionID: "worldview-0", Version: 0, Worldview: map[string]any{}, ForbiddenRules: []string{}}, nil
}

func (s *service) UpdateWorldview(ctx context.Context, projectID string, req UpdateWorldviewRequest) (UpdateWorldviewResponse, error) {
	if projectID == "" || req.Worldview == nil {
		return UpdateWorldviewResponse{}, ErrValidation
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	current := s.worldviews[projectID]
	s.worldviewNext++
	s.oplogNext++
	version := current.Version + 1
	versionID := fmt.Sprintf("worldview-%d", s.worldviewNext)
	s.worldviews[projectID] = WorldviewResponse{ProjectID: projectID, VersionID: versionID, Version: version, Worldview: req.Worldview, ForbiddenRules: req.ForbiddenRules}
	return UpdateWorldviewResponse{VersionID: versionID, OperationLogID: fmt.Sprintf("oplog-%d", s.oplogNext)}, nil
}

func (s *service) ListCharacters(ctx context.Context, projectID string, req ListCharactersRequest) (PagedCharactersResponse, error) {
	if projectID == "" {
		return PagedCharactersResponse{}, ErrValidation
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []CharacterResponse{}
	for _, item := range s.characters {
		if item.ProjectID != projectID {
			continue
		}
		if req.Role != "" && item.Role != req.Role {
			continue
		}
		items = append(items, item)
	}
	page, pageSize, start, end := paginate(len(items), req.Page, req.PageSize)
	return PagedCharactersResponse{Items: items[start:end], Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: end < len(items)}}, nil
}

func (s *service) CreateCharacter(ctx context.Context, projectID string, req CreateCharacterRequest) (CreateCharacterResponse, error) {
	if projectID == "" || req.Name == "" || req.Role == "" || req.Profile == nil {
		return CreateCharacterResponse{}, ErrValidation
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.characterNext++
	s.oplogNext++
	characterID := fmt.Sprintf("character-%d", s.characterNext)
	s.characters = append(s.characters, CharacterResponse{CharacterID: characterID, ProjectID: projectID, Name: req.Name, Role: req.Role, Profile: req.Profile})
	return CreateCharacterResponse{CharacterID: characterID, OperationLogID: fmt.Sprintf("oplog-%d", s.oplogNext)}, nil
}

func (s *service) ListArcs(ctx context.Context, projectID string, req ListArcsRequest) (PagedArcsResponse, error) {
	if projectID == "" {
		return PagedArcsResponse{}, ErrValidation
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	items := []ArcResponse{}
	for _, item := range s.arcs {
		if item.ProjectID == projectID {
			items = append(items, item)
		}
	}
	if req.Sort == "order_index" {
		sort.Slice(items, func(i, j int) bool {
			if req.Order == "desc" {
				return items[i].OrderIndex > items[j].OrderIndex
			}
			return items[i].OrderIndex < items[j].OrderIndex
		})
	}
	page, pageSize, start, end := paginate(len(items), req.Page, req.PageSize)
	return PagedArcsResponse{Items: items[start:end], Pagination: content.PaginationResponse{Page: page, PageSize: pageSize, Total: len(items), HasNext: end < len(items)}}, nil
}

func paginate(total int, page int, pageSize int) (int, int, int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	return page, pageSize, start, end
}
