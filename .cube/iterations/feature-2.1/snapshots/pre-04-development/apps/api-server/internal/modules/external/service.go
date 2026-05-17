package external

import (
	"context"
	"errors"
)

var (
	ErrValidation         = errors.New("validation error")
	ErrNotFound           = errors.New("not found")
	ErrConflict           = errors.New("conflict")
	ErrExternalAutomation = errors.New("external automation error")
)

type Service interface {
	ListProviders(ctx context.Context, req ListProvidersRequest) (PagedProvidersResponse, error)
	CreateProvider(ctx context.Context, req CreateProviderRequest) (CreateProviderResponse, error)
	ListBindings(ctx context.Context, req ListBindingsRequest) (PagedBindingsResponse, error)
	CreateBinding(ctx context.Context, req CreateBindingRequest) (CreateBindingResponse, error)
}

type service struct{}

func NewService() Service {
	return &service{}
}

func (s *service) ListProviders(context.Context, ListProvidersRequest) (PagedProvidersResponse, error) {
	return PagedProvidersResponse{}, ErrValidation
}

func (s *service) CreateProvider(context.Context, CreateProviderRequest) (CreateProviderResponse, error) {
	return CreateProviderResponse{}, ErrValidation
}

func (s *service) ListBindings(context.Context, ListBindingsRequest) (PagedBindingsResponse, error) {
	return PagedBindingsResponse{}, ErrValidation
}

func (s *service) CreateBinding(context.Context, CreateBindingRequest) (CreateBindingResponse, error) {
	return CreateBindingResponse{}, ErrValidation
}
