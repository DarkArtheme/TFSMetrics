package azure

import "context"

type config struct {
	OrganizationUrl string
	Token           string
	Context         context.Context
}

func NewConfig() *config {
	return &config{
		Context: context.Background(),
	}
}
