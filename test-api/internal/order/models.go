package order

import "github.com/google/uuid"

type Order struct {
	ID   uuid.UUID
	Name string
}
