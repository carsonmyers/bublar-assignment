package locations

import (
	"flag"
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
)

var getOpts struct {
	name string
}

func getCommand() *command.Command {
	flagSet := flag.NewFlagSet("get", flag.ExitOnError)
	flagSet.StringVar(&getOpts.name, "n", "", "Location name")

	return command.New("get", "Get a location's details", flagSet, runGet)
}

func runGet(cmd *command.Command) error {
	api := connect.API()

	url := api.URL(fmt.Sprintf("/client/locations/%s", getOpts.name))
	req, _ := api.NewRequest("GET", url, nil)
	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
