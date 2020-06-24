package players

import (
	"flag"
	"fmt"
	"syscall"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
	"github.com/carsonmyers/bublar-assignment/data"
	"golang.org/x/crypto/ssh/terminal"
)

var updateOpts struct {
	user        string
	newUser     string
	newPassword string
}

func updateCommand() *command.Command {
	flagSet := flag.NewFlagSet("update", flag.ExitOnError)
	flagSet.StringVar(&updateOpts.user, "u", "", "Username")
	flagSet.StringVar(&updateOpts.newUser, "nu", "", "New username")
	flagSet.StringVar(&updateOpts.newPassword, "p", "", "New password (will prompt if omitted)")

	return command.New("update", "Update a player's details", flagSet, runUpdate)
}

func runUpdate(cmd *command.Command) error {
	if len(updateOpts.newPassword) == 0 {
		fmt.Print("Password: ")
		passwdBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Println()

		createOpts.pass = string(passwdBytes)
	}

	body := &data.Player{
		Username: updateOpts.newUser,
		Password: &updateOpts.newPassword,
	}

	api := connect.API()

	var req *connect.Request
	if len(updateOpts.user) == 0 {
		req, _ = api.NewRequest("PATCH", api.URL("/client/player"), body)
	} else {
		url := api.URL(fmt.Sprintf("/admin/players/%s", updateOpts.user))
		req, _ = api.NewRequest("PATCH", url, body)
	}

	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)
	return nil
}
