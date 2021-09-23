package cli_metrics

import (
	"encoding/json"
	"errors"
	"fmt"
	"go-marathon-team-3/pkg/tfsmetrics"
	"go-marathon-team-3/pkg/tfsmetrics/azure"
	"go-marathon-team-3/pkg/tfsmetrics/repointerface"
	"go-marathon-team-3/pkg/tfsmetrics/store"

	"github.com/urfave/cli/v2"

	"io/ioutil"
	"os"
	"path"
)

type cliSettings struct {
	CacheEnabled bool `json:"cache-enabled"`
}

func CreateMetricsApp(prjPath *string) *cli.App {
	app := cli.NewApp()
	app.Name = "cli-metrics"
	app.Usage = "CLI для взаимодействия с библиотекой"
	//app.Action = func(c *cli.Context) error {
	//	fmt.Println("Hello, team 3!")
	//	return nil
	//}
	app.EnableBashCompletion = true
	app.Version = "0.5"
	app.Authors = []*cli.Author{
		{Name: "Андрей Назаренко"},
		{Name: "Артем Богданов"},
		{Name: "Алексей Вологдин"},
	}
	settingsPath := path.Join(*prjPath, "configs/cli-settings.json")
	settings, _ := ReadSettingsFile(&settingsPath)
	localStore, _ := store.NewStore()
	var url string
	var token string
	var cache string
	app.Commands = []*cli.Command {
		{
			Name:    "config",
			Aliases: []string{},
			Usage:   "установка параметров, необходимых для подключения к Azure",
			Flags: []cli.Flag{
				&cli.StringFlag {
					Name:        "organization-url",
					Aliases:     []string{"url", "u"},
					Usage:       "url для подключения к Azure",
					Destination: &url,
				},
				&cli.StringFlag {
					Name:        "access-token",
					Aliases:     []string{"token", "t"},
					Usage:       "personal access token для подключения к Azure",
					Destination: &token,
				},
				&cli.StringFlag {
					Name: "cache-enabled",
					Aliases: []string{"cache", "c"},
					Usage: "логический флаг следует ли использовать кеш при работе программы",
					Destination: &cache,
				},
			},
			Action: func(c *cli.Context) error {
				configPath := path.Join(*prjPath, "configs/config.json")
				config, err := ReadConfigFile(&configPath)
				if err != nil {
					return err
				}
				if url != "" {
					config.OrganizationUrl = url
				}
				if token != "" {
					config.Token = token
				}
				if cache == "true"  {
					settings.CacheEnabled = true
				} else if cache != "" {
					settings.CacheEnabled = false
				}
				err = WriteConfigFile(&configPath, config)
				err = WriteSettingsFile(&settingsPath, settings)
				fmt.Printf("Current config:\nURL: %s\nToken: %s\nCacheEnabled: %t\n", config.OrganizationUrl,
					config.Token, settings.CacheEnabled)
				return err
			},
		},
		{
			Name:	"log",
			Aliases: []string{},
			Usage:   "получение информации обо всех коммитах",
			Action: func(context *cli.Context) error {
				var err error
				prjName := context.Args().Get(0)
				azureClient, err := connect(prjPath)
				settings, _ := ReadSettingsFile(&settingsPath)
				if err != nil {
					return err
				}
				projectNames, err := azureClient.ListOfProjects()
				if err != nil {
					return err
				}
				if prjName == "" {
					fmt.Println("Название проекта не было указано, информация по коммитам будет выведена по всем проектам:")
					for _, project := range projectNames {
						_ = processProject(project, &azureClient, settings.CacheEnabled, &localStore)
					}
				} else {
					for _, project := range projectNames {
						if *project == prjName {
							err = processProject(project, &azureClient, settings.CacheEnabled, &localStore)
							if err != nil {
								return err
							}
						}
					}
				}
				return nil
			},
		},
		{
			Name:    "list",
			Aliases: []string{},
			Usage:   "вывод на экран названий всех проектов в репозитории",
			Action: func(context *cli.Context) error {
				var err error
				azureClient, err := connect(prjPath)
				if err != nil {
					return err
				}
				projectNames, err := azureClient.ListOfProjects()
				if err != nil {
					return err
				}
				fmt.Println("Доступны следующие проекты:")
				for ind, project := range projectNames {
					fmt.Printf("%d) %s\n", ind+1, *project)
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

func ReadConfigFile(filePath *string) (config *azure.Config, err error) {
	config = azure.NewConfig()
	ex, _ := exists(*filePath)
	if !ex {
		output, _ := os.Create(*filePath)
		defer output.Close()
		jsonEncoder := json.NewEncoder(output)
		err = jsonEncoder.Encode(config)
	}
	data, err := ioutil.ReadFile(*filePath)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &config)
	return
}

func WriteConfigFile(filePath *string, config *azure.Config) error {
	output, err := os.Create(*filePath)
	if err != nil {
		return err
	}
	defer output.Close()
	jsonEncoder := json.NewEncoder(output)
	err = jsonEncoder.Encode(config)
	return err
}

func WriteSettingsFile(filePath *string, settings *cliSettings) error {
	output, err := os.Create(*filePath)
	if err != nil {
		return err
	}
	defer output.Close()
	jsonEncoder := json.NewEncoder(output)
	err = jsonEncoder.Encode(settings)
	return err
}

func ReadSettingsFile(filePath *string) (settings *cliSettings, err error) {
	settings = &cliSettings{CacheEnabled: true}
	ex, _ := exists(*filePath)
	if !ex {
		output, _ := os.Create(*filePath)
		defer output.Close()
		jsonEncoder := json.NewEncoder(output)
		err = jsonEncoder.Encode(settings)
		return
	}
	data, err := ioutil.ReadFile(*filePath)
	if err != nil {
		return
	}
	err = json.Unmarshal(data, &settings)
	return
}

func printFullCommit(commit *repointerface.Commit) {
	fmt.Printf("Author: %s <%s>\n", commit.Author, commit.Email)
	fmt.Printf("Date: %s\n", commit.Date.Format("2006-01-02 15:04:05"))
	fmt.Printf("%d rows added and %d rows deleted\n", commit.AddedRows, commit.DeletedRows)
	fmt.Printf("Commit message:\n\n\t%s\n\n", commit.Message)
	fmt.Println("---------------------------------------------------------------------------------------------------")
}

func printProjectName(name *string) {
	fmt.Println("---------------------------------------------------------------------------------------------------")
	fmt.Printf("\t\t\tПроект %s:\n", *name)
	fmt.Println("---------------------------------------------------------------------------------------------------")
	fmt.Printf("\n\n")
}

func processProject(project *string, azureClient *azure.AzureInterface, cacheEnabled bool, localStore *store.Store) error {
	printProjectName(project)
	commits := tfsmetrics.NewCommitCollection(*project, *azureClient, cacheEnabled, *localStore)
	err := commits.Open()
	if err != nil {
		return err
	}
	iter, err := commits.GetCommitIterator()
	if err != nil {
		return err
	}
	for commit, err := iter.Next(); err == nil; commit, err = iter.Next() {
		printFullCommit(commit)
	}
	return nil
}


func connect(prjPath *string) (azure.AzureInterface, error) {
	filePath := path.Join(*prjPath, "configs/config.json")
	config, err := ReadConfigFile(&filePath)
	if err != nil {
		return nil, err
	}
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
