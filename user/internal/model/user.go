package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type (
	// User consist data of users
	User struct {
		ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
		FullName  string             `bson:"fullname,omitempty" json:"full_name"`
		Email     string             `bson:"email,omitempty" json:"email"`
		AvatarURL string             `bson:"avatar_url" json:"avatar_url"`
		CreatedAt time.Time          `bson:"created_at" json:"created_at"`
	}

	// UserShorts consist data of user shorts
	UserShorts struct {
		ID       string `json:"id"`
		FullURL  string `json:"full_url"`
		ShortURL string `json:"short_url"`
		Visited  int64  `json:"visited"`
	}

	// ShortUserRequest consist request data generate/update short users
	ShortUserRequest struct {
		FullURL string `json:"full_url"`
	}

	// ShortUserResponse consist response data when success generate/update short users
	ShortUserResponse struct {
		ShortURL string `json:"short_url"`
		Method   string `json:"method"`
	}

	// GenerateShortUserMessage consist message short users to publish
	GenerateShortUserMessage struct {
		FullURL  string `json:"full_url"`
		ShortURL string `json:"short_url"`
		UserID   string `json:"user_id"`
	}

	// EditProfileRequest consist request data edit profile users
	EditProfileRequest struct {
		FullName string `json:"full_name"`
	}

	// UploadAvatarRequest consist request data upload avatar users
	UploadAvatarRequest struct {
		FileName    string
		ContentType string
		Avatars     []byte
	}

	// UploadAvatarResponse consist response data when success upload avatar users
	UploadAvatarResponse struct {
		FileURL string
	}
)
