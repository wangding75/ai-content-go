package strategy

import (
	"context"
	"sort"
	"sync"
	"time"
)

type memoryStore struct {
	mu             sync.Mutex
	suggestionRuns map[string]StrategySuggestionRunResponse
	suggestions    map[string]StrategySuggestionDetailResponse
	executionLogs  map[string][]ExecutionLogResponse
	idempotency    map[string]idempotencyEntry
}

type idempotencyEntry struct {
	hash      string
	refType   string
	refID     string
	createdAt time.Time
}

func NewMemoryStore() Store {
	return &memoryStore{
		suggestionRuns: map[string]StrategySuggestionRunResponse{},
		suggestions:    map[string]StrategySuggestionDetailResponse{},
		executionLogs:  map[string][]ExecutionLogResponse{},
		idempotency:    map[string]idempotencyEntry{},
	}
}

func (m *memoryStore) InsertSuggestionRun(_ context.Context, run StrategySuggestionRunResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.suggestionRuns[run.ID] = run
	return nil
}

func (m *memoryStore) FindSuggestionRunByID(_ context.Context, id string) (*StrategySuggestionRunResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	run, ok := m.suggestionRuns[id]
	if !ok {
		return nil, nil
	}
	return &run, nil
}

func (m *memoryStore) UpdateSuggestionRunStatus(_ context.Context, id, status, failureReason string, suggestionCount int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	run, ok := m.suggestionRuns[id]
	if !ok {
		return nil
	}
	run.Status = status
	run.FailureReason = failureReason
	run.SuggestionCount = suggestionCount
	run.UpdatedAt = time.Now().UTC()
	m.suggestionRuns[id] = run
	return nil
}

func (m *memoryStore) InsertSuggestion(_ context.Context, s StrategySuggestionDetailResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.suggestions[s.ID] = s
	return nil
}

func (m *memoryStore) FindSuggestionByID(_ context.Context, id string) (*StrategySuggestionDetailResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.suggestions[id]
	if !ok {
		return nil, nil
	}
	return &s, nil
}

func (m *memoryStore) UpdateSuggestionStatus(_ context.Context, id, status string, fields map[string]any) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, ok := m.suggestions[id]
	if !ok {
		return nil
	}
	s.Status = status
	s.UpdatedAt = time.Now().UTC()
	for k, v := range fields {
		switch k {
		case "ignored_reason":
			if val, ok := v.(string); ok {
				s.IgnoredReason = val
			}
		case "ignored_note":
			if val, ok := v.(string); ok {
				s.IgnoredNote = val
			}
		case "confirmed_at":
			if val, ok := v.(*time.Time); ok {
				s.ConfirmedAt = val
			}
		case "ignored_at":
			if val, ok := v.(*time.Time); ok {
				s.IgnoredAt = val
			}
		case "executed_at":
			if val, ok := v.(*time.Time); ok {
				s.ExecutedAt = val
			}
		}
	}
	m.suggestions[id] = s
	return nil
}

func (m *memoryStore) ListSuggestions(_ context.Context, projectID string, req ListStrategySuggestionsRequest) ([]StrategySuggestionDetailResponse, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]StrategySuggestionDetailResponse, 0)
	for _, s := range m.suggestions {
		if s.ProjectID != projectID {
			continue
		}
		if req.Status != "" && s.Status != req.Status {
			continue
		}
		if req.SuggestionType != "" && s.SuggestionType != req.SuggestionType {
			continue
		}
		if req.RiskLevel != "" && s.RiskLevel != req.RiskLevel {
			continue
		}
		if req.Confidence != "" && s.Confidence != req.Confidence {
			continue
		}
		if req.DateFrom != "" && s.DateFrom < req.DateFrom {
			continue
		}
		if req.DateTo != "" && s.DateTo > req.DateTo {
			continue
		}
		items = append(items, s)
	}
	sort.Slice(items, func(i, j int) bool {
		if req.Order == "asc" {
			return items[i].CreatedAt.Before(items[j].CreatedAt)
		}
		return items[i].CreatedAt.After(items[j].CreatedAt)
	})
	return items, len(items), nil
}

func (m *memoryStore) InsertExecutionLog(_ context.Context, log ExecutionLogResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.executionLogs[log.SuggestionID] = append(m.executionLogs[log.SuggestionID], log)
	return nil
}

func (m *memoryStore) ListExecutionLogs(_ context.Context, suggestionID string, page, pageSize int) ([]ExecutionLogResponse, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	logs := m.executionLogs[suggestionID]
	total := len(logs)
	start := (page - 1) * pageSize
	if start > total {
		start = total
	}
	end := start + pageSize
	if end > total {
		end = total
	}
	result := make([]ExecutionLogResponse, 0)
	if start < total {
		result = append(result, logs[start:end]...)
	}
	return result, total, nil
}

func (m *memoryStore) CheckIdempotency(_ context.Context, scope, endpoint, key, hash string) (string, string, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ik := scope + ":" + endpoint + ":" + key
	entry, ok := m.idempotency[ik]
	if !ok {
		return "", "", false, nil
	}
	if entry.hash != hash {
		return "", "", true, nil
	}
	return entry.refType, entry.refID, false, nil
}

func (m *memoryStore) StoreIdempotency(_ context.Context, scope, endpoint, key, hash, refType, refID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	ik := scope + ":" + endpoint + ":" + key
	m.idempotency[ik] = idempotencyEntry{hash: hash, refType: refType, refID: refID, createdAt: time.Now().UTC()}
	return nil
}
