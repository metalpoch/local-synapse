package user

import (
	"context"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/metalpoch/local-synapse/internal/repository"
)

type UpdateUser struct {
	repo repository.UserRepository
}

func NewUpdateUser(repo repository.UserRepository) *UpdateUser {
	return &UpdateUser{repo}
}

func (uc *UpdateUser) Execute(ctx context.Context, userID string, name string, file *multipart.FileHeader) error {
	var imageURL string

	if file != nil {
		// Open uploaded file
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		// Decode image
		img, format, err := image.Decode(src)
		if err != nil {
			return fmt.Errorf("invalid image format: %w", err)
		}

		// Validate format (jpg or png)
		if format != "jpeg" && format != "png" {
			return fmt.Errorf("only jpg and png are allowed, got %s", format)
		}

		// Ensure public/assets exists
		if err := os.MkdirAll("public/assets", 0755); err != nil {
			return err
		}

		// Save as {userID}.jpg
		dstPath := filepath.Join("public/assets", userID+".jpg")
		dst, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		// Encode as JPEG
		if err := jpeg.Encode(dst, img, &jpeg.Options{Quality: 80}); err != nil {
			return err
		}

		imageURL = "/assets/" + userID + ".jpg"
	}

	return uc.repo.Update(ctx, userID, name, imageURL)
}
