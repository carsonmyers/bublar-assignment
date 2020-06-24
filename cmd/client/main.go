package main

import (
	"fmt"
	"os"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/cmd/client/commands/auth"
	"github.com/carsonmyers/bublar-assignment/cmd/client/commands/locations"
	"github.com/carsonmyers/bublar-assignment/cmd/client/commands/players"
	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/carsonmyers/bublar-assignment/logger"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

var log = logger.GetLogger()

type config struct {
	API *configure.APIConfig
}

var defaultConfig = config{
	API: &configure.DefaultAPIConfig,
}

func main() {
	config := defaultConfig
	envconfig.MustProcess("api", config.API)

	session, err := auth.ReadSession()
	if err != nil {
		log.Error("Error reading session file", zap.Error(err))
	}

	if len(session) > 0 {
		config.API.Session = session
	}

	configure.API(config.API)

	cmd := command.Init("client", "Licensing administration tool", nil, run)
	cmd.AddCommand(auth.LoginCommand())
	cmd.AddCommand(auth.LogoutCommand())
	cmd.AddCommand(players.Command())
	cmd.AddCommand(locations.Command())

	if err := cmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *command.Command) error {
	next, err := cmd.Next()
	if err != nil {
		return err
	}

	if next == nil {
		fmt.Print(cmd.Help())
		return nil
	}

	return next.Execute()
}
