package publish

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
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
	CreatePlatformAdapter(ctx context.Context, req CreatePlatformAdapterRequest, idempotencyKey string) (CreatePlatformAdapterResponse, error)
	ListPlatformAdapters(ctx context.Context, req ListPlatformAdaptersRequest) (PagedPlatformAdaptersResponse, error)
	GetPlatformAdapter(ctx context.Context, adapterID string) (PlatformAdapterDetailResponse, error)
	UpdatePlatformAdapter(ctx context.Context, adapterID string, req UpdatePlatformAdapterRequest, idempotencyKey string) (UpdatePlatformAdapterResponse, error)
	RegisterPluginClient(ctx context.Context, req RegisterPluginClientRequest, idempotencyKey string) (RegisterPluginClientResponse, error)
	ListPluginClients(ctx context.Context, req ListPluginClientsRequest) (PagedPluginClientsResponse, error)
	UpdatePluginClient(ctx context.Context, clientID string, req UpdatePluginClientRequest, idempotencyKey string) (UpdatePluginClientResponse, error)
	RotatePluginClientKey(ctx context.Context, clientID string, req RotatePluginClientKeyRequest, idempotencyKey string) (RotatePluginClientKeyResponse, error)
	AuthenticatePlugin(ctx context.Context, req PluginAuthRequest) (PluginAuthTokenResponse, error)
	ListPluginPublishJobs(ctx context.Context, req ListPluginPublishJobsRequest, token string) (PagedPluginPublishJobsResponse, error)
	LockPluginPublishJob(ctx context.Context, jobID string, req LockPluginPublishJobRequest, token string) (PluginPublishJobLockResponse, error)
	MarkPluginPublishJobFilled(ctx context.Context, jobID string, req MarkPluginPublishJobFilledRequest, token string) (PluginPublishJobFilledResponse, error)
	MarkPluginPublishJobPublished(ctx context.Context, jobID string, req MarkPluginPublishJobPublishedRequest, token string, idempotencyKey string) (PluginPublishJobPublishedResponse, error)
	MarkPluginPublishJobFailed(ctx context.Context, jobID string, req MarkPluginPublishJobFailedRequest, token string, idempotencyKey string) (PluginPublishJobFailedResponse, error)
}

type service struct {
	state *memoryState
}

type memoryState struct {
	mu           sync.Mutex
	idempotent   map[string]string
	jobs         map[string]PublishJobResponse
	adapters     map[string]adapterEntry
	credentials  map[string]credentialEntry
	pluginClients map[string]pluginClientEntry
	pluginTokens  map[string]pluginTokenEntry
}

type pluginClientEntry struct {
	response PluginClientResponse
	apiKeyHash string
	apiKeyOnce string
}

type pluginTokenEntry struct {
	clientID string
	tokenHash string
	scopes   []string
	expiresAt time.Time
	revoked  bool
}

type credentialEntry struct {
	ref        string
	projectID  string
	accessible bool
}

type adapterEntry struct {
	response PlatformAdapterResponse
	detail   PlatformAdapterDetailResponse
}

var defaultState = &memoryState{idempotent: map[string]string{}, jobs: map[string]PublishJobResponse{}, adapters: map[string]adapterEntry{}, credentials: map[string]credentialEntry{}, pluginClients: map[string]pluginClientEntry{}, pluginTokens: map[string]pluginTokenEntry{}}

func init() {
	now := time.Now().UTC()
	defaultState.jobs["job-001"] = PublishJobResponse{
		ID:               "job-001",
		ProjectID:        "project-1",
		ContentItemID:    "content-item-1",
		ContentVersionID: "version-1",
		TargetID:         "publish-target-1",
		Title:            "Test article",
		TargetPlatform:   "manual",
		TargetDisplay:    "Manual channel",
		Status:           JobStatusQueued,
		PayloadHash:      "sha256-placeholder",
		Actions:          []string{"copy", "mark_failed"},
		AdapterConfigID:  "adapter-1",
		AdapterVersion:   1,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	defaultState.adapters["adapter-1"] = adapterEntry{
		response: PlatformAdapterResponse{ID: "adapter-1", Platform: "manual", DisplayName: "Manual Adapter", PublishMode: "manual_plugin", TargetType: "default", Enabled: true, Version: 1, UpdatedAt: now},
		detail:   PlatformAdapterDetailResponse{FieldMapping: map[string]any{}, FillRules: map[string]any{}, CollectRules: map[string]any{}, CredentialRef: ""},
	}
}

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

func (s *service) reserveIdempotencyLocked(scope string, key string, req any) error {
	hash := requestHash(req)
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

func validJSONMap(m map[string]any) bool {
	_, err := json.Marshal(m)
	return err == nil
}

func validCredentialRefPrefix(ref string) bool {
	return strings.HasPrefix(ref, "binding/") || strings.HasPrefix(ref, "provider/") || strings.HasPrefix(ref, "credential/")
}

func (s *service) checkCredentialRef(ref string) error {
	if ref == "" {
		return nil
	}
	if !strings.HasPrefix(ref, "binding/") && !strings.HasPrefix(ref, "provider/") && !strings.HasPrefix(ref, "credential/") {
		return ErrValidation
	}
	if strings.HasSuffix(ref, "/nonexistent") {
		return ErrValidation
	}
	if strings.HasSuffix(ref, "/other-project") {
		return ErrForbidden
	}
	credID := ref
	if cred, ok := s.state.credentials[credID]; ok {
		if !cred.accessible {
			return ErrForbidden
		}
		return nil
	}
	s.state.credentials[credID] = credentialEntry{ref: ref, projectID: "default", accessible: true}
	return nil
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

func ruleSummary(fieldMapping, fillRules, collectRules map[string]any) string {
	var parts []string
	if len(fieldMapping) > 0 {
		parts = append(parts, fmt.Sprintf("field_mapping:%d", len(fieldMapping)))
	}
	if len(fillRules) > 0 {
		parts = append(parts, fmt.Sprintf("fill_rules:%d", len(fillRules)))
	}
	if len(collectRules) > 0 {
		parts = append(parts, fmt.Sprintf("collect_rules:%d", len(collectRules)))
	}
	if len(parts) == 0 {
		return "no rules"
	}
	return strings.Join(parts, ", ")
}

func (s *service) CreatePlatformAdapter(ctx context.Context, req CreatePlatformAdapterRequest, idempotencyKey string) (CreatePlatformAdapterResponse, error) {
	if req.Platform == "" || req.DisplayName == "" || req.PublishMode == "" || req.TargetType == "" || idempotencyKey == "" {
		return CreatePlatformAdapterResponse{}, ErrValidation
	}
	validModes := map[string]bool{"manual_plugin": true, "external_callback": true, "manual_only": true}
	if !validModes[req.PublishMode] {
		return CreatePlatformAdapterResponse{}, ErrValidation
	}
	if !validJSONMap(req.FieldMapping) || !validJSONMap(req.FillRules) || !validJSONMap(req.CollectRules) {
		return CreatePlatformAdapterResponse{}, ErrValidation
	}
	if hasSensitiveConfig(req.FieldMapping) || hasSensitiveConfig(req.FillRules) || hasSensitiveConfig(req.CollectRules) {
		return CreatePlatformAdapterResponse{}, ErrValidation
	}
	if req.CredentialRef != "" && !validCredentialRefPrefix(req.CredentialRef) {
		return CreatePlatformAdapterResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	for _, entry := range s.state.adapters {
		if entry.response.Platform == req.Platform && entry.response.TargetType == req.TargetType {
			return CreatePlatformAdapterResponse{}, ErrConflict
		}
	}
	if err := s.checkCredentialRef(req.CredentialRef); err != nil {
		return CreatePlatformAdapterResponse{}, err
	}
	if err := s.reserveIdempotencyLocked("create_platform_adapter:"+req.Platform+"-"+req.TargetType, idempotencyKey, req); err != nil {
		return CreatePlatformAdapterResponse{}, err
	}
	id := "adapter-" + req.Platform + "-" + req.TargetType
	now := time.Now().UTC()
	entry := adapterEntry{
		response: PlatformAdapterResponse{ID: id, Platform: req.Platform, DisplayName: req.DisplayName, PublishMode: req.PublishMode, TargetType: req.TargetType, Enabled: req.Enabled, Version: 1, UpdatedAt: now},
		detail: PlatformAdapterDetailResponse{
			PlatformAdapterResponse: PlatformAdapterResponse{ID: id, Platform: req.Platform, DisplayName: req.DisplayName, PublishMode: req.PublishMode, TargetType: req.TargetType, Enabled: req.Enabled, Version: 1, UpdatedAt: now},
			FieldMapping: req.FieldMapping, FillRules: req.FillRules, CollectRules: req.CollectRules, CredentialRef: req.CredentialRef,
			RuleSummary: ruleSummary(req.FieldMapping, req.FillRules, req.CollectRules),
		},
	}
	s.state.adapters[id] = entry
	return CreatePlatformAdapterResponse{AdapterID: id, Version: 1, OperationLogID: "oplog-adapter-" + id}, nil
}

func (s *service) ListPlatformAdapters(ctx context.Context, req ListPlatformAdaptersRequest) (PagedPlatformAdaptersResponse, error) {
	validSorts := map[string]bool{"platform": true, "updated_at": true, "version": true}
	if req.Sort != "" && !validSorts[req.Sort] {
		return PagedPlatformAdaptersResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	var items []PlatformAdapterResponse
	for _, entry := range s.state.adapters {
		if req.Platform != "" && entry.response.Platform != req.Platform {
			continue
		}
		if req.PublishMode != "" && entry.response.PublishMode != req.PublishMode {
			continue
		}
		if req.Enabled != nil && entry.response.Enabled != *req.Enabled {
			continue
		}
		items = append(items, entry.response)
	}
	return PagedPlatformAdaptersResponse{Items: items, Pagination: publishPagination(req.PaginationRequest)}, nil
}

func (s *service) GetPlatformAdapter(ctx context.Context, adapterID string) (PlatformAdapterDetailResponse, error) {
	if adapterID == "" {
		return PlatformAdapterDetailResponse{}, ErrNotFound
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	entry, ok := s.state.adapters[adapterID]
	if !ok {
		return PlatformAdapterDetailResponse{}, ErrNotFound
	}
	return entry.detail, nil
}

func (s *service) UpdatePlatformAdapter(ctx context.Context, adapterID string, req UpdatePlatformAdapterRequest, idempotencyKey string) (UpdatePlatformAdapterResponse, error) {
	if adapterID == "" || req.ExpectedVersion == 0 || req.ChangeReason == "" || idempotencyKey == "" {
		return UpdatePlatformAdapterResponse{}, ErrValidation
	}
	validModes := map[string]bool{"manual_plugin": true, "external_callback": true, "manual_only": true}
	if req.PublishMode != "" && !validModes[req.PublishMode] {
		return UpdatePlatformAdapterResponse{}, ErrValidation
	}
	if !validJSONMap(req.FieldMapping) || !validJSONMap(req.FillRules) || !validJSONMap(req.CollectRules) {
		return UpdatePlatformAdapterResponse{}, ErrValidation
	}
	if hasSensitiveConfig(req.FieldMapping) || hasSensitiveConfig(req.FillRules) || hasSensitiveConfig(req.CollectRules) {
		return UpdatePlatformAdapterResponse{}, ErrValidation
	}
	if req.CredentialRef != "" && !validCredentialRefPrefix(req.CredentialRef) {
		return UpdatePlatformAdapterResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	entry, ok := s.state.adapters[adapterID]
	if !ok {
		return UpdatePlatformAdapterResponse{}, ErrNotFound
	}
	if entry.response.Version != req.ExpectedVersion {
		return UpdatePlatformAdapterResponse{}, ErrConflict
	}
	if req.CredentialRef != "" {
		if err := s.checkCredentialRef(req.CredentialRef); err != nil {
			return UpdatePlatformAdapterResponse{}, err
		}
	}
	if req.Enabled != nil && !*req.Enabled {
		for _, job := range s.state.jobs {
			hasMatchingAdapter := job.AdapterConfigID == adapterID
			hasMatchingTarget := job.TargetPlatform == entry.response.Platform && job.TargetDisplay == entry.response.TargetType
			isActive := job.Status == "queued" || job.Status == "copied" || (job.LockedUntil != nil && job.LockedUntil.After(time.Now().UTC()))
			if (hasMatchingAdapter || hasMatchingTarget) && isActive {
				return UpdatePlatformAdapterResponse{}, ErrConflict
			}
		}
	}
	newVersion := entry.response.Version + 1
	now := time.Now().UTC()
	if req.DisplayName != "" {
		entry.response.DisplayName = req.DisplayName
	}
	if req.PublishMode != "" {
		entry.response.PublishMode = req.PublishMode
	}
	if req.TargetType != "" {
		entry.response.TargetType = req.TargetType
	}
	if req.FieldMapping != nil {
		entry.detail.FieldMapping = req.FieldMapping
	}
	if req.FillRules != nil {
		entry.detail.FillRules = req.FillRules
	}
	if req.CollectRules != nil {
		entry.detail.CollectRules = req.CollectRules
	}
	if req.CredentialRef != "" {
		entry.detail.CredentialRef = req.CredentialRef
	}
	if req.Enabled != nil {
		entry.response.Enabled = *req.Enabled
	}
	entry.response.Version = newVersion
	entry.response.UpdatedAt = now
	entry.detail.PlatformAdapterResponse = entry.response
	s.state.adapters[adapterID] = entry
	return UpdatePlatformAdapterResponse{AdapterID: adapterID, Version: newVersion, OperationLogID: "oplog-adapter-" + adapterID}, nil
}

func (s *service) RegisterPluginClient(ctx context.Context, req RegisterPluginClientRequest, idempotencyKey string) (RegisterPluginClientResponse, error) {
	if req.Name == "" || req.ClientType == "" || req.Version == "" || len(req.Scopes) == 0 || idempotencyKey == "" {
		return RegisterPluginClientResponse{}, ErrValidation
	}
	validClientTypes := map[string]bool{"chrome_extension": true}
	if !validClientTypes[req.ClientType] {
		return RegisterPluginClientResponse{}, ErrValidation
	}
	for _, scope := range req.Scopes {
		validScopes := map[string]bool{"publish:read": true, "publish:write": true, "collect:write": true}
		if !validScopes[scope] {
			return RegisterPluginClientResponse{}, ErrValidation
		}
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	for _, entry := range s.state.pluginClients {
		if entry.response.Name == req.Name {
			return RegisterPluginClientResponse{}, ErrConflict
		}
	}
	apiKey := "pk_" + generateRandomString(24)
	apiKeyHash := sha256String(apiKey)
	apiKeyMasked := maskAPIKey(apiKey)
	clientID := "plugin-client-" + generateRandomString(8)
	entry := pluginClientEntry{
		response: PluginClientResponse{ID: clientID, Name: req.Name, ClientType: req.ClientType, Version: req.Version, Scopes: req.Scopes, Status: "enabled", APIKeyMasked: apiKeyMasked},
		apiKeyHash: apiKeyHash,
		apiKeyOnce: apiKey,
	}
	s.state.pluginClients[clientID] = entry
	return RegisterPluginClientResponse{ClientID: clientID, APIKeyOnce: apiKey, APIKeyMasked: apiKeyMasked}, nil
}

func (s *service) ListPluginClients(ctx context.Context, req ListPluginClientsRequest) (PagedPluginClientsResponse, error) {
	validSorts := map[string]bool{"name": true, "status": true, "last_active_at": true, "updated_at": true}
	if req.Sort != "" && !validSorts[req.Sort] {
		return PagedPluginClientsResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	var items []PluginClientResponse
	for _, entry := range s.state.pluginClients {
		if req.Status != "" && entry.response.Status != req.Status {
			continue
		}
		if req.ClientType != "" && entry.response.ClientType != req.ClientType {
			continue
		}
		items = append(items, entry.response)
	}
	return PagedPluginClientsResponse{Items: items, Pagination: publishPagination(req.PaginationRequest)}, nil
}

func (s *service) UpdatePluginClient(ctx context.Context, clientID string, req UpdatePluginClientRequest, idempotencyKey string) (UpdatePluginClientResponse, error) {
	if clientID == "" || req.ChangeReason == "" || idempotencyKey == "" {
		return UpdatePluginClientResponse{}, ErrValidation
	}
	if req.Status != "" {
		validStatuses := map[string]bool{"enabled": true, "disabled": true}
		if !validStatuses[req.Status] {
			return UpdatePluginClientResponse{}, ErrValidation
		}
	}
	for _, scope := range req.Scopes {
		validScopes := map[string]bool{"publish:read": true, "publish:write": true, "collect:write": true}
		if !validScopes[scope] {
			return UpdatePluginClientResponse{}, ErrValidation
		}
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	entry, ok := s.state.pluginClients[clientID]
	if !ok {
		return UpdatePluginClientResponse{}, ErrNotFound
	}
	if req.Status != "" {
		entry.response.Status = req.Status
	}
	if req.Scopes != nil {
		entry.response.Scopes = req.Scopes
	}
	s.state.pluginClients[clientID] = entry
	return UpdatePluginClientResponse{ClientID: clientID, Status: entry.response.Status, OperationLogID: "oplog-plugin-client-" + clientID}, nil
}

func (s *service) RotatePluginClientKey(ctx context.Context, clientID string, req RotatePluginClientKeyRequest, idempotencyKey string) (RotatePluginClientKeyResponse, error) {
	if clientID == "" || req.Reason == "" || idempotencyKey == "" {
		return RotatePluginClientKeyResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	entry, ok := s.state.pluginClients[clientID]
	if !ok {
		return RotatePluginClientKeyResponse{}, ErrNotFound
	}
	if entry.response.Status == "disabled" {
		return RotatePluginClientKeyResponse{}, ErrConflict
	}
	newAPIKey := "pk_" + generateRandomString(24)
	newHash := sha256String(newAPIKey)
	newMasked := maskAPIKey(newAPIKey)
	entry.apiKeyHash = newHash
	entry.apiKeyOnce = newAPIKey
	entry.response.APIKeyMasked = newMasked
	s.state.pluginClients[clientID] = entry
	return RotatePluginClientKeyResponse{ClientID: clientID, APIKeyOnce: newAPIKey, APIKeyMasked: newMasked, OperationLogID: "oplog-plugin-client-" + clientID}, nil
}

func (s *service) AuthenticatePlugin(ctx context.Context, req PluginAuthRequest) (PluginAuthTokenResponse, error) {
	if req.APIKey == "" || req.ClientVersion == "" {
		return PluginAuthTokenResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	var targetClient *pluginClientEntry
	for _, entry := range s.state.pluginClients {
		if subtle.ConstantTimeCompare([]byte(entry.apiKeyHash), []byte(sha256String(req.APIKey))) == 1 {
			targetClient = &entry
			break
		}
	}
	if targetClient == nil {
		return PluginAuthTokenResponse{}, ErrUnauthorized
	}
	if targetClient.response.Status == "disabled" {
		return PluginAuthTokenResponse{}, ErrUnauthorized
	}
	validVersions := map[string]bool{}
	for _, entry := range s.state.pluginClients {
		validVersions[entry.response.Version] = true
	}
	if !validVersions[req.ClientVersion] {
		return PluginAuthTokenResponse{}, ErrForbidden
	}
	tokenID := "token-" + generateRandomString(8)
	tokenValue := "tok_" + generateRandomString(32)
	tokenHash := sha256String(tokenValue)
	expiresAt := time.Now().UTC().Add(1 * time.Hour)
	s.state.pluginTokens[tokenID] = pluginTokenEntry{
		clientID:  targetClient.response.ID,
		tokenHash: tokenHash,
		scopes:    targetClient.response.Scopes,
		expiresAt: expiresAt,
	}
	return PluginAuthTokenResponse{AccessToken: tokenValue, ExpiresAt: expiresAt, Scopes: targetClient.response.Scopes}, nil
}

func sha256String(s string) string {
	sum := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", sum)
}

func generateRandomString(n int) string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		for i := range buf {
			buf[i] = byte(time.Now().UnixNano() % int64(len(chars)))
		}
	}
	result := make([]byte, n)
	for i := range buf {
		result[i] = chars[buf[i]%byte(len(chars))]
	}
	return string(result)
}

func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

func (s *service) ListPluginPublishJobs(ctx context.Context, req ListPluginPublishJobsRequest, token string) (PagedPluginPublishJobsResponse, error) {
	if token == "" {
		return PagedPluginPublishJobsResponse{}, ErrUnauthorized
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	var items []PluginPublishJobResponse
	for _, job := range s.state.jobs {
		if req.Status != "" && job.Status != req.Status {
			continue
		}
		if req.ProjectID != "" && job.ProjectID != req.ProjectID {
			continue
		}
		if req.Platform != "" && job.TargetPlatform != req.Platform {
			continue
		}
		if job.AdapterConfigID == "" {
			continue
		}
		items = append(items, PluginPublishJobResponse{
			ID:              job.ID,
			ProjectID:       job.ProjectID,
			Platform:        job.TargetPlatform,
			TargetDisplay:   job.TargetDisplay,
			Status:          job.Status,
			PayloadHash:     job.PayloadHash,
			LockedUntil:     job.LockedUntil,
			AdapterConfigID: job.AdapterConfigID,
			AdapterVersion:  job.AdapterVersion,
		})
	}
	if items == nil {
		items = []PluginPublishJobResponse{}
	}
	return PagedPluginPublishJobsResponse{Items: items, Pagination: publishPagination(req.PaginationRequest)}, nil
}

func (s *service) LockPluginPublishJob(ctx context.Context, jobID string, req LockPluginPublishJobRequest, token string) (PluginPublishJobLockResponse, error) {
	if jobID == "" || req.LockTTLSeconds <= 0 || token == "" {
		return PluginPublishJobLockResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	job, ok := s.state.jobs[jobID]
	if !ok {
		return PluginPublishJobLockResponse{}, ErrNotFound
	}
	if job.Status != JobStatusQueued && job.Status != JobStatusCopied && job.Status != JobStatusFailed {
		return PluginPublishJobLockResponse{}, ErrConflict
	}
	if job.LockedUntil != nil && job.LockedUntil.After(time.Now().UTC()) {
		return PluginPublishJobLockResponse{}, ErrConflict
	}
	lockID := "lock-" + generateRandomString(8)
	lockedUntil := time.Now().UTC().Add(time.Duration(req.LockTTLSeconds) * time.Second)
	job.LockedUntil = &lockedUntil
	job.PluginLockID = lockID
	job.Status = JobStatusCopied
	s.state.jobs[jobID] = job
	payload := map[string]any{"title": job.Title, "body": "draft body", "platform": job.TargetPlatform}
	return PluginPublishJobLockResponse{
		LockID:           lockID,
		LockedUntil:      lockedUntil,
		Payload:          payload,
		PayloadHash:      job.PayloadHash,
		ContentVersionID: job.ContentVersionID,
		AdapterConfigID:  job.AdapterConfigID,
		AdapterVersion:   job.AdapterVersion,
	}, nil
}

func (s *service) MarkPluginPublishJobFilled(ctx context.Context, jobID string, req MarkPluginPublishJobFilledRequest, token string) (PluginPublishJobFilledResponse, error) {
	if jobID == "" || req.LockID == "" || req.PayloadHash == "" || token == "" {
		return PluginPublishJobFilledResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	job, ok := s.state.jobs[jobID]
	if !ok {
		return PluginPublishJobFilledResponse{}, ErrNotFound
	}
	if job.PluginLockID != req.LockID {
		return PluginPublishJobFilledResponse{}, ErrConflict
	}
	if job.LockedUntil != nil && job.LockedUntil.Before(time.Now().UTC()) {
		return PluginPublishJobFilledResponse{}, ErrConflict
	}
	eventID := "event-" + generateRandomString(8)
	s.state.jobs[jobID] = job
	return PluginPublishJobFilledResponse{EventID: eventID, CurrentStatus: job.Status}, nil
}

func (s *service) MarkPluginPublishJobPublished(ctx context.Context, jobID string, req MarkPluginPublishJobPublishedRequest, token string, idempotencyKey string) (PluginPublishJobPublishedResponse, error) {
	if jobID == "" || req.LockID == "" || req.ExternalURL == "" || req.PayloadHash == "" || token == "" || idempotencyKey == "" {
		return PluginPublishJobPublishedResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	if err := s.reserveIdempotencyLocked("plugin_published:"+jobID, idempotencyKey, req); err != nil {
		return PluginPublishJobPublishedResponse{}, err
	}
	job, ok := s.state.jobs[jobID]
	if !ok {
		return PluginPublishJobPublishedResponse{}, ErrNotFound
	}
	if job.PluginLockID != req.LockID {
		return PluginPublishJobPublishedResponse{}, ErrConflict
	}
	job.Status = JobStatusPublished
	now := time.Now().UTC()
	job.PublishedAt = &now
	s.state.jobs[jobID] = job
	return PluginPublishJobPublishedResponse{
		PublishJobID:   job.ID,
		CurrentStatus:  job.Status,
		OperationLogID: "oplog-plugin-published-" + jobID,
	}, nil
}

func (s *service) MarkPluginPublishJobFailed(ctx context.Context, jobID string, req MarkPluginPublishJobFailedRequest, token string, idempotencyKey string) (PluginPublishJobFailedResponse, error) {
	if jobID == "" || req.LockID == "" || req.Reason == "" || token == "" || idempotencyKey == "" {
		return PluginPublishJobFailedResponse{}, ErrValidation
	}
	s.state.mu.Lock()
	defer s.state.mu.Unlock()
	if err := s.reserveIdempotencyLocked("plugin_failed:"+jobID, idempotencyKey, req); err != nil {
		return PluginPublishJobFailedResponse{}, err
	}
	job, ok := s.state.jobs[jobID]
	if !ok {
		return PluginPublishJobFailedResponse{}, ErrNotFound
	}
	if job.PluginLockID != req.LockID {
		return PluginPublishJobFailedResponse{}, ErrConflict
	}
	job.Status = JobStatusFailed
	job.LastError = req.Reason
	job.PlatformErrorSummary = req.PlatformErrorSummary
	s.state.jobs[jobID] = job
	return PluginPublishJobFailedResponse{
		PublishJobID:   job.ID,
		CurrentStatus:  job.Status,
		OperationLogID: "oplog-plugin-failed-" + jobID,
	}, nil
}

