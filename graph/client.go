package graph

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	auth "github.com/microsoft/kiota-authentication-azure-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
)

type GraphClient struct {
	Client *msgraphsdk.GraphServiceClient
}

func NewGraphClient(clientID, clientSecret, tenantID string) *GraphClient {
	cred, err := azidentity.NewClientSecretCredential(tenantID, clientID, clientSecret, nil)
	if err != nil {
		log.Fatalf("Failed to create credential: %v", err)
	}

	// Создаем auth provider с нужными правами
	authProvider, err := auth.NewAzureIdentityAuthenticationProviderWithScopes(cred, []string{
		"https://graph.microsoft.com/.default",
	})
	if err != nil {
		log.Fatalf("Failed to create auth provider: %v", err)
	}

	// Создаем адаптер с правильными сериализаторами
	adapter, err := msgraphsdk.NewGraphRequestAdapter(authProvider)
	if err != nil {
		log.Fatalf("Failed to create adapter: %v", err)
	}

	// Регистрируем сериализаторы
	adapter.SetBaseUrl("https://graph.microsoft.com/v1.0")

	// Создаем клиент
	client := msgraphsdk.NewGraphServiceClient(adapter)

	return &GraphClient{Client: client}
}
