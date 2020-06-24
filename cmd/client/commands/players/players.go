package players

import (
	"fmt"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
)

// Command - players subcommand
func Command() *command.Command {
	cmd := command.New("players", "Manage players", nil, run)
	cmd.AddCommand(createCommand())
	cmd.AddCommand(getCommand())
	cmd.AddCommand(listCommand())
	cmd.AddCommand(updateCommand())
	cmd.AddCommand(moveCommand())
	cmd.AddCommand(travelCommand())
	cmd.AddCommand(deleteCommand())

	return cmd
}

func run(cmd *command.Command) error {
	next, err := cmd.Next()
	if err != nil {
		return err
	}

	if next == nil {
		fmt.Print(cmd.Help())
		return nil
	}

	return next.Execute()
}
