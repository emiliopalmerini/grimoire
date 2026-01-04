package scaffold

import "fmt"

func sqliteRepositoryTemplate(name, namePascal, moduleImportPath string) string {
	return fmt.Sprintf(`package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"%s"
)

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

func (r *SQLiteRepository) Create(ctx context.Context, entity *%s.%s) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO %s (id, name) VALUES (?, ?)",
		entity.ID.String(), entity.Name)
	return err
}

func (r *SQLiteRepository) GetByID(ctx context.Context, id uuid.UUID) (*%s.%s, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, name FROM %s WHERE id = ?",
		id.String())

	var entity %s.%s
	var idStr string
	if err := row.Scan(&idStr, &entity.Name); err != nil {
		return nil, err
	}
	entity.ID, _ = uuid.Parse(idStr)
	return &entity, nil
}

func (r *SQLiteRepository) List(ctx context.Context) ([]*%s.%s, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM %s")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*%s.%s
	for rows.Next() {
		var entity %s.%s
		var idStr string
		if err := rows.Scan(&idStr, &entity.Name); err != nil {
			return nil, err
		}
		entity.ID, _ = uuid.Parse(idStr)
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func (r *SQLiteRepository) Update(ctx context.Context, entity *%s.%s) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE %s SET name = ? WHERE id = ?",
		entity.Name, entity.ID.String())
	return err
}

func (r *SQLiteRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM %s WHERE id = ?",
		id.String())
	return err
}
`, moduleImportPath,
		name, namePascal, name,
		name, namePascal, name, name, namePascal,
		name, namePascal, name, name, namePascal, name, namePascal,
		name, namePascal, name, name)
}

func postgresRepositoryTemplate(name, namePascal, moduleImportPath string) string {
	return fmt.Sprintf(`package persistence

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"%s"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Create(ctx context.Context, entity *%s.%s) error {
	_, err := r.db.ExecContext(ctx,
		"INSERT INTO %s (id, name) VALUES ($1, $2)",
		entity.ID, entity.Name)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*%s.%s, error) {
	row := r.db.QueryRowContext(ctx,
		"SELECT id, name FROM %s WHERE id = $1",
		id)

	var entity %s.%s
	if err := row.Scan(&entity.ID, &entity.Name); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *PostgresRepository) List(ctx context.Context) ([]*%s.%s, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id, name FROM %s")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*%s.%s
	for rows.Next() {
		var entity %s.%s
		if err := rows.Scan(&entity.ID, &entity.Name); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func (r *PostgresRepository) Update(ctx context.Context, entity *%s.%s) error {
	_, err := r.db.ExecContext(ctx,
		"UPDATE %s SET name = $1 WHERE id = $2",
		entity.Name, entity.ID)
	return err
}

func (r *PostgresRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx,
		"DELETE FROM %s WHERE id = $1",
		id)
	return err
}
`, moduleImportPath,
		name, namePascal, name,
		name, namePascal, name, name, namePascal,
		name, namePascal, name, name, namePascal, name, namePascal,
		name, namePascal, name, name)
}

func mongoRepositoryTemplate(name, namePascal, moduleImportPath string) string {
	return fmt.Sprintf(`package persistence

import (
	"context"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"%s"
)

type MongoRepository struct {
	collection *mongo.Collection
}

func NewMongoRepository(db *mongo.Database) *MongoRepository {
	return &MongoRepository{collection: db.Collection("%s")}
}

func (r *MongoRepository) Create(ctx context.Context, entity *%s.%s) error {
	_, err := r.collection.InsertOne(ctx, entity)
	return err
}

func (r *MongoRepository) GetByID(ctx context.Context, id uuid.UUID) (*%s.%s, error) {
	var entity %s.%s
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&entity)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *MongoRepository) List(ctx context.Context) ([]*%s.%s, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var entities []*%s.%s
	if err := cursor.All(ctx, &entities); err != nil {
		return nil, err
	}
	return entities, nil
}

func (r *MongoRepository) Update(ctx context.Context, entity *%s.%s) error {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": entity.ID}, entity)
	return err
}

func (r *MongoRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
`, moduleImportPath, name,
		name, namePascal,
		name, namePascal, name, namePascal,
		name, namePascal, name, namePascal,
		name, namePascal)
}
