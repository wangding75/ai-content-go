package schedule

import "context"

type Service interface {
	CreateSchedule(ctx context.Context, req CreateScheduleRequest) (ScheduleResponse, error)
}

type scheduleService struct{}

func NewService() Service {
	return &scheduleService{}
}

func (s *scheduleService) CreateSchedule(ctx context.Context, req CreateScheduleRequest) (ScheduleResponse, error) {
	return ScheduleResponse{ScheduleID: "schedule-placeholder", Status: "not_implemented"}, nil
}
