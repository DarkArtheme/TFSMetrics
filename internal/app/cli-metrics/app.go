package cli_metrics

import (
	"errors"
	"fmt"
	"go-marathon-team-3/pkg/tfsmetrics"
	"go-marathon-team-3/pkg/tfsmetrics/azure"
	"go-marathon-team-3/pkg/tfsmetrics/repointerface"

	"github.com/urfave/cli/v2"

	"io/ioutil"
	"os"
	"path"
	"time"

	"gopkg.in/yaml.v3"
)

func CreateMetricsApp(prjPath string) *cli.App {
	var azureClient *azure.Azure
	app := cli.NewApp()
	app.Name = "cli-metrics"
	app.Usage = "CLI для взаимодействия с библиотекой"
	//app.Action = func(c *cli.Context) error {
	//	fmt.Println("Hello, team 3!")
	//	return nil
	//}
	app.EnableBashCompletion = true
	app.Version = "0.3"
	app.Authors = []*cli.Author {
		{ Name: "Андрей Назаренко" },
		{ Name: "Артем Богданов" },
		{ Name: "Алексей Вологдин" },
	}
	var url string
	var token string
	app.Commands = []*cli.Command{
		{
			Name: "config",
			Aliases: []string{},
			Usage: "установка параметров, необходимых для подключения к Azure",
			Flags: []cli.Flag {
				&cli.StringFlag {
					Name: "organization-url",
					Aliases: []string{"url", "u"},
					Usage: "url для подключения к Azure",
					Destination: &url,
				},
				&cli.StringFlag{
					Name:        "access-token",
					Aliases:     []string{"token", "t"},
					Usage:       "personal access token для подключения к Azure",
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
			Name:    "log",
			Aliases: []string{},
			Usage:   "получение информации обо всех коммитах",
			Action: func(context *cli.Context) error {
				var err error
				prjName := context.Args().Get(0)
				azureClient, err = connect(prjPath)
				if err != nil {
					return err
				}
				projectNames, err := azureClient.ListOfProjects()
				if err != nil {
					return err
				}
				if prjName == "" {
					fmt.Println("Название проекта не было указано, информация по коммитам будет выведена по всем проектам:\n")
					for _, project := range projectNames {
						fmt.Printf("\t\t\t\t\t\tПроект %s:\n\n\n", *project)
						commits := tfsmetrics.NewCommitCollection(*project, azureClient, false, nil)
						iter, err := commits.GetCommitIterator()
						if err != nil {
							return err
						}
						for commit, err := iter.Next(); err == nil; commit, err = iter.Next() {
							printFullCommit(commit)
						}
					}
				} else {
					for _, project := range projectNames {
						if *project == prjName{
							fmt.Printf("\t\t\tПроект %s:\n\n\n", *project)
							commits := tfsmetrics.NewCommitCollection(*project, azureClient, false, nil)
							iter, err := commits.GetCommitIterator()
							if err != nil {
								return err
							}
							for commit, err := iter.Next(); err == nil; commit, err = iter.Next() {
								printFullCommit(commit)
							}
							break
						}
					}
				}
				commits := getCommits()
				for _, commit := range *commits {
					printFullCommit(&commit)
				}
				return nil
			},
		},
		{
			Name: "list",
			Aliases: []string{},
			Usage: "вывод на экран названий всех проектов в репозитории",
			Action: func(context *cli.Context) error {
				var err error
				azureClient, err = connect(prjPath)
				if err != nil {
					return err
				}
				projectNames, err := azureClient.ListOfProjects()
				if err != nil {
					return err
				}
				fmt.Println("Доступны следующие проекты:")
				for ind, project := range projectNames {
					fmt.Printf("%d) %s\n",ind + 1, *project)
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
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
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

func printFullCommit(commit *repointerface.Commit) {
	fmt.Printf("Author: %s <%s>\n", commit.Author, commit.Email)
	fmt.Printf("Date: %s\n", commit.Date.Format("2006-01-02 15:04:05"))
	fmt.Printf("%d rows added and %d rows deleted\n", commit.AddedRows, commit.DeletedRows)
	fmt.Printf("Commit message:\n\n\t%s\n\n", commit.Message)
}

// Эмуляция получения коммитов(ченджсетов). Будет удалена.
func getCommits() *[]repointerface.Commit {
	n := 10
	commits := make([]repointerface.Commit, 0, n)
	for i := 0; i < n; i++ {
		commits = append(commits, repointerface.Commit{Author: "Author's Name",
			Email: "testemail@gmail.com", AddedRows: 58, DeletedRows: 7,
			Date:    time.Date(2020, time.Month(i), i*2, i+1, i*3, 0, 0, time.UTC),
			Message: "Commit message", Hash: "2e4ca12s"})
	}
	return &commits
}

func connect(prjPath string) (*azure.Azure, error) {
	filePath := path.Join(prjPath, "configs/config.yaml")
	config, err := ReadConfigFile(filePath)
	if config.OrganizationUrl == "" && config.Token == "" {
		return nil, errors.New("отсутствуют параметры подключения (cli-metrics config)")
	} else if config.OrganizationUrl == "" {
		return nil, errors.New("отсутствует url подключения (cli-metrics config --url)")
	} else if config.Token == "" {
		return nil, errors.New("отсутствует token подключения (cli-metrics config --token)")
	}
	azureClient := azure.NewAzure(config)
	azureClient.Connect()
	err = azureClient.TfvcClientConnection()
	if err != nil {
		return nil, err
	}
	return azureClient, err
}

