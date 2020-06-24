package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carsonmyers/bublar-assignment/api"
	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/carsonmyers/bublar-assignment/logger"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
)

var (
	server  *http.Server
	log     = logger.GetLogger()
	signals = make(chan os.Signal, 1)
)

type config struct {
	API       *configure.APIConfig
	Locations *configure.LocationsConfig
	Players   *configure.PlayersConfig
}

var defaultConfig = config{
	API:       &configure.DefaultAPIConfig,
	Locations: &configure.DefaultLocationsConfig,
	Players:   &configure.DefaultPlayersConfig,
}

func main() {
	conf := defaultConfig
	envconfig.MustProcess("api", conf.API)
	envconfig.MustProcess("locations", conf.Locations)
	envconfig.MustProcess("players", conf.Players)

	confJSON, _ := json.MarshalIndent(conf, "", "\t")
	log.Info(fmt.Sprintf("Configuration: %s", confJSON))

	start(conf)

	signal.Notify(signals,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	quitting := false
	for s := range signals {
		if quitting || s == syscall.SIGHUP {
			log.Error("Forcing shutdown", zap.String("signal", s.String()))
			server.Close()
			os.Exit(1)
		}

		log.Info("Attempting to shut down gracefully", zap.String("signal", s.String()))
		quitting = true
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Fatal("Shutdown error", zap.Error(err))
		}
	}
}

func start(conf config) {
	configure.API(conf.API)
	configure.Players(conf.Players)
	configure.Locations(conf.Locations)

	server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", conf.API.Host, conf.API.Port),
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
		IdleTimeout:  60 * time.Second,
		Handler:      api.GetAPI(),
	}

	go func() {
		err := server.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal("Fatal server error", zap.Error(err))
		}

		close(signals)
	}()

	log.Info(fmt.Sprintf("API Server is listening on %s", conf.Locations.String()))
}
