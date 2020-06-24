package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/carsonmyers/bublar-assignment/logger"
	"github.com/carsonmyers/bublar-assignment/players"
	"github.com/carsonmyers/bublar-assignment/proto"
	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var (
	server  *grpc.Server
	log     = logger.GetLogger()
	signals = make(chan os.Signal, 1)
)

type config struct {
	Players  *configure.PlayersConfig
	Postgres *configure.PostgresConfig
	Redis    *configure.RedisConfig
}

var defaultConfig = config{
	Players:  &configure.DefaultPlayersConfig,
	Postgres: &configure.DefaultPostgresConfig,
	Redis:    &configure.DefaultRedisConfig,
}

func main() {
	conf := defaultConfig
	envconfig.MustProcess("players", conf.Players)
	envconfig.MustProcess("postgres", conf.Postgres)
	envconfig.MustProcess("redis", conf.Redis)

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
			server.Stop()
			os.Exit(1)
		}

		log.Info("Attempting to shut down gracefully", zap.String("signal", s.String()))
		quitting = true

		shutdownComplete := make(chan bool)

		go func() {
			server.GracefulStop()
			shutdownComplete <- true
		}()

		select {
		case <-shutdownComplete:
			return
		case <-time.After(10 * time.Second):
			log.Fatal("Shutdown timeout")
		}
	}
}

func start(conf config) {
	configure.Players(conf.Players)
	configure.Postgres(conf.Postgres)
	configure.Redis(conf.Redis)

	if err := players.Migrate(); err != nil {
		log.Fatal("Could not migrate data", zap.Error(err))
	}

	listen, err := net.Listen(conf.Players.Protocol, fmt.Sprintf("%s:%d", conf.Players.Host, conf.Players.Port))
	if err != nil {
		log.Fatal("Could not create listener", zap.Error(err))
	}

	server = grpc.NewServer()
	proto.RegisterPlayersServer(server, &Server{})

	go func() {
		if err := server.Serve(listen); err != nil {
			log.Fatal("Fatal server error", zap.Error(err))
		}

		close(signals)
	}()

	log.Info(fmt.Sprintf("Players service is listening on %s", conf.Players.String()))
}
