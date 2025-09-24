package services

import (
	"context"
	"log"
	"mime/multipart"
	"movie-ticket-booking/config"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

func UploadToCloudinary(fileHeader *multipart.FileHeader, folder string) (string, error) {
	cfg := config.GetCloudinaryConfig()

	cld, err := cloudinary.NewFromParams(
		cfg.Name,
		cfg.APIKEY,
		cfg.APISECRET,
	)
	if err != nil {
		return "", err
	}

	file, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	resp, err := cld.Upload.Upload(context.Background(), file, uploader.UploadParams{
		Folder:       folder,
		ResourceType: "auto",
	})
	log.Printf("Upload response: %+v\n", resp)
	if err != nil {
		log.Println("Upload error:", err)
	}

	return resp.SecureURL, nil
}
