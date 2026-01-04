package order

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo OrderRepository
}

func NewService(repo OrderRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, entity *Order) error {
	entity.ID = uuid.New()
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Order, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*Order, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, entity *Order) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
