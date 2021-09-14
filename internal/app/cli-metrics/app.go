package cli_metrics

import (
	"errors"
	"fmt"
	"github.com/urfave/cli/v2"
	tfsmetrics "go-marathon-team-3/pkg/TFSMetrics"
	"go-marathon-team-3/pkg/TFSMetrics/azure"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
	"path"
	"time"
)

func CreateMetricsApp(prjPath string) *cli.App {
	var azureClient *azure.Azure
	isConnected := false
	app := cli.NewApp()
	app.Name = "cli-metrics"
	app.Usage = "CLI для взаимодействия с библиотекой"
	//app.Action = func(c *cli.Context) error {
	//	fmt.Println("Hello, team 3!")
	//	return nil
	//}
	app.EnableBashCompletion = true
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
				fmt.Printf("Current config:\nURL: %s\nToken: %s\n", config.OrganizationUrl, config.Token)
				return err
			},
		},
		{
			Name: "log",
			Aliases: []string{},
			Usage: "получение информации обо всех коммитах",
			Action: func(context *cli.Context) error {
				commits := getCommits()
				for _, commit := range *commits {
					printFullCommit(&commit)
				}
				return nil
			},
		},
		{
			Name: "connect",
			Aliases: []string{},
			Usage: "подключение к TFS-репозиторию",
			Action: func(context *cli.Context) error {
				filePath := path.Join(prjPath, "configs/config.yaml")
				config, err := ReadConfigFile(filePath)
				if err != nil {
					return err
				}
				if config.OrganizationUrl == "" && config.Token == "" {
					return errors.New("отсутствуют параметры подключения (cli-metrics config)")
				} else if config.OrganizationUrl == "" {
					return errors.New("отсутствует url подключения (cli-metrics config --url)")
				} else if config.Token == "" {
					return errors.New("отсутствует token подключения (cli-metrics config --token)")
				}
				azureClient = azure.NewAzure(config)
				azureClient.Connect()
				err = azureClient.TfvcClientConnection()
				if err != nil {
					return err
				}
				isConnected = true
				fmt.Println("Успешное подключение")
				projectNames, err := azureClient.ListOfProjects()
				fmt.Println("Доступны следующие проекты:")
				for _, project := range projectNames {
					fmt.Println(*project)
				}
				return nil
			},
		},
	}
	azure.NewConfig()
	return app
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return false, err
}

func ReadConfigFile(filePath string) (config *azure.Config, err error) {
	config = azure.NewConfig()
	ex, _ := exists(filePath)
	if !ex {
		output, _ := os.Create(filePath)
		defer output.Close()
		yamlEncoder := yaml.NewEncoder(output)
		err = yamlEncoder.Encode(config)
	}
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

func printFullCommit(commit *tfsmetrics.Commit) {
	fmt.Printf("Author: %s <%s>\n", commit.Author, commit.Email)
	fmt.Printf("Date: %s\n", commit.Date.Format("2006-01-02 15:04:05"))
	fmt.Printf("%d rows added and %d rows deleted\n", commit.AddedRows, commit.DeletedRows)
	fmt.Printf("Commit message:\n\n\t%s\n\n", commit.Message)
}

// Эмуляция получения коммитов(ченджсетов). Будет удалена.
func getCommits() *[]tfsmetrics.Commit {
	n := 10
	commits := make([]tfsmetrics.Commit, 0, n)
	for i := 0; i < n; i++ {
		commits = append(commits, tfsmetrics.Commit{Author: "Author's Name",
			Email: "testemail@gmail.com", AddedRows: 58, DeletedRows: 7,
			Date: time.Date(2020, time.Month(i), i*2, i+1, i*3, 0, 0, time.UTC),
			Message: "Commit message", Hash: "2e4ca12s" })
	}
	return &commits
}