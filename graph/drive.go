package graph

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"

	abstractions "github.com/microsoft/kiota-abstractions-go"
	"github.com/microsoftgraph/msgraph-sdk-go-core/fileuploader"
	"github.com/microsoftgraph/msgraph-sdk-go/drives"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	msgraphsdkmodels "github.com/microsoftgraph/msgraph-sdk-go/models"
)

func (g *GraphClient) GetOrCreateFolder(ctx context.Context, driveId string, folderName string) (string, error) {
	items, err := g.Client.Drives().
		ByDriveId(driveId).
		Items().
		ByDriveItemId("root").
		Children().
		Get(ctx, nil)
	if err != nil {
		return "", fmt.Errorf("failed to list items: %w", err)
	}

	// Ищем папку среди существующих
	for _, item := range items.GetValue() {
		if item.GetName() != nil && *item.GetName() == folderName {
			return *item.GetId(), nil
		}
	}

	// Если папка не найдена, создаём новую
	requestBody := msgraphsdkmodels.NewDriveItem()
	requestBody.SetName(&folderName)
	folder := msgraphsdkmodels.NewFolder()
	requestBody.SetFolder(folder)
	requestBody.SetAdditionalData(map[string]interface{}{
		"@microsoft.graph.conflictBehavior": "rename",
	})

	newFolder, err := g.Client.Drives().
		ByDriveId(driveId).
		Items().
		ByDriveItemId("root").
		Children().
		Post(ctx, requestBody, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create folder: %w", err)
	}

	return *newFolder.GetId(), nil
}

func (g *GraphClient) UploadFile(ctx context.Context, driveID string, localPath, fileName, folderID string) (string, error) {
	// Открываем файл
	byteStream, err := os.Open(localPath)
	if err != nil {
		log.Print(localPath)
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer byteStream.Close()

	// Настраиваем свойства загрузки
	itemUploadProperties := models.NewDriveItemUploadableProperties()
	// Указываем что делать при конфликте имен
	itemUploadProperties.SetAdditionalData(map[string]any{
		"@microsoft.graph.conflictBehavior": "replace",
	})

	// Создаем тело запроса для создания сессии загрузки
	uploadSessionRequestBody := drives.NewItemItemsItemCreateUploadSessionPostRequestBody()
	uploadSessionRequestBody.SetItem(itemUploadProperties)

	// Создаем сессию загрузки
	uploadSession, err := g.Client.Drives().
		ByDriveId(driveID).
		Items().
		ByDriveItemId(folderID+":/"+fileName+":"). // Правильный путь для создания файла
		CreateUploadSession().
		Post(ctx, uploadSessionRequestBody, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create upload session: %w", err)
	}

	// Размер чанка должен быть кратен 320 KiB
	maxSliceSize := int64(320 * 1024)

	// Создаем задачу загрузки большого файла
	fileUploadTask := fileuploader.NewLargeFileUploadTask[models.DriveItemable](
		g.Client.RequestAdapter,
		uploadSession,
		byteStream,
		maxSliceSize,
		models.CreateDriveItemFromDiscriminatorValue,
		nil,
	)

	// Создаем callback для отслеживания прогресса
	progress := func(progress int64, total int64) {
		log.Printf("Uploaded %d of %d bytes (%.2f%%)",
			progress,
			total,
			float64(progress)/float64(total)*100,
		)
	}

	// Выполняем загрузку
	uploadResult := fileUploadTask.Upload(progress)
	if !uploadResult.GetUploadSucceeded() {
		return "", fmt.Errorf("upload failed")
	}

	itemResponse := uploadResult.GetItemResponse()
	if itemResponse == nil {
		return "", fmt.Errorf("failed to get upload response")
	}

	return *itemResponse.GetId(), nil
}

func (g *GraphClient) ConvertToPDF(ctx context.Context, driveID string, fileID string) (io.ReadCloser, error) {
	// Создаем полный URL с протоколом и хостом
	baseURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/drives/%s/items/%s/content", driveID, fileID)
	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse URL: %w", err)
	}

	// Добавляем query параметр для PDF
	query := parsedURL.Query()
	query.Add("format", "pdf")
	parsedURL.RawQuery = query.Encode()

	log.Printf("Making request to: %s", parsedURL.String())

	// Формируем запрос
	requestInfo := abstractions.NewRequestInformation()
	requestInfo.SetUri(*parsedURL)
	requestInfo.Method = abstractions.GET
	requestInfo.Headers.Add("Accept", "application/pdf")

	// Выполняем запрос
	response, err := g.Client.GetAdapter().SendPrimitive(ctx, requestInfo, "[]byte", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to convert to PDF: %w", err)
	}

	// Преобразуем ответ в поток для чтения
	pdfContent, ok := response.([]byte)
	if !ok {
		return nil, fmt.Errorf("failed to parse PDF response")
	}

	// Проверяем сигнатуру PDF файла
	if len(pdfContent) > 4 && string(pdfContent[:4]) != "%PDF" {
		log.Printf("Warning: Response might not be a PDF. First bytes: %x", pdfContent[:min(20, len(pdfContent))])
	}

	return io.NopCloser(bytes.NewReader(pdfContent)), nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (g *GraphClient) ListDrives(ctx context.Context) ([]string, error) {
	drives, err := g.Client.Drives().Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to list drives: %w", err)
	}

	var driveIds []string
	for _, drive := range drives.GetValue() {
		driveIds = append(driveIds, *drive.GetId())
	}

	return driveIds, nil
}
