package players

import (
	"flag"
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
)

var deleteOpts struct {
	user string
}

func deleteCommand() *command.Command {
	flagSet := flag.NewFlagSet("delete", flag.ExitOnError)
	flagSet.StringVar(&deleteOpts.user, "u", "", "Username")

	return command.New("delete", "Delete a player", flagSet, runDelete)
}

func runDelete(cmd *command.Command) error {
	api := connect.API()

	url := api.URL(fmt.Sprintf("/admin/players/%s", deleteOpts.user))
	req, _ := api.NewRequest("DELETE", url, nil)
	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
