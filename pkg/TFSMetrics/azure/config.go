package azure

import "context"

type Config struct {
	OrganizationUrl string `yaml:"organization_url"`
	Token           string `yaml:"personal_access_token"`
	Context         context.Context `yaml:"-"`
}

func NewConfig() *Config {
	return &Config{
		Context: context.Background(),
		OrganizationUrl: "",
		Token: "",
	}
}
