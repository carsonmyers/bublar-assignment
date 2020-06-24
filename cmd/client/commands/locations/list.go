package locations

import (
	"flag"
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
)

var listOpts struct {
	name string
}

func listCommand() *command.Command {
	flagSet := flag.NewFlagSet("list-players", flag.ExitOnError)
	flagSet.StringVar(&listOpts.name, "n", "", "Location name")

	players := command.New("players", "List players in a location", flagSet, runListPlayers)

	cmd := command.New("list", "List all locations", nil, runList)
	cmd.AddCommand(players)

	return cmd
}

func runList(cmd *command.Command) error {
	next, err := cmd.Next()
	if err != nil {
		return err
	}

	if next == nil {
		return runListLocations(cmd)
	}

	return next.Execute()
}

func runListLocations(cmd *command.Command) error {
	api := connect.API()

	req, _ := api.NewRequest("GET", api.URL("/client/locations"), nil)
	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}

func runListPlayers(cmd *command.Command) error {
	api := connect.API()

	url := api.URL(fmt.Sprintf("/client/locations/%s/players", listOpts.name))
	req, _ := api.NewRequest("GET", url, nil)
	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
