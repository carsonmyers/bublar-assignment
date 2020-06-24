package players

import (
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
)

func listCommand() *command.Command {
	return command.New("list", "List all players", nil, runList)
}

func runList(cmd *command.Command) error {
	api := connect.API()

	req, _ := api.NewRequest("GET", api.URL("/client/players"), nil)
	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
