package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	Short struct {
		ID        primitive.ObjectID `bson:"_id"`
		UserID    string             `bson:"user_id"`
		FullURL   string             `bson:"full_url"`
		ShortURL  string             `bson:"short_url"`
		Visited   int64              `bson:"visited"`
		ExpiresAt *time.Time         `bson:"expires_at"`
		CreatedAt time.Time          `bson:"created_at"`
		UpdatedAt *time.Time         `bson:"updated_at"`
	}
)
