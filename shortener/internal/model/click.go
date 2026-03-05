package model

import "go.mongodb.org/mongo-driver/bson/primitive"

type Click struct {
	ID        primitive.ObjectID `bson:"_id"`
	ShortURL  string             `bson:"short_url"`
	UserAgent string             `bson:"user_agent"`
	IPAddress string             `bson:"ip_address"`
	Referer   string             `bson:"referer"`
	CreatedAt int64              `bson:"created_at"`
}
