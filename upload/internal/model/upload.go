package model

type (
	// UploadAvatarRequest consist request data upload avatar users
	UploadAvatarRequest struct {
		FileName    string
		ContentType string
		Avatars     []byte
	}
)
