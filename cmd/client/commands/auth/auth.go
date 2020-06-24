package auth

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/carsonmyers/bublar-assignment/cmd/client/command"
	"github.com/carsonmyers/bublar-assignment/connect"
	"github.com/carsonmyers/bublar-assignment/data"
	"github.com/gbrlsnchs/jwt"
	"github.com/mitchellh/go-homedir"
	"go.uber.org/zap"
	"golang.org/x/crypto/ssh/terminal"
)

var authOpts struct {
	user string
	pass string
}

// LoginCommand - log in to the API
func LoginCommand() *command.Command {
	flagSet := flag.NewFlagSet("login", flag.ExitOnError)
	flagSet.StringVar(&authOpts.user, "u", "", "Username")
	flagSet.StringVar(&authOpts.pass, "p", "", "Password (will prompt if omitted)")

	return command.New("login", "Login to a licensing service", flagSet, runLogin)
}

// LogoutCommand - log out of the API
func LogoutCommand() *command.Command {
	return command.New("logout", "Log out from the licensing service", nil, runLogout)
}

func runLogin(cmd *command.Command) error {
	if len(authOpts.pass) == 0 {
		fmt.Print("Password: ")
		passwdBytes, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return err
		}
		fmt.Println()

		authOpts.pass = string(passwdBytes)
	}

	api := connect.API()

	req, log := api.NewRequest("POST", api.URL("/client/login"), &data.Player{
		Username: authOpts.user,
		Password: &authOpts.pass,
	})

	res, _, err := req.Do()
	if err != nil {
		return err
	}

	cookies := res.Cookies()
	for _, cookie := range cookies {
		log.Debug("Processing cookie", zap.String("name", cookie.Name), zap.String("value", cookie.Value))
		if cookie.Name != "AUTH" {
			continue
		}

		if err := writeSession(cookie.Value); err != nil {
			log.Error("Failed to save session", zap.Error(err))
			return err
		}

		log.Info("Session saved", zap.String("username", authOpts.user))
		return nil
	}

	log.Error("Endpoint did not return auth token")
	return errors.New("auth cookie not found")
}

func runLogout(cmd *command.Command) error {
	session, err := ReadSession()
	if err != nil {
		return err
	}

	if len(session) == 0 {
		return nil
	}

	return rmSession()
}

func sessionFile() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ".client-session"), nil
}

// ReadSession - load the session file from disk
func ReadSession() (string, error) {
	filename, err := sessionFile()
	if err != nil {
		return "", err
	}

	sessBytes, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}

		return "", err
	}

	return string(sessBytes), nil
}

// CurrentUser - decode the session file and read the current username
func CurrentUser() (string, error) {
	sess, err := ReadSession()
	if err != nil {
		return "", err
	}

	if len(sess) == 0 {
		return "", nil
	}

	decoded, err := base64.StdEncoding.DecodeString(sess)
	if err != nil {
		return "", err
	}

	var token *jwt.JWT
	err = json.Unmarshal(decoded, &token)
	if err != nil {
		return "", err
	}

	return token.Audience(), nil
}

func writeSession(token string) error {
	filename, err := sessionFile()
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, []byte(token), os.ModePerm)
}

func rmSession() error {
	filename, err := sessionFile()
	if err != nil {
		return err
	}

	return os.Remove(filename)
}
