package scaffold

import "fmt"

func modelsTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package %s

import "github.com/google/uuid"

type %s struct {
	ID   uuid.UUID
	Name string
}
`, name, namePascal)
}

func serviceTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package %s

import (
	"context"

	"github.com/google/uuid"
)

type Service struct {
	repo %sRepository
}

func NewService(repo %sRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, entity *%s) error {
	entity.ID = uuid.New()
	return s.repo.Create(ctx, entity)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*%s, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]*%s, error) {
	return s.repo.List(ctx)
}

func (s *Service) Update(ctx context.Context, entity *%s) error {
	return s.repo.Update(ctx, entity)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
`, name, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal)
}

func repositoryInterfaceTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package %s

import (
	"context"

	"github.com/google/uuid"
)

type %sRepository interface {
	Create(ctx context.Context, entity *%s) error
	GetByID(ctx context.Context, id uuid.UUID) (*%s, error)
	List(ctx context.Context) ([]*%s, error)
	Update(ctx context.Context, entity *%s) error
	Delete(ctx context.Context, id uuid.UUID) error
}
`, name, namePascal, namePascal, namePascal, namePascal, namePascal)
}
