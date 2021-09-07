package azure

import (
	"github.com/microsoft/azure-devops-go-api/azuredevops"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
)

type AzureClientInterface interface {
	Connect(config config) error
}

type Client struct {
	Client *core.Client
}

func NewClient() *Client {
	return &Client{}
}

func (ac *Client) Connect(config config) error {
	organizationUrl := config.OrganizationUrl
	personalAccessToken := config.Token

	connection := azuredevops.NewPatConnection(organizationUrl, personalAccessToken)

	ctx := config.Context

	coreClient, err := core.NewClient(ctx, connection)
	if err != nil {
		return err
	}
	ac.Client = &coreClient
	return nil
}
