package generation

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
)

var (
	ErrValidation          = errors.New("validation error")
	ErrNotFound            = errors.New("not found")
	ErrConflict            = errors.New("conflict")
	ErrIdempotencyConflict = errors.New("idempotency conflict")
	ErrWorkflowRunFailed   = errors.New("workflow run failed")
	ErrLLMProviderError    = errors.New("llm provider error")
)

type Service interface {
	ValidateGenerationRun(ctx context.Context, projectID string, req CreateGenerationRunRequest, idempotencyKey string) error
	CreateGenerationRun(ctx context.Context, projectID string, req CreateGenerationRunRequest, workflowRunID string, idempotencyKey string) (CreateGenerationRunResponse, error)
	CreateBatchGenerationRuns(ctx context.Context, projectID string, req CreateBatchGenerationRunsRequest, workflowRunIDs []string, idempotencyKey string) (CreateBatchGenerationRunsResponse, error)
	ListGenerationRuns(ctx context.Context, projectID string, req ListGenerationRunsRequest) (PagedGenerationRunsResponse, error)
	GetGenerationRun(ctx context.Context, id string) (GenerationRunDetailResponse, error)
	ValidateRetryGenerationRun(ctx context.Context, id string, req RetryGenerationRunRequest, idempotencyKey string) error
	RetryGenerationRun(ctx context.Context, id string, req RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (RetryGenerationRunResponse, error)
	ReconcileWorkflowResult(ctx context.Context, workflowRunID string) error
	ListContentItems(ctx context.Context, projectID string, req ListContentItemsRequest) (PagedContentItemsResponse, error)
	GetContentItem(ctx context.Context, id string) (ContentItemDetailResponse, error)
}

type service struct {
	mu          sync.Mutex
	nextRun     int
	nextItem    int
	runs        map[string]GenerationRunDetailResponse
	items       map[string]ContentItemDetailResponse
	idempotency map[string]idempotencyRecord
}

type idempotencyRecord struct {
	payload string
	result  any
	pending bool
}

func NewService() Service {
	return &service{
		runs:        map[string]GenerationRunDetailResponse{},
		items:       map[string]ContentItemDetailResponse{},
		idempotency: map[string]idempotencyRecord{},
	}
}

func (s *service) ValidateGenerationRun(ctx context.Context, projectID string, req CreateGenerationRunRequest, idempotencyKey string) error {
	if projectID == "" || idempotencyKey == "" || req.TargetCount <= 0 || req.StartSequenceNo <= 0 || req.TemplateVersionID == "" {
		return ErrValidation
	}
	return nil
}

func (s *service) ReplayGenerationRun(ctx context.Context, projectID string, req CreateGenerationRunRequest, idempotencyKey string) (CreateGenerationRunResponse, bool, error) {
	if err := s.ValidateGenerationRun(ctx, projectID, req, idempotencyKey); err != nil {
		return CreateGenerationRunResponse{}, false, err
	}
	payload, err := idempotencyPayload(projectID, req)
	if err != nil {
		return CreateGenerationRunResponse{}, false, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.idempotency["create:"+projectID+":"+idempotencyKey]
	if !ok {
		s.idempotency["create:"+projectID+":"+idempotencyKey] = idempotencyRecord{payload: payload, pending: true}
		return CreateGenerationRunResponse{}, false, nil
	}
	if record.payload != payload {
		return CreateGenerationRunResponse{}, false, ErrIdempotencyConflict
	}
	if record.pending {
		return CreateGenerationRunResponse{}, false, ErrConflict
	}
	result, ok := record.result.(CreateGenerationRunResponse)
	if !ok {
		return CreateGenerationRunResponse{}, false, ErrConflict
	}
	return result, true, nil
}

func (s *service) ReleaseGenerationRunReservation(ctx context.Context, projectID string, req CreateGenerationRunRequest, idempotencyKey string) error {
	return s.releaseReservation("create:"+projectID+":"+idempotencyKey, projectID, req)
}

func (s *service) ReleaseBatchGenerationRunsReservation(ctx context.Context, projectID string, req CreateBatchGenerationRunsRequest, idempotencyKey string) error {
	return s.releaseReservation("batch:"+projectID+":"+idempotencyKey, projectID, req)
}

func (s *service) ReleaseRetryGenerationRunReservation(ctx context.Context, id string, req RetryGenerationRunRequest, idempotencyKey string) error {
	return s.releaseReservation("retry:"+id+":"+idempotencyKey, id, req)
}

func (s *service) releaseReservation(key string, parts ...any) error {
	payload, err := idempotencyPayload(parts...)
	if err != nil {
		return ErrValidation
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	record, ok := s.idempotency[key]
	if !ok {
		return nil
	}
	if record.payload != payload {
		return ErrIdempotencyConflict
	}
	s.removeIdempotencyResult(record.result)
	delete(s.idempotency, key)
	return nil
}

func (s *service) removeIdempotencyResult(result any) {
	switch value := result.(type) {
	case CreateGenerationRunResponse:
		s.removeGenerationRun(value.GenerationRunID)
	case CreateBatchGenerationRunsResponse:
		for _, generationRunID := range value.GenerationRunIDs {
			s.removeGenerationRun(generationRunID)
		}
	case RetryGenerationRunResponse:
		s.removeGenerationRun(value.NewGenerationRunID)
	}
}

func (s *service) removeGenerationRun(generationRunID string) {
	if generationRunID == "" {
		return
	}
	delete(s.runs, generationRunID)
	for itemID, item := range s.items {
		if item.GenerationRunID == generationRunID {
			delete(s.items, itemID)
		}
	}
}

func (s *service) CreateGenerationRun(ctx context.Context, projectID string, req CreateGenerationRunRequest, workflowRunID string, idempotencyKey string) (CreateGenerationRunResponse, error) {
	if err := s.ValidateGenerationRun(ctx, projectID, req, idempotencyKey); err != nil {
		return CreateGenerationRunResponse{}, err
	}
	if workflowRunID == "" {
		return CreateGenerationRunResponse{}, ErrWorkflowRunFailed
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	payload, err := idempotencyPayload(projectID, req)
	if err != nil {
		return CreateGenerationRunResponse{}, ErrValidation
	}
	key := "create:" + projectID + ":" + idempotencyKey
	if record, ok := s.idempotency[key]; ok {
		if record.payload != payload {
			return CreateGenerationRunResponse{}, ErrIdempotencyConflict
		}
		if !record.pending {
			result, ok := record.result.(CreateGenerationRunResponse)
			if !ok {
				return CreateGenerationRunResponse{}, ErrConflict
			}
			return result, nil
		}
	}

	s.nextRun++
	s.nextItem++
	now := time.Now().UTC()
	runID := fmt.Sprintf("genrun-%d", s.nextRun)
	itemID := fmt.Sprintf("ci-%d", s.nextItem)
	item := ContentItemResponse{
		ID:              itemID,
		ProjectID:       projectID,
		GenerationRunID: runID,
		ContentTypeCode: "novel",
		Status:          ContentItemPlanned,
		SequenceNo:      req.StartSequenceNo,
		Title:           fmt.Sprintf("Content %d", req.StartSequenceNo),
		Version:         1,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	detail := GenerationRunDetailResponse{
		GenerationRunResponse: GenerationRunResponse{
			ID:                runID,
			ProjectID:         projectID,
			WorkflowRunID:     workflowRunID,
			TemplateVersionID: req.TemplateVersionID,
			Status:            GenerationRunPending,
			TriggerType:       "manual",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
		ContentItems: []ContentItemResponse{item},
	}
	s.runs[runID] = detail
	s.items[itemID] = ContentItemDetailResponse{
		ContentItemResponse: item,
		Body:                "generated body",
		Metadata:            map[string]any{},
		Extension: NovelChapterExtensionResponse{
			ConfirmedTopicID:   req.ConfirmedTopicID,
			WorldviewVersionID: req.WorldviewVersionID,
			ArcID:              req.ArcID,
			ChapterNo:          req.StartSequenceNo,
			Script:             map[string]any{},
		},
	}
	result := CreateGenerationRunResponse{GenerationRunID: runID, WorkflowRunID: workflowRunID, Status: GenerationRunPending}
	s.idempotency[key] = idempotencyRecord{payload: payload, result: result}
	return result, nil
}

func (s *service) ReplayBatchGenerationRuns(ctx context.Context, projectID string, req CreateBatchGenerationRunsRequest, idempotencyKey string) (CreateBatchGenerationRunsResponse, bool, error) {
	if projectID == "" || idempotencyKey == "" || req.BatchSize <= 0 || req.TemplateVersionID == "" {
		return CreateBatchGenerationRunsResponse{}, false, ErrValidation
	}
	payload, err := idempotencyPayload(projectID, req)
	if err != nil {
		return CreateBatchGenerationRunsResponse{}, false, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.idempotency["batch:"+projectID+":"+idempotencyKey]
	if !ok {
		s.idempotency["batch:"+projectID+":"+idempotencyKey] = idempotencyRecord{payload: payload, pending: true}
		return CreateBatchGenerationRunsResponse{}, false, nil
	}
	if record.payload != payload {
		return CreateBatchGenerationRunsResponse{}, false, ErrIdempotencyConflict
	}
	if record.pending {
		return CreateBatchGenerationRunsResponse{}, false, ErrConflict
	}
	result, ok := record.result.(CreateBatchGenerationRunsResponse)
	if !ok {
		return CreateBatchGenerationRunsResponse{}, false, ErrConflict
	}
	return result, true, nil
}

func (s *service) validateBatchGenerationRun(projectID string, req CreateBatchGenerationRunsRequest, workflowRunIDs []string, idempotencyKey string) error {
	batchCount := req.Range.EndSequenceNo - req.Range.StartSequenceNo + 1
	if projectID == "" || idempotencyKey == "" || req.TemplateVersionID == "" || req.BatchSize <= 0 || req.Range.StartSequenceNo <= 0 || req.Range.EndSequenceNo < req.Range.StartSequenceNo || req.BatchSize != batchCount || len(workflowRunIDs) != batchCount {
		return ErrValidation
	}
	return nil
}

func (s *service) CreateBatchGenerationRuns(ctx context.Context, projectID string, req CreateBatchGenerationRunsRequest, workflowRunIDs []string, idempotencyKey string) (CreateBatchGenerationRunsResponse, error) {
	if err := s.validateBatchGenerationRun(projectID, req, workflowRunIDs, idempotencyKey); err != nil {
		return CreateBatchGenerationRunsResponse{}, err
	}
	payload, err := idempotencyPayload(projectID, req)
	if err != nil {
		return CreateBatchGenerationRunsResponse{}, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	key := "batch:" + projectID + ":" + idempotencyKey
	if record, ok := s.idempotency[key]; ok {
		if record.payload != payload {
			return CreateBatchGenerationRunsResponse{}, ErrIdempotencyConflict
		}
		if !record.pending {
			result, ok := record.result.(CreateBatchGenerationRunsResponse)
			if !ok {
				return CreateBatchGenerationRunsResponse{}, ErrConflict
			}
			return result, nil
		}
	}

	now := time.Now().UTC()
	generationRunIDs := make([]string, 0, req.BatchSize)
	for i, workflowRunID := range workflowRunIDs {
		if workflowRunID == "" {
			return CreateBatchGenerationRunsResponse{}, ErrWorkflowRunFailed
		}
		s.nextRun++
		s.nextItem++
		runID := fmt.Sprintf("genrun-%d", s.nextRun)
		itemID := fmt.Sprintf("ci-%d", s.nextItem)
		sequenceNo := req.Range.StartSequenceNo + i
		item := ContentItemResponse{
			ID:              itemID,
			ProjectID:       projectID,
			GenerationRunID: runID,
			ContentTypeCode: "novel",
			Status:          ContentItemPlanned,
			SequenceNo:      sequenceNo,
			Title:           fmt.Sprintf("Content %d", sequenceNo),
			Version:         1,
			CreatedAt:       now,
			UpdatedAt:       now,
		}
		s.runs[runID] = GenerationRunDetailResponse{
			GenerationRunResponse: GenerationRunResponse{
				ID:                runID,
				ProjectID:         projectID,
				WorkflowRunID:     workflowRunID,
				TemplateVersionID: req.TemplateVersionID,
				Status:            GenerationRunPending,
				TriggerType:       "batch",
				CreatedAt:         now,
				UpdatedAt:         now,
			},
			ContentItems: []ContentItemResponse{item},
		}
		s.items[itemID] = ContentItemDetailResponse{
			ContentItemResponse: item,
			Body:                "generated body",
			Metadata:            map[string]any{},
			Extension:           NovelChapterExtensionResponse{ChapterNo: sequenceNo, Script: map[string]any{}},
		}
		generationRunIDs = append(generationRunIDs, runID)
	}
	result := CreateBatchGenerationRunsResponse{GenerationRunIDs: generationRunIDs, WorkflowRunIDs: workflowRunIDs, AcceptedCount: len(generationRunIDs)}
	s.idempotency[key] = idempotencyRecord{payload: payload, result: result}
	return result, nil
}

func (s *service) ListGenerationRuns(ctx context.Context, projectID string, req ListGenerationRunsRequest) (PagedGenerationRunsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := make([]GenerationRunResponse, 0)
	for _, run := range s.runs {
		if run.ProjectID == projectID && (req.Status == "" || run.Status == req.Status) {
			items = append(items, run.GenerationRunResponse)
		}
	}
	return PagedGenerationRunsResponse{Items: items, Pagination: pagination(len(items), req.PaginationRequest)}, nil
}

func (s *service) GetGenerationRun(ctx context.Context, id string) (GenerationRunDetailResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	run, ok := s.runs[id]
	if !ok {
		return GenerationRunDetailResponse{}, ErrNotFound
	}
	return run, nil
}

func (s *service) ValidateRetryGenerationRun(ctx context.Context, id string, req RetryGenerationRunRequest, idempotencyKey string) error {
	if id == "" || idempotencyKey == "" {
		return ErrValidation
	}
	return nil
}

func (s *service) ReplayRetryGenerationRun(ctx context.Context, id string, req RetryGenerationRunRequest, idempotencyKey string) (RetryGenerationRunResponse, bool, error) {
	if err := s.ValidateRetryGenerationRun(ctx, id, req, idempotencyKey); err != nil {
		return RetryGenerationRunResponse{}, false, err
	}
	payload, err := idempotencyPayload(id, req)
	if err != nil {
		return RetryGenerationRunResponse{}, false, ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	record, ok := s.idempotency["retry:"+id+":"+idempotencyKey]
	if !ok {
		s.idempotency["retry:"+id+":"+idempotencyKey] = idempotencyRecord{payload: payload, pending: true}
		return RetryGenerationRunResponse{}, false, nil
	}
	if record.payload != payload {
		return RetryGenerationRunResponse{}, false, ErrIdempotencyConflict
	}
	if record.pending {
		return RetryGenerationRunResponse{}, false, ErrConflict
	}
	result, ok := record.result.(RetryGenerationRunResponse)
	if !ok {
		return RetryGenerationRunResponse{}, false, ErrConflict
	}
	return result, true, nil
}

func (s *service) RetryGenerationRun(ctx context.Context, id string, req RetryGenerationRunRequest, workflowRunID string, idempotencyKey string) (RetryGenerationRunResponse, error) {
	if err := s.ValidateRetryGenerationRun(ctx, id, req, idempotencyKey); err != nil {
		return RetryGenerationRunResponse{}, err
	}
	if workflowRunID == "" {
		return RetryGenerationRunResponse{}, ErrWorkflowRunFailed
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	original, ok := s.runs[id]
	if !ok {
		return RetryGenerationRunResponse{}, ErrNotFound
	}
	payload, err := idempotencyPayload(id, req)
	if err != nil {
		return RetryGenerationRunResponse{}, ErrValidation
	}
	key := "retry:" + id + ":" + idempotencyKey
	if record, ok := s.idempotency[key]; ok {
		if record.payload != payload {
			return RetryGenerationRunResponse{}, ErrIdempotencyConflict
		}
		if !record.pending {
			result, ok := record.result.(RetryGenerationRunResponse)
			if !ok {
				return RetryGenerationRunResponse{}, ErrConflict
			}
			return result, nil
		}
	}

	s.nextRun++
	now := time.Now().UTC()
	runID := fmt.Sprintf("genrun-%d", s.nextRun)
	detail := GenerationRunDetailResponse{GenerationRunResponse: original.GenerationRunResponse}
	detail.ID = runID
	detail.WorkflowRunID = workflowRunID
	detail.Status = GenerationRunPending
	detail.TriggerType = "retry"
	detail.RetryOfGenerationRunID = original.ID
	detail.CreatedAt = now
	detail.UpdatedAt = now
	detail.ContentItems = nil
	s.runs[runID] = detail
	result := RetryGenerationRunResponse{NewGenerationRunID: runID, WorkflowRunID: workflowRunID, OperationLogID: fmt.Sprintf("operation_log-%d", s.nextRun)}
	s.idempotency[key] = idempotencyRecord{payload: payload, result: result}
	return result, nil
}

func (s *service) ReconcileWorkflowResult(ctx context.Context, workflowRunID string) error {
	if workflowRunID == "" {
		return ErrValidation
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for runID, run := range s.runs {
		if run.WorkflowRunID != workflowRunID {
			continue
		}
		now := time.Now().UTC()
		run.Status = GenerationRunSucceeded
		run.UpdatedAt = now
		for index, item := range run.ContentItems {
			item.Status = ContentItemPendingReview
			item.UpdatedAt = now
			run.ContentItems[index] = item
			if detail, ok := s.items[item.ID]; ok {
				detail.Status = ContentItemPendingReview
				detail.UpdatedAt = now
				s.items[item.ID] = detail
			}
		}
		s.runs[runID] = run
		return nil
	}
	return ErrNotFound
}

func (s *service) ListContentItems(ctx context.Context, projectID string, req ListContentItemsRequest) (PagedContentItemsResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	items := make([]ContentItemResponse, 0)
	for _, item := range s.items {
		if item.ProjectID == projectID && (req.Status == "" || item.Status == req.Status) {
			items = append(items, item.ContentItemResponse)
		}
	}
	return PagedContentItemsResponse{Items: items, Pagination: pagination(len(items), req.PaginationRequest)}, nil
}

func (s *service) GetContentItem(ctx context.Context, id string) (ContentItemDetailResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return ContentItemDetailResponse{}, ErrNotFound
	}
	return item, nil
}

func idempotencyPayload(parts ...any) (string, error) {
	data, err := json.Marshal(parts)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func pagination(total int, req content.PaginationRequest) content.PaginationResponse {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	return content.PaginationResponse{Page: page, PageSize: pageSize, Total: total, HasNext: page*pageSize < total}
}
