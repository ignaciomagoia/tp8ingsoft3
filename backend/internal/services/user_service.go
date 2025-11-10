package services

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	// ErrNotFound signals that a requested entity does not exist.
	ErrNotFound = errors.New("not found")
	// ErrInvalidUserInput indicates missing or malformed user data.
	ErrInvalidUserInput = errors.New("invalid user input")
	// ErrUserAlreadyExists is returned when trying to create a duplicated user.
	ErrUserAlreadyExists = errors.New("user already exists")
	// ErrInvalidCredentials is returned when the email/password combination is wrong.
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserRepository is the storage contract required by the user service.
type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (User, error)
	Insert(ctx context.Context, user User) error
	List(ctx context.Context) ([]User, error)
	Clear(ctx context.Context) error
}

// MongoUserRepository implements UserRepository backed by MongoDB.
type MongoUserRepository struct {
	collection *mongo.Collection
}

// NewMongoUserRepository creates a new repository wrapper around a Mongo collection.
func NewMongoUserRepository(collection *mongo.Collection) *MongoUserRepository {
	return &MongoUserRepository{collection: collection}
}

// FindByEmail retrieves a user by email or returns ErrNotFound.
func (m *MongoUserRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	var user User
	err := m.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return User{}, ErrNotFound
	}
	return user, err
}

// Insert stores the provided user in MongoDB.
func (m *MongoUserRepository) Insert(ctx context.Context, user User) error {
	_, err := m.collection.InsertOne(ctx, user)
	return err
}

// List retrieves all users.
func (m *MongoUserRepository) List(ctx context.Context) ([]User, error) {
	cursor, err := m.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

// Clear removes all users from the collection.
func (m *MongoUserRepository) Clear(ctx context.Context) error {
	_, err := m.collection.DeleteMany(ctx, bson.M{})
	return err
}

// UserService encapsulates business logic for user operations.
type UserService struct {
	repo UserRepository
}

// NewUserService builds a new UserService instance.
func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

// Register validates and stores a user; returns high-level domain errors.
func (s *UserService) Register(ctx context.Context, user User) error {
	user.Email = NormalizeEmail(user.Email)
	user.Password = NormalizeText(user.Password)

	if user.Email == "" || user.Password == "" {
		return ErrInvalidUserInput
	}

	_, err := s.repo.FindByEmail(ctx, user.Email)
	if err == nil {
		return ErrUserAlreadyExists
	}
	if !errors.Is(err, ErrNotFound) {
		return err
	}

	return s.repo.Insert(ctx, user)
}

// Login validates the provided credentials.
func (s *UserService) Login(ctx context.Context, email, password string) error {
	email = NormalizeEmail(email)
	password = NormalizeText(password)

	if email == "" || password == "" {
		return ErrInvalidCredentials
	}

	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return ErrInvalidCredentials
		}
		return err
	}
	if user.Password != password {
		return ErrInvalidCredentials
	}
	return nil
}

// List returns all users in their public representation.
func (s *UserService) List(ctx context.Context) ([]PublicUser, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	public := make([]PublicUser, 0, len(users))
	for _, u := range users {
		public = append(public, u.ToPublic())
	}
	return public, nil
}

// Clear removes all user records.
func (s *UserService) Clear(ctx context.Context) error {
	return s.repo.Clear(ctx)
}
