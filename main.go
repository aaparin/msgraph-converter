package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"msgraph-converter/config"
	"msgraph-converter/graph"

	"github.com/gin-gonic/gin"
)

// getPort returns the port from environment variable or default value
func getPort() string {
	port := os.Getenv("SERVICE_PORT")
	if port == "" {
		// Логируем использование дефолтного порта
		log.Println("Warning: SERVICE_PORT not set, using default port 8181")
		return "8181"
	}
	return port
}

func main() {
	// Загружаем конфигурацию
	cfg := config.LoadConfig()

	// Инициализируем клиента Microsoft Graph
	client := graph.NewGraphClient(cfg.AzureClientID, cfg.AzureClientSecret, cfg.AzureTenantID)

	r := gin.Default()

	r.POST("/convert", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
			return
		}

		localPath := "./" + file.Filename
		if err := c.SaveUploadedFile(file, localPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		defer os.Remove(localPath) // Удаляем временный файл после обработки

		ctx := context.Background()

		// Создаем или находим папку
		folderID, err := client.GetOrCreateFolder(ctx, cfg.DriveId, cfg.UploadDirectory)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		log.Printf("Folder ID: %s", folderID)

		// Загружаем файл в OneDrive
		fileID, err := client.UploadFile(ctx, cfg.DriveId, localPath, file.Filename, folderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Конвертируем в PDF
		pdfReader, err := client.ConvertToPDF(ctx, cfg.DriveId, fileID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer pdfReader.Close()

		// Отправляем PDF в ответе
		outputPath := "output.pdf"
		outFile, err := os.Create(outputPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create output file"})
			return
		}
		defer outFile.Close()
		defer os.Remove(outputPath) // Удаляем PDF после отправки

		if _, err := io.Copy(outFile, pdfReader); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save PDF"})
			return
		}

		c.File(outputPath)
	})

	r.GET("/drives", func(c *gin.Context) {
		ctx := context.Background()

		drives, err := client.ListDrives(ctx)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"drives": drives,
		})
	})

	port := getPort()
	serverAddr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Starting server on port %s", port)

	if err := r.Run(serverAddr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
