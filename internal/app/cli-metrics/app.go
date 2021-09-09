package cli_metrics

import (
	"fmt"
	"github.com/urfave/cli/v2"
	"go-marathon-team-3/pkg/TFSMetrics/azure"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
)

func CreateMetricsApp(prjPath string) *cli.App {
	app := cli.NewApp()
	app.Name = "cli-metrics"
	app.Usage = "CLI для взаимодействия с библиотекой"
	app.Action = func(c *cli.Context) error {
		fmt.Println("Hello, team 3!")
		return nil
	}
	app.Version = "0.01"
	app.Authors = []*cli.Author {
		{ Name: "Андрей Назаренко" },
		{ Name: "Артем Богданов" },
		{ Name: "Василий Грязных" },
		{ Name: "Алексей Вологдин" },
	}
	var url string
	var token string
	app.Commands = []*cli.Command {
		{
			Name: "config",
			Aliases: []string{"c"},
			Usage: "установка параметров, необходимых для подключения к Azure",
			Flags: []cli.Flag {
				&cli.StringFlag {
					Name: "organization-url",
					Aliases: []string{"url", "u"},
					Usage: "url для подключения к Azure",
					Destination: &url,
				},
				&cli.StringFlag {
					Name: "access-token",
					Aliases: []string{"token", "t"},
					Usage: "personal access token для подключения к Azure",
					Destination: &token,
				},
			},
			Action: func(c *cli.Context) error {
				filePath := path.Join(prjPath, "configs/config.yaml")
				config, err := ReadConfigFile(filePath)
				if err != nil {
					return err
				}
				if url != "" {
					config.OrganizationUrl = url
				}
				if token != "" {
					config.Token = token
				}
				err = WriteConfigFile(filePath, config)
				fmt.Printf("Current config:\n\nURL: %s\nToken: %s\n", config.OrganizationUrl, config.Token)
				return err
			},
		},
	}
	azure.NewConfig()
	return app
}


func ReadConfigFile(filePath string) (config *azure.Config, err error) {
	config = azure.NewConfig()
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(data, &config)
	return
}

func WriteConfigFile(filePath string, config *azure.Config) error {
	output, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer output.Close()
	yamlEncoder := yaml.NewEncoder(output)
	err = yamlEncoder.Encode(config)
	return err
}