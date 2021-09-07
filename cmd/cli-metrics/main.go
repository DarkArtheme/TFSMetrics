package main

import (
	"go-marathon-team-3/internal/app/cli-metrics"
	"log"
	"os"
)

func main() {
	app := cli_metrics.CreateMetricsApp()
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

