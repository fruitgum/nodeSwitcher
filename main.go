package main

import (
	"failover/app"
	"failover/config"
	"flag"
	"fmt"
)

func main() {
	configPath := flag.String("config", "/etc/failoverd/config.yaml", "Path to config file")
	action := flag.String("action", "healthcheck", "healthcheck or failover")
	state := flag.String("state", "standby", "active or standby")
	dryRun := flag.Bool("dryRun", false, "dry run")
	flag.Parse()

	config.ConfigPath = *configPath

	app.Action = *action
	app.State = *state
	app.DryRun = *dryRun
	a, err := app.New()
	if err != nil {
		fmt.Println(err)
		a.Logs.Fatal(err.Error())
		return
	}
	a.Router()
}
