package user

import (
	"context"

	"github.com/google/uuid"
)

type UserRepository interface {
	Create(ctx context.Context, entity *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context) ([]*User, error)
	Update(ctx context.Context, entity *User) error
	Delete(ctx context.Context, id uuid.UUID) error
}
