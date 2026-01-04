package product

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo ProductRepository
}

func NewService(repo ProductRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, entity *Product) error {
	entity.ID = uuid.New()
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*Product, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*Product, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, entity *Product) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
