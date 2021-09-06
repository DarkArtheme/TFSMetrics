package main

import (
	"fmt"
	"github.com/urfave/cli/v2" // imports as package "cli"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		Name: "cliMetrics",
		Usage: "CLI to interact with the lib",
		Action: func(c *cli.Context) error {
			fmt.Println("Hello, team 3!")
			return nil
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

