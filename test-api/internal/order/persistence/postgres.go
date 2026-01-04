package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/test/api/internal/order"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, entity *order.Order) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO order (id, name) VALUES ($1, $2)",
		entity.ID, entity.Name)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*order.Order, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, name FROM order WHERE id = $1",
		id)

	var entity order.Order
	if err := row.Scan(&entity.ID, &entity.Name); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]*order.Order, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM order")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*order.Order
	for rows.Next() {
		var entity order.Order
		if err := rows.Scan(&entity.ID, &entity.Name); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func (r *PostgresRepository) Update(ctx context.Context, entity *order.Order) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE order SET name = $1 WHERE id = $2",
		entity.Name, entity.ID)
	return err
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM order WHERE id = $1",
		id)
	return err
}
