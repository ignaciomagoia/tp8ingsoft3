package services

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	// ErrInvalidTodoInput indicates missing or malformed todo data.
	ErrInvalidTodoInput = errors.New("invalid todo input")
	// ErrInvalidTodoID indicates the todo ID could not be parsed.
	ErrInvalidTodoID = errors.New("invalid todo id")
)

// TodoUpdate models the fields that can be updated on a Todo.
type TodoUpdate struct {
	Title     *string
	Completed *bool
}

// TodoRepository is the storage contract required by the todo service.
type TodoRepository interface {
	List(ctx context.Context, email string) ([]Todo, error)
	Create(ctx context.Context, todo Todo) (Todo, error)
	Update(ctx context.Context, id primitive.ObjectID, update TodoUpdate) (Todo, error)
	Delete(ctx context.Context, id primitive.ObjectID) error
	Clear(ctx context.Context, email string) error
}

// MongoTodoRepository implements TodoRepository backed by MongoDB.
type MongoTodoRepository struct {
	collection *mongo.Collection
}

// NewMongoTodoRepository creates a new repository wrapper around a Mongo collection.
func NewMongoTodoRepository(collection *mongo.Collection) *MongoTodoRepository {
	return &MongoTodoRepository{collection: collection}
}

// List returns todos optionally filtered by email.
func (m *MongoTodoRepository) List(ctx context.Context, email string) ([]Todo, error) {
	filter := bson.M{}
	if email != "" {
		filter["email"] = email
	}

	cursor, err := m.collection.Find(ctx, filter, options.Find().SetSort(bson.M{"createdAt": 1}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var todos []Todo
	if err := cursor.All(ctx, &todos); err != nil {
		return nil, err
	}
	return todos, nil
}

// Create stores a todo in MongoDB and returns it with the generated ID.
func (m *MongoTodoRepository) Create(ctx context.Context, todo Todo) (Todo, error) {
	res, err := m.collection.InsertOne(ctx, todo)
	if err != nil {
		return Todo{}, err
	}
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		todo.ID = oid
	}
	return todo, nil
}

// Update modifies a todo and returns the updated version.
func (m *MongoTodoRepository) Update(ctx context.Context, id primitive.ObjectID, update TodoUpdate) (Todo, error) {
	updateDoc := bson.M{}
	if update.Title != nil {
		updateDoc["title"] = *update.Title
	}
	if update.Completed != nil {
		updateDoc["completed"] = *update.Completed
	}

	res := m.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": updateDoc},
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	var todo Todo
	if err := res.Decode(&todo); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return Todo{}, ErrNotFound
		}
		return Todo{}, err
	}
	return todo, nil
}

// Delete removes a todo by ID.
func (m *MongoTodoRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	res, err := m.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return ErrNotFound
	}
	return nil
}

// Clear remove todos optionally filtered by email.
func (m *MongoTodoRepository) Clear(ctx context.Context, email string) error {
	filter := bson.M{}
	if email != "" {
		filter["email"] = email
	}
	_, err := m.collection.DeleteMany(ctx, filter)
	return err
}

// TodoService encapsulates business logic for todo operations.
type TodoService struct {
	repo TodoRepository
	now  func() time.Time
}

// NewTodoService builds a new TodoService instance.
func NewTodoService(repo TodoRepository, now func() time.Time) *TodoService {
	if now == nil {
		now = time.Now
	}
	return &TodoService{repo: repo, now: now}
}

// List returns todos optionally filtered by user email.
func (s *TodoService) List(ctx context.Context, email string) ([]TodoResponse, error) {
	email = NormalizeEmail(email)

	todos, err := s.repo.List(ctx, email)
	if err != nil {
		return nil, err
	}

	responses := make([]TodoResponse, 0, len(todos))
	for _, todo := range todos {
		responses = append(responses, todo.ToResponse())
	}
	return responses, nil
}

// Create validates input and stores a new todo.
func (s *TodoService) Create(ctx context.Context, email, title string) (TodoResponse, error) {
	email = NormalizeEmail(email)
	title = NormalizeText(title)

	if email == "" || title == "" {
		return TodoResponse{}, ErrInvalidTodoInput
	}

	todo := Todo{
		Email:     email,
		Title:     title,
		Completed: false,
		CreatedAt: s.now(),
	}

	created, err := s.repo.Create(ctx, todo)
	if err != nil {
		return TodoResponse{}, err
	}

	return created.ToResponse(), nil
}

// Update applies the provided modification to a todo and returns the updated todo.
func (s *TodoService) Update(ctx context.Context, id string, update TodoUpdate) (TodoResponse, error) {
	if update.Title == nil && update.Completed == nil {
		return TodoResponse{}, ErrInvalidTodoInput
	}

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return TodoResponse{}, ErrInvalidTodoID
	}

	if update.Title != nil {
		title := NormalizeText(*update.Title)
		if title == "" {
			return TodoResponse{}, ErrInvalidTodoInput
		}
		update.Title = &title
	}

	updated, err := s.repo.Update(ctx, objID, update)
	if err != nil {
		return TodoResponse{}, err
	}

	return updated.ToResponse(), nil
}

// Delete removes a todo by ID.
func (s *TodoService) Delete(ctx context.Context, id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return ErrInvalidTodoID
	}
	return s.repo.Delete(ctx, objID)
}

// Clear removes todos optionally filtered by email.
func (s *TodoService) Clear(ctx context.Context, email string) error {
	email = NormalizeEmail(email)
	return s.repo.Clear(ctx, email)
}
