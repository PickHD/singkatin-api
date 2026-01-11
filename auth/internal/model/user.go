package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	// User consist data of users
	User struct {
		ID         primitive.ObjectID `bson:"_id,omitempty"`
		FullName   string             `bson:"fullname,omitempty"`
		IsVerified bool               `bson:"is_verified,omitempty"`
		Email      string             `bson:"email,omitempty"`
		Password   string             `bson:"password,omitempty"`
		CreatedAt  time.Time          `bson:"created_at,omitempty"`
	}
)
