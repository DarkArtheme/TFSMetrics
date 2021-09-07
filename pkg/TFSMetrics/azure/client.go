package azure

import (
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/tfvc"
)

type AzureClientInterface interface {
	Connect(config config) error
	GetChangesets()
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

	coreClient, err := tfvc.NewClient(ctx, connection)
	if err != nil {
		return err
	}
	ac.Client = &coreClient
	return nil
}
