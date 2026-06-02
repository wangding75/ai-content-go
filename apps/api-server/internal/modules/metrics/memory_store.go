package metrics

import (
	"context"
	"sort"
	"strings"
	"sync"
	"time"
)

type memoryStore struct {
	mu          sync.Mutex
	templates   map[string]MetricTemplateResponse
	records     map[string]MetricRecordResponse
	idempotency map[string]idempotencyEntry
	snapshots   map[string]SummarySnapshotRow
}

type idempotencyEntry struct {
	hash      string
	refType   string
	refID     string
	createdAt time.Time
}

func NewMemoryStore() Store {
	return newMemoryStore()
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		templates:   map[string]MetricTemplateResponse{},
		records:     map[string]MetricRecordResponse{},
		idempotency: map[string]idempotencyEntry{},
		snapshots:   map[string]SummarySnapshotRow{},
	}
}

func (m *memoryStore) InsertTemplate(_ context.Context, t MetricTemplateResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := templateKey(t.ContentType, t.Platform, t.MetricCode)
	m.templates[key] = t
	return nil
}

func (m *memoryStore) FindTemplateByKey(_ context.Context, contentType, platform, metricCode string) (*MetricTemplateResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := templateKey(contentType, platform, metricCode)
	t, ok := m.templates[key]
	if !ok {
		return nil, nil
	}
	return &t, nil
}

func (m *memoryStore) ListTemplates(_ context.Context, req ListMetricTemplatesRequest) ([]MetricTemplateResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]MetricTemplateResponse, 0, len(m.templates))
	for _, item := range m.templates {
		if req.ContentType != "" && item.ContentType != req.ContentType {
			continue
		}
		if req.Platform != "" && item.Platform != req.Platform {
			continue
		}
		if req.Enabled != nil && item.Enabled != *req.Enabled {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].MetricCode < items[j].MetricCode
	})
	return items, nil
}

func (m *memoryStore) InsertRecord(_ context.Context, r MetricRecordResponse) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := recordUniqueKey(r.ProjectID, r.Platform, r.TargetID, r.ContentVersionID, r.MetricCode, r.MetricDate, r.Period)
	m.records[key] = r
	return nil
}

func (m *memoryStore) FindRecordByUniqueKey(_ context.Context, projectID, platform, targetID, contentVersionID, metricCode, metricDate, period string) (*MetricRecordResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	key := recordUniqueKey(projectID, platform, targetID, contentVersionID, metricCode, metricDate, period)
	r, ok := m.records[key]
	if !ok {
		return nil, nil
	}
	return &r, nil
}

func (m *memoryStore) ListRecords(_ context.Context, req ListMetricRecordsRequest) ([]MetricRecordResponse, int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]MetricRecordResponse, 0, len(m.records))
	for _, item := range m.records {
		if item.ProjectID != req.ProjectID {
			continue
		}
		if req.Platform != "" && item.Platform != req.Platform {
			continue
		}
		if req.TargetID != "" && item.TargetID != req.TargetID {
			continue
		}
		if req.ContentItemID != "" && item.ContentItemID != req.ContentItemID {
			continue
		}
		if req.MetricCode != "" && item.MetricCode != req.MetricCode {
			continue
		}
		if req.DateFrom != "" && item.MetricDate < req.DateFrom {
			continue
		}
		if req.DateTo != "" && item.MetricDate > req.DateTo {
			continue
		}
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].MetricDate > items[j].MetricDate
	})
	return items, len(items), nil
}

func (m *memoryStore) InsertSummarySnapshot(_ context.Context, snap SummarySnapshotRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.snapshots[snap.ID] = snap
	return nil
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

func (m *memoryStore) QuerySummary(ctx context.Context, projectID string, req MetricSummaryRequest) ([]MetricSummaryItem, int, error) {
	records, _, err := m.ListRecords(ctx, ListMetricRecordsRequest{
		ProjectID: projectID,
		Platform:  req.Platform,
		TargetID:  req.TargetID,
		DateFrom:  req.DateFrom,
		DateTo:    req.DateTo,
	})
	if err != nil {
		return nil, 0, err
	}

	codeSet := map[string]bool{}
	for _, code := range req.MetricCodes {
		codeSet[code] = true
	}

	type acc struct {
		values      []float64
		unit        string
		aggregation string
	}
	accs := map[string]*acc{}
	for _, code := range req.MetricCodes {
		accs[code] = &acc{aggregation: AggregationSum}
	}

	m.mu.Lock()
	for _, record := range records {
		if !codeSet[record.MetricCode] {
			continue
		}
		a := accs[record.MetricCode]
		for _, t := range m.templates {
			if t.ID == record.MetricTemplateID {
				a.unit = t.Unit
				a.aggregation = t.AggregationMethod
				break
			}
		}
		a.values = append(a.values, record.NormalizedValue)
	}
	m.mu.Unlock()

	items := make([]MetricSummaryItem, 0)
	total := 0
	for _, code := range req.MetricCodes {
		a := accs[code]
		if len(a.values) == 0 {
			continue
		}
		total += len(a.values)
		items = append(items, MetricSummaryItem{
			MetricCode:        code,
			Value:             aggregate(a.aggregation, a.values),
			Unit:              a.unit,
			AggregationMethod: a.aggregation,
			SourceRecordCount: len(a.values),
		})
	}
	return items, total, nil
}

func (m *memoryStore) QueryTrends(ctx context.Context, projectID string, req MetricTrendRequest) ([]MetricTrendPoint, []MetricMissingPoint, string, int, error) {
	records, _, err := m.ListRecords(ctx, ListMetricRecordsRequest{
		ProjectID:  projectID,
		Platform:   req.Platform,
		TargetID:   req.TargetID,
		MetricCode: req.MetricCode,
		DateFrom:   req.DateFrom,
		DateTo:     req.DateTo,
	})
	if err != nil {
		return nil, nil, "", 0, err
	}

	m.mu.Lock()
	var aggregation string
	for _, t := range m.templates {
		if t.MetricCode == req.MetricCode {
			aggregation = t.AggregationMethod
			break
		}
	}
	m.mu.Unlock()
	if aggregation == "" {
		aggregation = AggregationSum
	}

	buckets := bucketRecords(records, req.Bucket)
	series := make([]MetricTrendPoint, 0, len(buckets))
	sourceCount := 0
	for _, bucket := range buckets {
		val := aggregate(aggregation, bucket.values)
		series = append(series, MetricTrendPoint{
			BucketStart:       bucket.start,
			Value:             val,
			SourceRecordCount: len(bucket.values),
			Missing:           false,
		})
		sourceCount += len(bucket.values)
	}

	signature := querySignature(projectID, req)
	return series, []MetricMissingPoint{}, signature, sourceCount, nil
}

func (m *memoryStore) QueryMissingDates(_ context.Context, projectID string, req MissingMetricDatesRequest) ([]MissingMetricDateItem, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	items := make([]MissingMetricDateItem, 0)
	for _, template := range m.templates {
		if !template.Enabled || !template.Required {
			continue
		}
		if req.Platform != "" && template.Platform != req.Platform {
			continue
		}
		if req.MetricCode != "" && template.MetricCode != req.MetricCode {
			continue
		}
		for _, date := range generateDateRange(req.DateFrom, template.Period, req.DateTo) {
			key := recordUniqueKey(projectID, template.Platform, defaultTargetID(req.TargetID), "content-version-approved-1", template.MetricCode, date, template.Period)
			if _, exists := m.records[key]; exists {
				continue
			}
			items = append(items, MissingMetricDateItem{
				ContentItemID:    "content-item-1",
				ContentVersionID: "content-version-approved-1",
				PublishJobID:     "publish-job-1",
				TargetID:         defaultTargetID(req.TargetID),
				Platform:         template.Platform,
				MetricCode:       template.MetricCode,
				Period:           template.Period,
				MetricDate:       date,
				MissingReason:    "required_metric_missing",
				BackfillHint:     "backfill " + template.MetricCode + " " + date,
			})
		}
	}
	return items, nil
}

func defaultTargetID(targetID string) string {
	if targetID == "" {
		return "publish-target-1"
	}
	return targetID
}

func generateDateRange(from, periodStr, to string) []string {
	fromDate, err := time.Parse("2006-01-02", from)
	if err != nil {
		return nil
	}
	toDate, err := time.Parse("2006-01-02", to)
	if err != nil {
		return nil
	}
	var step time.Duration
	switch periodStr {
	case PeriodDay:
		step = 24 * time.Hour
	case PeriodWeek:
		step = 7 * 24 * time.Hour
	case PeriodMonth:
		step = 30 * 24 * time.Hour
	default:
		step = 24 * time.Hour
	}
	var dates []string
	for d := fromDate; !d.After(toDate); d = d.Add(step) {
		dates = append(dates, d.Format("2006-01-02"))
	}
	return dates
}

type bucketGroup struct {
	start  string
	values []float64
}

func bucketRecords(records []MetricRecordResponse, bucket string) []bucketGroup {
	groups := map[string][]float64{}
	var order []string
	for _, r := range records {
		bucketStart := truncateToBucket(r.MetricDate, bucket)
		if _, exists := groups[bucketStart]; !exists {
			order = append(order, bucketStart)
		}
		groups[bucketStart] = append(groups[bucketStart], r.NormalizedValue)
	}
	result := make([]bucketGroup, 0, len(order))
	for _, key := range order {
		result = append(result, bucketGroup{start: key, values: groups[key]})
	}
	return result
}

func truncateToBucket(dateStr, bucket string) string {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return dateStr
	}
	switch bucket {
	case PeriodWeek:
		weekday := int(t.Weekday())
		t = t.AddDate(0, 0, -weekday)
	case PeriodMonth:
		t = time.Date(t.Year(), t.Month(), 1, 0, 0, 0, 0, t.Location())
	}
	return t.Format("2006-01-02")
}

func aggregate(method string, values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	switch method {
	case AggregationSum:
		s := 0.0
		for _, v := range values {
			s += v
		}
		return s
	case AggregationAvg:
		s := 0.0
		for _, v := range values {
			s += v
		}
		return s / float64(len(values))
	case AggregationMax:
		m := values[0]
		for _, v := range values[1:] {
			if v > m {
				m = v
			}
		}
		return m
	case AggregationMin:
		m := values[0]
		for _, v := range values[1:] {
			if v < m {
				m = v
			}
		}
		return m
	case AggregationLatest:
		return values[len(values)-1]
	default:
		return values[len(values)-1]
	}
}

func querySignature(projectID string, req MetricTrendRequest) string {
	return strings.Join([]string{projectID, req.MetricCode, req.Platform, req.TargetID, req.DateFrom, req.DateTo, req.Bucket}, ":")
}

func (m *memoryStore) InsertPlatformCollectLog(ctx context.Context, log PlatformCollectLogDetailResponse) error {
	panic("not implemented")
}

func (m *memoryStore) ListPlatformCollectLogs(ctx context.Context, req ListPlatformCollectLogsRequest) ([]PlatformCollectLogResponse, int, error) {
	panic("not implemented")
}

func (m *memoryStore) GetPlatformCollectLog(ctx context.Context, collectLogID string) (*PlatformCollectLogDetailResponse, error) {
	panic("not implemented")
}

func (m *memoryStore) UpdatePlatformCollectLogStatus(ctx context.Context, collectLogID string, status string, operationLogID string) error {
	panic("not implemented")
}
