package players

import (
	"flag"
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
)

var moveOpts struct {
	user string
	x    int
	y    int
}

func moveCommand() *command.Command {
	flagSet := flag.NewFlagSet("move", flag.ExitOnError)
	flagSet.StringVar(&moveOpts.user, "u", "", "Username (current user if omitted)")
	flagSet.IntVar(&moveOpts.x, "x", 0, "X-position within location")
	flagSet.IntVar(&moveOpts.y, "y", 0, "Y-position within location")

	return command.New("move", "Move to a new position withing the current location", flagSet, runMove)
}

func runMove(cmd *command.Command) error {
	api := connect.API()

	body := struct {
		X int `json:"x"`
		Y int `json:"y"`
	}{
		X: moveOpts.x,
		Y: moveOpts.y,
	}

	var req *connect.Request
	if len(moveOpts.user) == 0 {
		req, _ = api.NewRequest("POST", api.URL("/client/player/move"), &body)
	} else {
		url := api.URL(fmt.Sprintf("/admin/players/%s/move", moveOpts.user))
		req, _ = api.NewRequest("POST", url, &body)
	}

	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
