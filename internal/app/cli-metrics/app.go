package cli_metrics

import (
	"fmt"
	"github.com/urfave/cli/v2"
)

func CreateMetricsApp() *cli.App {
	app := cli.NewApp()
	app.Name = "cli-metrics"
	app.Usage = "CLI для взаимодействия с библиотекой"
	app.Action = func(c *cli.Context) error {
		fmt.Println("Hello, team 3!")
		return nil
	}
	app.Version = "0.01"
	app.Authors = []*cli.Author{
		{Name: "Андрей Назаренко"},
		{Name: "Артем Богданов"},
		{Name: "Василий"},
		{Name: "Алексей"},
	}
	return app
}