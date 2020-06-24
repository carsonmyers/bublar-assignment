package locations

import (
	"flag"
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
	"github.com/carsonmyers/bublar-assignment/data"
)

var updateOpts struct {
	name    string
	newName string
	newX    int
	newY    int
}

func updateCommand() *command.Command {
	flagSet := flag.NewFlagSet("update", flag.ExitOnError)
	flagSet.StringVar(&updateOpts.name, "n", "", "Location name")
	flagSet.StringVar(&updateOpts.newName, "nn", "", "New location name")
	flagSet.IntVar(&updateOpts.newX, "x", 0, "New X-position for location")
	flagSet.IntVar(&updateOpts.newY, "y", 0, "New Y-position for location")

	return command.New("update", "Update a location's details", flagSet, runUpdate)
}

func runUpdate(cmd *command.Command) error {
	api := connect.API()

	url := api.URL(fmt.Sprintf("/admin/locations/%s", updateOpts.name))
	req, _ := api.NewRequest("PATCH", url, &data.Location{
		Name: updateOpts.newName,
		X:    updateOpts.newX,
		Y:    updateOpts.newY,
	})

	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
