package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AzureClientID     string
	AzureClientSecret string
	AzureTenantID     string
	UploadDirectory   string
	DriveId           string
}

func LoadConfig() Config {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return Config{
		AzureClientID:     os.Getenv("AZURE_CLIENT_ID"),
		AzureClientSecret: os.Getenv("AZURE_CLIENT_SECRET"),
		AzureTenantID:     os.Getenv("AZURE_TENANT_ID"),
		UploadDirectory:   os.Getenv("UPLOAD_DIRECTORY"),
		DriveId:           os.Getenv("DRIVE_ID"),
	}
}
