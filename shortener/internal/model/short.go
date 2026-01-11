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
		CreatedAt time.Time          `bson:"created_at"`
		UpdatedAt *time.Time         `bson:"updated_at"`
	}

	CreateShortRequest struct {
		UserID   string `json:"user_id"`
		FullURL  string `json:"full_url"`
		ShortURL string `json:"short_url"`
	}

	ClickShortResponse struct {
		FullURL string `json:"full_url"`
	}

	UpdateVisitorRequest struct {
		ShortURL string `json:"short_url"`
	}

	UpdateShortRequest struct {
		ID      string `json:"id"`
		FullURL string `json:"full_url"`
	}

	DeleteShortRequest struct {
		ID string `json:"id"`
	}
)
