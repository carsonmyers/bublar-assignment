package players

import (
	"flag"
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
)

var travelOpts struct {
	user     string
	location string
}

func travelCommand() *command.Command {
	flagSet := flag.NewFlagSet("travel", flag.ExitOnError)
	flagSet.StringVar(&travelOpts.user, "u", "", "Username (current user if omitted)")
	flagSet.StringVar(&travelOpts.location, "l", "", "Location")

	return command.New("travel", "Travel to a new location", flagSet, runTravel)
}

func runTravel(cmd *command.Command) error {
	api := connect.API()

	body := struct {
		Location string `json:"location"`
	}{
		Location: travelOpts.location,
	}

	var req *connect.Request
	if len(travelOpts.user) == 0 {
		req, _ = api.NewRequest("POST", api.URL("/client/player/travel"), &body)
	} else {
		url := api.URL(fmt.Sprintf("/admin/players/%s/travel", travelOpts.user))
		req, _ = api.NewRequest("POST", url, &body)
	}

	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
