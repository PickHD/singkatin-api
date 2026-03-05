package request

// UploadAvatarRequest consist request data upload avatar users
type UploadAvatarRequest struct {
	FileName    string `json:"file_name"`
	ContentType string `json:"content_type"`
	Avatars     []byte `json:"avatars"`
}
