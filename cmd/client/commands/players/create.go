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

var createOpts struct {
	user string
	pass string
}

func createCommand() *command.Command {
	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	flagSet.StringVar(&createOpts.user, "u", "", "Username")
	flagSet.StringVar(&createOpts.pass, "p", "", "Password (will prompt if omitted)")

	return command.New("create", "Create a new player", flagSet, runCreate)
}

func runCreate(cmd *command.Command) error {
	if len(createOpts.pass) == 0 {
		fmt.Print("Password: ")
		passwdBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Println()

		createOpts.pass = string(passwdBytes)
	}

	api := connect.API()

	req, _ := api.NewRequest("POST", api.URL("/client/players"), &data.Player{
		Username: createOpts.user,
		Password: &createOpts.pass,
	})

	_, output, err := req.Do()
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}
