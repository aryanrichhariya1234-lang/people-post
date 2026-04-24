package services

import (
	"context"
	"mime/multipart"

	"peoplepost/internal/config"

	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

type UploadResult struct {
	URL      string `json:"url"`
	PublicID string `json:"public_id"`
}

func UploadToCloudinary(file multipart.File) (*UploadResult, error) {
	result, err := config.Cloudinary.Upload.Upload(
		context.Background(),
		file,
		uploader.UploadParams{
			Folder: "posts",
		},
	)
	if err != nil {
		return nil, err
	}

	return &UploadResult{
		URL:      result.SecureURL,
		PublicID: result.PublicID,
	}, nil
}