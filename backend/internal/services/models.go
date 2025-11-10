package services

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a registered user in the system.
type User struct {
	Email    string `json:"email" bson:"email"`
	Password string `json:"password,omitempty" bson:"password"`
}

// PublicUser hides sensitive user data when returning it through the API.
type PublicUser struct {
	Email string `json:"email"`
}

// ToPublic converts the User into a PublicUser without exposing the password.
func (u User) ToPublic() PublicUser {
	return PublicUser{Email: u.Email}
}

// Todo models a task stored in MongoDB.
type Todo struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email"`
	Title     string             `json:"title" bson:"title"`
	Completed bool               `json:"completed" bson:"completed"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
}

// TodoResponse is the representation exposed through the API.
type TodoResponse struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	CreatedAt time.Time `json:"createdAt"`
}

// ToResponse converts a Todo into an externally safe representation.
func (t Todo) ToResponse() TodoResponse {
	return TodoResponse{
		ID:        t.ID.Hex(),
		Email:     t.Email,
		Title:     t.Title,
		Completed: t.Completed,
		CreatedAt: t.CreatedAt,
	}
}
