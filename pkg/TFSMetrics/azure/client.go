package azure

import (
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/tfvc"
)

type AzureClientInterface interface {
	Connect(config config) error
	//Получает все ченджсеты
	GetChangesets() 
	//Получаем изменения для отдельного ченджсета.
	//В нём есть HashValue для структуры из тз
	GetChangesetChanges()
	//Получение нового файла из ченджсета
	GetCurrentFile()
	//Получение старого файла из ченджсета
	GetPreviousFile()
	//Получение изменений (добавленные и удаленные строки)
	GetFileChanges()
}

type Client struct {
	Client *tfvc.Client
}

func NewClient() *Client {
	return &Client{}
}

func (ac *Client) Connect(config config) error {
	organizationUrl := config.OrganizationUrl
	personalAccessToken := config.Token

	connection := azuredevops.NewPatConnection(organizationUrl, personalAccessToken)

	ctx := config.Context

	tfvcClient, err := tfvc.NewClient(ctx, connection)
	if err != nil {
		return err
	}
	ac.Client = &tfvcClient
	return nil
}
