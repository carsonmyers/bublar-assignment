package connect

import (
	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/carsonmyers/bublar-assignment/errors"
	locationsRPC "github.com/carsonmyers/bublar-assignment/locations/rpc"
	"go.uber.org/zap"
)

var locationsClient *locationsRPC.Client

// Locations - connect to the locations rpc service
func Locations() (*locationsRPC.Client, error) {
	if locationsClient != nil {
		return locationsClient, nil
	}

	config := configure.GetLocations()
	log.Info("Connecting to Locations RPC", zap.String("config", config.String()))

	client, err := locationsRPC.NewClient()
	if err != nil {
		log.Error("Error connecting to Locations RPC", zap.String("config", config.String()), zap.Error(err))
		return nil, errors.ERPCConnection.NewError(err)
	}

	locationsClient = client
	return locationsClient, nil
}
