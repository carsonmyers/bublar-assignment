package connect

import (
	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/carsonmyers/bublar-assignment/errors"
	playersRPC "github.com/carsonmyers/bublar-assignment/players/rpc"
	"go.uber.org/zap"
)

var playersClient *playersRPC.Client

// Players - connect to the players rpc service
func Players() (*playersRPC.Client, error) {
	if playersClient != nil {
		return playersClient, nil
	}

	config := configure.GetPlayers()
	log.Info("Connecting to Players RPC", zap.String("config", config.String()))

	client, err := playersRPC.NewClient()
	if err != nil {
		log.Error("Error connecting to Players RPC", zap.String("config", config.String()), zap.Error(err))
		return nil, errors.ERPCConnection.NewError(err)
	}

	playersClient = client
	return playersClient, nil
}
