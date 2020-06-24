package locations

import (
	"flag"
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
	"github.com/carsonmyers/bublar-assignment/data"
)

var createOpts struct {
	name string
	x    int
	y    int
}

func createCommand() *command.Command {
	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	flagSet.StringVar(&createOpts.name, "n", "", "Location name")
	flagSet.IntVar(&createOpts.x, "x", 0, "X-position of location")
	flagSet.IntVar(&createOpts.y, "y", 0, "Y-position of location")

	return command.New("create", "Create a new location", flagSet, runCreate)
}

func runCreate(cmd *command.Command) error {
	api := connect.API()

	req, _ := api.NewRequest("POST", api.URL("/admin/locations"), &data.Location{
		Name: createOpts.name,
		X:    createOpts.x,
		Y:    createOpts.y,
	})

	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}
