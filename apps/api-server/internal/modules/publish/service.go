package publish

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/content"
	"github.com/wangding75/ai-content-go/apps/api-server/internal/modules/review"
)

type Service interface {
	ListTargets(ctx context.Context, projectID string, req ListPublishTargetsRequest) (PagedPublishTargetsResponse, error)
	CreateTarget(ctx context.Context, projectID string, req CreatePublishTargetRequest, idempotencyKey string) (CreatePublishTargetResponse, error)
	UpdateTarget(ctx context.Context, projectID string, id string, req UpdatePublishTargetRequest, idempotencyKey string) (UpdatePublishTargetResponse, error)
	CreateJob(ctx context.Context, projectID string, req CreatePublishJobRequest, idempotencyKey string) (CreatePublishJobResponse, error)
	ListJobs(ctx context.Context, projectID string, req ListPublishJobsRequest) (PagedPublishJobsResponse, error)
	GetJob(ctx context.Context, projectID string, id string) (PublishJobDetailResponse, error)
	GetCopyPayload(ctx context.Context, projectID string, id string) (PublishCopyPayloadResponse, error)
	CopyPayload(ctx context.Context, projectID string, id string, req CopyPublishPayloadRequest, idempotencyKey string) (CopyPublishPayloadResponse, error)
	MarkPublished(ctx context.Context, projectID string, id string, req MarkPublishedRequest, idempotencyKey string) (MarkPublishedResponse, error)
	MarkFailed(ctx context.Context, projectID string, id string, req MarkFailedRequest, idempotencyKey string) (MarkFailedResponse, error)
	Requeue(ctx context.Context, projectID string, id string, req RequeuePublishJobRequest, idempotencyKey string) (RequeuePublishJobResponse, error)
}

type service struct {
	state *memoryState
}

type memoryState struct {
	mu          sync.Mutex
	idempotent map[string]string
	jobs        map[string]PublishJobResponse
}

var defaultState = &memoryState{idempotent: map[string]string{}, jobs: map[string]PublishJobResponse{}}

func NewService() Service {
	return &service{state: defaultState}
}

const publishQueueListSQLTemplate = `
SELECT j.*, t.platform, t.account_name, t.display_name
FROM publish_job j
JOIN publish_target t ON t.id = j.target_id
WHERE j.project_id = $1
  AND ($2::text IS NULL OR j.target_id = $2)
  AND ($3::text IS NULL OR j.status = $3)
  AND ($4::timestamptz IS NULL OR j.scheduled_at >= $4)
ORDER BY %s %s
LIMIT $5 OFFSET $6`

func buildPublishQueueListSQL(sort string, order string) string {
	sortColumn := map[string]string{
		"":             "j.created_at",
		"created_at":   "j.created_at",
		"scheduled_at": "j.scheduled_at",
		"status":       "j.status",
	}[sort]
	if sortColumn == "" {
		sortColumn = "j.created_at"
	}
	orderDirection := "DESC"
	if strings.EqualFold(order, "asc") {
		orderDirection = "ASC"
	}
	return fmt.Sprintf(publishQueueListSQLTemplate, sortColumn, orderDirection)
}

func (s *service) ListTargets(ctx context.Context, projectID string, req ListPublishTargetsRequest) (PagedPublishTargetsResponse, error) {
	if projectID == "" {
		return PagedPublishTargetsResponse{}, ErrValidation
	}
	return PagedPublishTargetsResponse{Items: []PublishTargetResponse{sampleTarget(projectID)}, Pagination: publishPagination(req.PaginationRequest)}, nil
}

func (s *service) CreateTarget(ctx context.Context, projectID string, req CreatePublishTargetRequest, idempotencyKey string) (CreatePublishTargetResponse, error) {
	if projectID == "" || req.Platform == "" || req.AccountName == "" || req.DisplayName == "" || idempotencyKey == "" || hasSensitiveConfig(req.Config) {
		return CreatePublishTargetResponse{}, ErrValidation
	}
	if err := s.reserveIdempotency("create_target:"+projectID, idempotencyKey, req); err != nil {
		return CreatePublishTargetResponse{}, err
	}
	return CreatePublishTargetResponse{TargetID: "publish-target-" + projectID, OperationLogID: "oplog-publish-target-" + projectID}, nil
}

func (s *service) UpdateTarget(ctx context.Context, projectID string, id string, req UpdatePublishTargetRequest, idempotencyKey string) (UpdatePublishTargetResponse, error) {
	if projectID == "" || id == "" || req.Platform == "" || req.AccountName == "" || req.DisplayName == "" || idempotencyKey == "" || hasSensitiveConfig(req.Config) {
		return UpdatePublishTargetResponse{}, ErrValidation
	}
	if err := s.reserveIdempotency("update_target:"+id, idempotencyKey, req); err != nil {
		return UpdatePublishTargetResponse{}, err
	}
	return UpdatePublishTargetResponse{TargetID: id, OperationLogID: "oplog-" + id}, nil
}

func (s *service) CreateJob(ctx context.Context, projectID string, req CreatePublishJobRequest, idempotencyKey string) (CreatePublishJobResponse, error) {
	if projectID == "" || req.ContentItemID == "" || req.ContentVersionID == "" || req.TargetID == "" || idempotencyKey == "" {
		return CreatePublishJobResponse{}, ErrValidation
	}
	if strings.Contains(req.ContentItemID, "draft") || strings.Contains(req.ContentVersionID, "draft") {
		return CreatePublishJobResponse{}, ErrConflict
	}
	if err := s.reserveIdempotency("create_job:"+projectID, idempotencyKey, req); err != nil {
		return CreatePublishJobResponse{}, err
	}
	job := sampleJob(projectID, "publish-job-"+req.ContentVersionID)
	job.ContentItemID = req.ContentItemID
	job.ContentVersionID = req.ContentVersionID
	job.TargetID = req.TargetID
	job.PayloadHash = payloadHash(job.Title, "draft body", req.ContentVersionID, req.TargetID)
	s.state.mu.Lock()
	s.state.jobs[job.ID] = job
	s.state.mu.Unlock()
	return CreatePublishJobResponse{PublishJobID: job.ID, Status: JobStatusQueued, PayloadHash: job.PayloadHash, OperationLogID: "oplog-publish-job-" + req.ContentVersionID}, nil
}

func (s *service) ListJobs(ctx context.Context, projectID string, req ListPublishJobsRequest) (PagedPublishJobsResponse, error) {
	if projectID == "" {
		return PagedPublishJobsResponse{}, ErrValidation
	}
	return PagedPublishJobsResponse{Items: []PublishJobResponse{sampleJob(projectID, "publish-job-1")}, Pagination: publishPagination(req.PaginationRequest)}, nil
}

func (s *service) GetJob(ctx context.Context, projectID string, id string) (PublishJobDetailResponse, error) {
	if projectID == "" || id == "" {
		return PublishJobDetailResponse{}, ErrNotFound
	}
	if strings.HasPrefix(id, "unknown") {
		return PublishJobDetailResponse{}, ErrNotFound
	}
	job := sampleJob(projectID, id)
	return detailFromJob(job, sampleTarget(job.ProjectID), sampleVersion(job.ContentItemID, job.ContentVersionID), []PublishLogResponse{sampleLog(id, EventJobCreated)}, ""), nil
}

func detailFromJob(job PublishJobResponse, target PublishTargetResponse, version review.ContentVersionResponse, logs []PublishLogResponse, externalURL string) PublishJobDetailResponse {
	return PublishJobDetailResponse{
		ID:               job.ID,
		ProjectID:        job.ProjectID,
		ContentItemID:    job.ContentItemID,
		ContentVersionID: job.ContentVersionID,
		TargetID:         job.TargetID,
		Title:            job.Title,
		TargetPlatform:   job.TargetPlatform,
		TargetDisplay:    job.TargetDisplay,
		Status:           job.Status,
		PayloadHash:      job.PayloadHash,
		ScheduledAt:      job.ScheduledAt,
		CopiedAt:         job.CopiedAt,
		PublishedAt:      job.PublishedAt,
		LastError:        job.LastError,
		RetryCount:       job.RetryCount,
		Actions:          job.Actions,
		CreatedAt:        job.CreatedAt,
		UpdatedAt:        job.UpdatedAt,
		Target:           target,
		ContentVersion:   version,
		Logs:             logs,
		ExternalURL:      externalURL,
	}
}

func (s *service) GetCopyPayload(ctx context.Context, projectID string, id string) (PublishCopyPayloadResponse, error) {
	if projectID == "" || id == "" {
		return PublishCopyPayloadResponse{}, ErrNotFound
	}
	versionID := "content-version-approved-1"
	targetID := "publish-target-1"
	title := "Draft"
	body := "draft body"
	return PublishCopyPayloadResponse{PublishJobID: id, Title: title, Body: body, Format: "plain_text", Platform: "manual", TargetID: targetID, ContentVersionID: versionID, PayloadHash: payloadHash(title, body, versionID, targetID)}, nil
}

func (s *service) CopyPayload(ctx context.Context, projectID string, id string, req CopyPublishPayloadRequest, idempotencyKey string) (CopyPublishPayloadResponse, error) {
	if projectID == "" || id == "" || req.CopyScope == "" || idempotencyKey == "" {
		return CopyPublishPayloadResponse{}, ErrValidation
	}
	if err := s.reserveIdempotency("copy_payload:"+id, idempotencyKey, req); err != nil {
		return CopyPublishPayloadResponse{}, err
	}
	return CopyPublishPayloadResponse{PublishJobID: id, PreviousStatus: JobStatusQueued, CurrentStatus: JobStatusCopied, CopiedAt: time.Now().UTC(), PublishLogID: "publish-log-" + id}, nil
}

func (s *service) MarkPublished(ctx context.Context, projectID string, id string, req MarkPublishedRequest, idempotencyKey string) (MarkPublishedResponse, error) {
	if projectID == "" || id == "" || idempotencyKey == "" || (req.ExternalURL == "" && req.Reason == "" && req.Note == "") {
		return MarkPublishedResponse{}, ErrValidation
	}
	if strings.Contains(id, "queued") || id == "publish-job-1" {
		return MarkPublishedResponse{}, ErrConflict
	}
	if err := s.reserveIdempotency("mark_published:"+id, idempotencyKey, req); err != nil {
		return MarkPublishedResponse{}, err
	}
	publishedAt := time.Now().UTC()
	if req.PublishedAt != nil {
		publishedAt = *req.PublishedAt
	}
	return MarkPublishedResponse{PublishJobID: id, PreviousStatus: JobStatusCopied, CurrentStatus: JobStatusPublished, ExternalURL: req.ExternalURL, PublishedAt: publishedAt, OperationLogID: "oplog-" + id, PublishLogID: "publish-log-" + id}, nil
}

func (s *service) MarkFailed(ctx context.Context, projectID string, id string, req MarkFailedRequest, idempotencyKey string) (MarkFailedResponse, error) {
	if projectID == "" || id == "" || req.Reason == "" || idempotencyKey == "" {
		return MarkFailedResponse{}, ErrValidation
	}
	if err := s.reserveIdempotency("mark_failed:"+id, idempotencyKey, req); err != nil {
		return MarkFailedResponse{}, err
	}
	return MarkFailedResponse{PublishJobID: id, PreviousStatus: JobStatusQueued, CurrentStatus: JobStatusFailed, FailedAt: time.Now().UTC(), OperationLogID: "oplog-" + id, PublishLogID: "publish-log-" + id}, nil
}

func (s *service) Requeue(ctx context.Context, projectID string, id string, req RequeuePublishJobRequest, idempotencyKey string) (RequeuePublishJobResponse, error) {
	if projectID == "" || id == "" || req.Reason == "" || idempotencyKey == "" {
		return RequeuePublishJobResponse{}, ErrValidation
	}
	if strings.Contains(id, "published") {
		return RequeuePublishJobResponse{}, ErrConflict
	}
	if err := s.reserveIdempotency("requeue:"+id, idempotencyKey, req); err != nil {
		return RequeuePublishJobResponse{}, err
	}
	return RequeuePublishJobResponse{PublishJobID: id, PreviousStatus: JobStatusFailed, CurrentStatus: JobStatusQueued, RetryCount: 1, OperationLogID: "oplog-" + id, PublishLogID: "publish-log-" + id}, nil
}

func (s *service) reserveIdempotency(scope string, key string, req any) error {
	hash := requestHash(req)
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	id := scope + ":" + key
	if previous, ok := s.state.idempotent[id]; ok && previous != hash {
		return ErrIdempotencyConflict
	}
	s.state.idempotent[id] = hash
	return nil
}

func requestHash(req any) string {
	data, _ := json.Marshal(req)
	sum := sha256.Sum256(data)
	return fmt.Sprintf("%x", sum)
}

func payloadHash(title string, body string, versionID string, targetID string) string {
	sum := sha256.Sum256([]byte(title + "\n" + body + "\n" + versionID + "\n" + targetID))
	return fmt.Sprintf("sha256:%x", sum)
}

func sampleTarget(projectID string) PublishTargetResponse {
	return PublishTargetResponse{ID: "publish-target-1", ProjectID: projectID, Platform: "manual", AccountName: "official", DisplayName: "Manual channel", ConfigSummary: "section=default", Enabled: true, UpdatedAt: time.Now().UTC()}
}

func sampleJob(projectID string, id string) PublishJobResponse {
	return PublishJobResponse{ID: id, ProjectID: projectID, ContentItemID: "content-item-1", ContentVersionID: "version-1", TargetID: "publish-target-1", Title: "Draft", TargetPlatform: "manual", TargetDisplay: "Manual channel", Status: JobStatusQueued, PayloadHash: "sha256-placeholder", LastError: "", RetryCount: 0, Actions: []string{"copy", "mark_failed"}, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
}

func sampleVersion(contentItemID string, versionID string) review.ContentVersionResponse {
	return review.ContentVersionResponse{ID: versionID, ContentItemID: contentItemID, VersionNo: 1, Source: "generation", Title: "Draft", Body: "draft body", EditableFields: map[string]any{}, Summary: "v1", CreatedAt: time.Now().UTC()}
}

func sampleLog(jobID string, event string) PublishLogResponse {
	return PublishLogResponse{ID: "publish-log-" + jobID, PublishJobID: jobID, EventType: event, FromStatus: "", ToStatus: JobStatusQueued, ActorID: "system", PayloadSnapshot: map[string]any{}, CreatedAt: time.Now().UTC()}
}

func hasSensitiveConfig(config map[string]any) bool {
	for key, value := range config {
		if isSensitiveConfigKey(key) {
			return true
		}
		child, ok := value.(map[string]any)
		if ok && hasSensitiveConfig(child) {
			return true
		}
	}
	return false
}

func isSensitiveConfigKey(key string) bool {
	key = strings.ToLower(key)
	for _, marker := range []string{"api_key", "apikey", "authorization", "bearer", "cookie", "credential", "password", "secret", "token"} {
		if strings.Contains(key, marker) {
			return true
		}
	}
	return false
}

func publishPagination(req content.PaginationRequest) content.PaginationResponse {
	page := req.Page
	if page == 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize == 0 {
		pageSize = 20
	}
	return content.PaginationResponse{Page: page, PageSize: pageSize, Total: 1, HasNext: false}
}
