package order

import (
	"context"

	"github.com/google/uuid"
)

type OrderRepository interface {
	Create(ctx context.Context, entity *Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*Order, error)
	List(ctx context.Context) ([]*Order, error)
	Update(ctx context.Context, entity *Order) error
	Delete(ctx context.Context, id uuid.UUID) error
}
