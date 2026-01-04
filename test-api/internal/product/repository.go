package product

import (
	"context"

	"github.com/google/uuid"
)

type ProductRepository interface {
	Create(ctx context.Context, entity *Product) error
	GetByID(ctx context.Context, id uuid.UUID) (*Product, error)
	List(ctx context.Context) ([]*Product, error)
	Update(ctx context.Context, entity *Product) error
	Delete(ctx context.Context, id uuid.UUID) error
}
