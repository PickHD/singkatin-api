package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User consist data of users
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	FullName  string             `bson:"fullname,omitempty" json:"full_name"`
	Email     string             `bson:"email,omitempty" json:"email"`
	AvatarURL string             `bson:"avatar_url" json:"avatar_url"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}
