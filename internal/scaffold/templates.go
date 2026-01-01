package scaffold

import "fmt"

func modelsTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package %s

type %s struct {
	ID string
}

type Create%sRequest struct{}

type Update%sRequest struct {
	ID string
}
`, name, namePascal, namePascal, namePascal)
}

func serviceTemplate(name, namePascal, moduleImportPath string) string {
	return fmt.Sprintf(`package %s

type Service struct {
	repo %sRepository
}

func NewService(repo %sRepository) *Service {
	return &Service{repo: repo}
}
`, name, namePascal, namePascal)
}

func repositoryInterfaceTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package %s

import "context"

type %sRepository interface {
	Create(ctx context.Context, entity *%s) error
	GetByID(ctx context.Context, id string) (*%s, error)
	List(ctx context.Context) ([]*%s, error)
	Update(ctx context.Context, entity *%s) error
	Delete(ctx context.Context, id string) error
}
`, name, namePascal, namePascal, namePascal, namePascal, namePascal)
}

func createCommandTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package commands

import "context"

type Create%sCommand struct {
	// Add fields
}

type Create%sHandler struct {
	// Add dependencies
}

func NewCreate%sHandler() *Create%sHandler {
	return &Create%sHandler{}
}

func (h *Create%sHandler) Handle(ctx context.Context, cmd Create%sCommand) error {
	return nil
}
`, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal)
}

func updateCommandTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package commands

import "context"

type Update%sCommand struct {
	ID string
	// Add fields
}

type Update%sHandler struct {
	// Add dependencies
}

func NewUpdate%sHandler() *Update%sHandler {
	return &Update%sHandler{}
}

func (h *Update%sHandler) Handle(ctx context.Context, cmd Update%sCommand) error {
	return nil
}
`, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal)
}

func deleteCommandTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package commands

import "context"

type Delete%sCommand struct {
	ID string
}

type Delete%sHandler struct {
	// Add dependencies
}

func NewDelete%sHandler() *Delete%sHandler {
	return &Delete%sHandler{}
}

func (h *Delete%sHandler) Handle(ctx context.Context, cmd Delete%sCommand) error {
	return nil
}
`, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal)
}

func getQueryTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package queries

import "context"

type Get%sQuery struct {
	ID string
}

type Get%sResult struct {
	// Add fields
}

type Get%sHandler struct {
	// Add dependencies
}

func NewGet%sHandler() *Get%sHandler {
	return &Get%sHandler{}
}

func (h *Get%sHandler) Handle(ctx context.Context, query Get%sQuery) (*Get%sResult, error) {
	return nil, nil
}
`, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal)
}

func listQueryTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package queries

import "context"

type List%sQuery struct {
	Limit  int
	Offset int
}

type List%sResult struct {
	Items []%sItem
	Total int
}

type %sItem struct {
	// Add fields
}

type List%sHandler struct {
	// Add dependencies
}

func NewList%sHandler() *List%sHandler {
	return &List%sHandler{}
}

func (h *List%sHandler) Handle(ctx context.Context, query List%sQuery) (*List%sResult, error) {
	return nil, nil
}
`, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal, namePascal)
}

func httpHandlerTemplate(name, namePascal, apiType string) string {
	responseType := "JSON"
	if apiType == "html" {
		responseType = "HTML"
	}

	return fmt.Sprintf(`package http

import (
	"net/http"
)

type Handler struct {
	// Add dependencies (command/query handlers)
}

func NewHandler() *Handler {
	return &Handler{}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	// %s response
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	// %s response
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	// %s response
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	// %s response
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	// %s response
}
`, responseType, responseType, responseType, responseType, responseType)
}

func httpRoutesTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package http

import "github.com/go-chi/chi/v5"

func RegisterRoutes(r chi.Router, h *Handler) {
	r.Route("/%s", func(r chi.Router) {
		r.Get("/", h.List)
		r.Post("/", h.Create)
		r.Get("/{id}", h.Get)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}
`, name)
}

func amqpConsumerTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package amqp

type Consumer struct {
	// Add dependencies
}

func NewConsumer() *Consumer {
	return &Consumer{}
}

func (c *Consumer) Handle%sCreated(body []byte) error {
	return nil
}

func (c *Consumer) Handle%sUpdated(body []byte) error {
	return nil
}

func (c *Consumer) Handle%sDeleted(body []byte) error {
	return nil
}
`, namePascal, namePascal, namePascal)
}

func indexViewTemplate(name, namePascal string) string {
	return fmt.Sprintf(`<div id="%s-list">
  <h1>%s List</h1>

  <table>
    <thead>
      <tr>
        <th>ID</th>
        <th>Actions</th>
      </tr>
    </thead>
    <tbody hx-target="closest tr" hx-swap="outerHTML">
      <!-- Items rendered here -->
    </tbody>
  </table>

  <button hx-get="/%s/new" hx-target="#%s-form" hx-swap="innerHTML">
    New %s
  </button>

  <div id="%s-form"></div>
</div>
`, name, namePascal, name, name, namePascal, name)
}

func formViewTemplate(name, namePascal string) string {
	return fmt.Sprintf(`<form hx-post="/%s" hx-target="#%s-list" hx-swap="outerHTML">
  <h2>Create %s</h2>

  <!-- Add form fields -->

  <button type="submit">Save</button>
  <button type="button" hx-get="/%s" hx-target="#%s-list" hx-swap="outerHTML">Cancel</button>
</form>
`, name, name, namePascal, name, name)
}

func sqliteRepositoryTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package persistence

import (
	"context"
	"database/sql"

	"%s"
)

type SQLite%sRepository struct {
	db *sql.DB
}

func NewSQLite%sRepository(db *sql.DB) *SQLite%sRepository {
	return &SQLite%sRepository{db: db}
}

func (r *SQLite%sRepository) Create(ctx context.Context, entity *%s.%s) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO %s (id) VALUES (?)", entity.ID)
	return err
}

func (r *SQLite%sRepository) GetByID(ctx context.Context, id string) (*%s.%s, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id FROM %s WHERE id = ?", id)
	var entity %s.%s
	if err := row.Scan(&entity.ID); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *SQLite%sRepository) List(ctx context.Context) ([]*%s.%s, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id FROM %s")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*%s.%s
	for rows.Next() {
		var entity %s.%s
		if err := rows.Scan(&entity.ID); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func (r *SQLite%sRepository) Update(ctx context.Context, entity *%s.%s) error {
	_, err := r.db.ExecContext(ctx, "UPDATE %s SET id = ? WHERE id = ?", entity.ID, entity.ID)
	return err
}

func (r *SQLite%sRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM %s WHERE id = ?", id)
	return err
}
`, name,
		namePascal,
		namePascal, namePascal, namePascal,
		namePascal, name, namePascal, name,
		namePascal, name, namePascal, name, name, namePascal,
		namePascal, name, namePascal, name, name, namePascal, name, namePascal,
		namePascal, name, namePascal, name,
		namePascal, name)
}

func postgresRepositoryTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package persistence

import (
	"context"
	"database/sql"

	"%s"
)

type Postgres%sRepository struct {
	db *sql.DB
}

func NewPostgres%sRepository(db *sql.DB) *Postgres%sRepository {
	return &Postgres%sRepository{db: db}
}

func (r *Postgres%sRepository) Create(ctx context.Context, entity *%s.%s) error {
	_, err := r.db.ExecContext(ctx, "INSERT INTO %s (id) VALUES ($1)", entity.ID)
	return err
}

func (r *Postgres%sRepository) GetByID(ctx context.Context, id string) (*%s.%s, error) {
	row := r.db.QueryRowContext(ctx, "SELECT id FROM %s WHERE id = $1", id)
	var entity %s.%s
	if err := row.Scan(&entity.ID); err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *Postgres%sRepository) List(ctx context.Context) ([]*%s.%s, error) {
	rows, err := r.db.QueryContext(ctx, "SELECT id FROM %s")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entities []*%s.%s
	for rows.Next() {
		var entity %s.%s
		if err := rows.Scan(&entity.ID); err != nil {
			return nil, err
		}
		entities = append(entities, &entity)
	}
	return entities, rows.Err()
}

func (r *Postgres%sRepository) Update(ctx context.Context, entity *%s.%s) error {
	_, err := r.db.ExecContext(ctx, "UPDATE %s SET id = $1 WHERE id = $2", entity.ID, entity.ID)
	return err
}

func (r *Postgres%sRepository) Delete(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM %s WHERE id = $1", id)
	return err
}
`, name,
		namePascal,
		namePascal, namePascal, namePascal,
		namePascal, name, namePascal, name,
		namePascal, name, namePascal, name, name, namePascal,
		namePascal, name, namePascal, name, name, namePascal, name, namePascal,
		namePascal, name, namePascal, name,
		namePascal, name)
}

func mongoRepositoryTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package persistence

import (
	"context"

	"%s"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Mongo%sRepository struct {
	collection *mongo.Collection
}

func NewMongo%sRepository(db *mongo.Database) *Mongo%sRepository {
	return &Mongo%sRepository{collection: db.Collection("%s")}
}

func (r *Mongo%sRepository) Create(ctx context.Context, entity *%s.%s) error {
	_, err := r.collection.InsertOne(ctx, entity)
	return err
}

func (r *Mongo%sRepository) GetByID(ctx context.Context, id string) (*%s.%s, error) {
	var entity %s.%s
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&entity)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *Mongo%sRepository) List(ctx context.Context) ([]*%s.%s, error) {
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

func (r *Mongo%sRepository) Update(ctx context.Context, entity *%s.%s) error {
	_, err := r.collection.ReplaceOne(ctx, bson.M{"_id": entity.ID}, entity)
	return err
}

func (r *Mongo%sRepository) Delete(ctx context.Context, id string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}
`, name,
		namePascal,
		namePascal, namePascal, namePascal, name,
		namePascal, name, namePascal,
		namePascal, name, namePascal, name, namePascal,
		namePascal, name, namePascal, name, namePascal,
		namePascal, name, namePascal,
		namePascal)
}
