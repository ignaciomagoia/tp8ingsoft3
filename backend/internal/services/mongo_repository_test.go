package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func collectionNamespace(mt *mtest.T) string {
	return fmt.Sprintf("%s.%s", mt.Coll.Database().Name(), mt.Coll.Name())
}

// TestMongoUserRepositoryExercisesCRUD covers the Mongo-backed user repository with mock responses.
func TestMongoUserRepositoryExercisesCRUD(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock).CreateCollection(false))

	mt.Run("find by email success", func(mt *mtest.T) {
		repo := NewMongoUserRepository(mt.Coll)
		doc := bson.D{
			{"email", "user@example.com"},
			{"password", "secret"},
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, collectionNamespace(mt), mtest.FirstBatch, doc))

		user, err := repo.FindByEmail(context.Background(), "user@example.com")
		if err != nil {
			mt.Fatalf("expected user, got error: %v", err)
		}
		if user.Email != "user@example.com" || user.Password != "secret" {
			mt.Fatalf("unexpected user: %+v", user)
		}
	})

	mt.Run("find by email not found", func(mt *mtest.T) {
		repo := NewMongoUserRepository(mt.Coll)
		mt.AddMockResponses(mtest.CreateCursorResponse(0, collectionNamespace(mt), mtest.FirstBatch))

		_, err := repo.FindByEmail(context.Background(), "missing@example.com")
		if err == nil || err != ErrNotFound {
			mt.Fatalf("expected ErrNotFound, got %v", err)
		}
	})

	mt.Run("insert and list users", func(mt *mtest.T) {
		repo := NewMongoUserRepository(mt.Coll)
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))

		err := repo.Insert(context.Background(), User{Email: "alice@example.com", Password: "secret"})
		if err != nil {
			mt.Fatalf("insert failed: %v", err)
		}

		docs := []bson.D{
			{{"email", "alice@example.com"}, {"password", "secret"}},
			{{"email", "bob@example.com"}, {"password", "hidden"}},
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, collectionNamespace(mt), mtest.FirstBatch, docs...))

		users, err := repo.List(context.Background())
		if err != nil {
			mt.Fatalf("list failed: %v", err)
		}
		if len(users) != 2 {
			mt.Fatalf("expected 2 users, got %d", len(users))
		}
	})

	mt.Run("clear users succeeds", func(mt *mtest.T) {
		repo := NewMongoUserRepository(mt.Coll)
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 2}))

		if err := repo.Clear(context.Background()); err != nil {
			mt.Fatalf("clear failed: %v", err)
		}
	})
}

// TestMongoTodoRepositoryExercisesCRUD covers the Mongo-backed todo repository with mock responses.
func TestMongoTodoRepositoryExercisesCRUD(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock).CreateCollection(false))

	mt.Run("create todo stores document", func(mt *mtest.T) {
		repo := NewMongoTodoRepository(mt.Coll)
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))

		created, err := repo.Create(context.Background(), Todo{
			Email:     "user@example.com",
			Title:     "Test task",
			Completed: false,
			CreatedAt: time.Now(),
		})
		if err != nil {
			mt.Fatalf("create failed: %v", err)
		}
		if created.Email != "user@example.com" {
			mt.Fatalf("unexpected email: %s", created.Email)
		}
	})

	mt.Run("list todos returns stored data", func(mt *mtest.T) {
		repo := NewMongoTodoRepository(mt.Coll)
		now := time.Now()
		doc := bson.D{
			{"_id", primitive.NewObjectID()},
			{"email", "user@example.com"},
			{"title", "Sample"},
			{"completed", false},
			{"createdAt", now},
		}
		mt.AddMockResponses(mtest.CreateCursorResponse(0, collectionNamespace(mt), mtest.FirstBatch, doc))

		todos, err := repo.List(context.Background(), "user@example.com")
		if err != nil {
			mt.Fatalf("list failed: %v", err)
		}
		if len(todos) != 1 || todos[0].Title != "Sample" {
			mt.Fatalf("unexpected todos: %+v", todos)
		}
	})

	mt.Run("update todo returns modified document", func(mt *mtest.T) {
		repo := NewMongoTodoRepository(mt.Coll)
		id := primitive.NewObjectID()
		doc := bson.D{
			{"_id", id},
			{"email", "user@example.com"},
			{"title", "Updated"},
			{"completed", true},
			{"createdAt", time.Now()},
		}
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "value", Value: doc}))

		title := "Updated"
		completed := true
		updated, err := repo.Update(context.Background(), id, TodoUpdate{
			Title:     &title,
			Completed: &completed,
		})
		if err != nil {
			mt.Fatalf("update failed: %v", err)
		}
		if updated.ID != id || !updated.Completed {
			mt.Fatalf("unexpected updated todo: %+v", updated)
		}
	})

	mt.Run("delete todo removes document", func(mt *mtest.T) {
		repo := NewMongoTodoRepository(mt.Coll)
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}))

		if err := repo.Delete(context.Background(), primitive.NewObjectID()); err != nil {
			mt.Fatalf("delete failed: %v", err)
		}
	})

	mt.Run("clear todos removes by email", func(mt *mtest.T) {
		repo := NewMongoTodoRepository(mt.Coll)
		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 3}))

		if err := repo.Clear(context.Background(), "user@example.com"); err != nil {
			mt.Fatalf("clear failed: %v", err)
		}
	})
}

// TestConnectMongoCancelledContext ensures ConnectMongo respects context cancellation.
func TestConnectMongoCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := ConnectMongo(ctx, "mongodb://localhost:27099")
	if err == nil {
		t.Fatalf("expected error from cancelled context")
	}
}
