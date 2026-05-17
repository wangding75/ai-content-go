package dashboard

import "context"

type Service interface {
	Summary(ctx context.Context) (SummaryResponse, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) Summary(ctx context.Context) (SummaryResponse, error) {
	return SummaryResponse{}, nil
}
