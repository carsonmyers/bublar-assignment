package locations

import (
	"flag"
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
)

var deleteOpts struct {
	name string
}

func deleteCommand() *command.Command {
	flagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	flagSet.StringVar(&deleteOpts.name, "n", "", "Location name")

	return command.New("delete", "Delete a location", flagSet, runDelete)
}

func runDelete(cmd *command.Command) error {
	api := connect.API()

	url := api.URL(fmt.Sprintf("/admin/locations/%s", deleteOpts.name))
	req, _ := api.NewRequest("DELETE", url, nil)
	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
