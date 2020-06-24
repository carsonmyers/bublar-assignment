package players

import (
	"flag"
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
)

var getOpts struct {
	user string
}

func getCommand() *command.Command {
	flagSet := flag.NewFlagSet("get", flag.ExitOnError)
	flagSet.StringVar(&getOpts.user, "u", "", "Username")

	return command.New("get", "Get a user's data", flagSet, runGet)
}

func runGet(cmd *command.Command) error {
	api := connect.API()

	var req *connect.Request
	if len(getOpts.user) == 0 {
		req, _ = api.NewRequest("GET", api.URL("/client/player"), nil)
	} else {
		url := api.URL(fmt.Sprintf("/client/players/%s", getOpts.user))
		req, _ = api.NewRequest("GET", url, nil)
	}

	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
