package v1

import (
	"net/http"

	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/carsonmyers/bublar-assignment/errors"
	"github.com/carsonmyers/bublar-assignment/logger"
	"github.com/gorilla/mux"
)

var log = logger.GetLogger()

// Init - initialize a router with the V1 API
func Init(r *mux.Router) {
	r.Use(RIDMiddleware())
	r.Use(LoggingMiddleware)
	r.Use(authMiddleware)

	r.HandleFunc("/ping", pingHandler).Methods("POST")

	conf := configure.GetAPI()

	initAdminRoutes(conf, r)
	initClientRoutes(r)

	r.PathPrefix("/").HandlerFunc(notFoundHandler)

	log.Info("Initialized v1 router")
}

func initAdminRoutes(conf *configure.APIConfig, base *mux.Router) {
	if !conf.EnableAdmin {
		return
	}

	r := base.PathPrefix("/admin").Subrouter()

	r.HandleFunc("/players/{id}", createPlayerHandler).Methods("PUT")
	r.HandleFunc("/players/{id}", updatePlayerHandler).Methods("PATCH")
	r.HandleFunc("/players/{id}", deletePlayerHandler).Methods("DELETE")
	r.HandleFunc("/players/{id}/move", movePlayerHandler).Methods("POST")
	r.HandleFunc("/players/{id}/travel", travelPlayerHandler).Methods("POST")

	r.HandleFunc("/locations", createLocationHandler).Methods("POST")
	r.HandleFunc("/locations/{id}", createLocationHandler).Methods("PUT")
	r.HandleFunc("/locations/{id}", updateLocationHandler).Methods("PATCH")
	r.HandleFunc("/locations/{id}", deleteLocationHandler).Methods("DELETE")
}

func initClientRoutes(base *mux.Router) {
	r := base.PathPrefix("/client").Subrouter()

	r.HandleFunc("/login", loginHandler).Methods("POST")
	r.HandleFunc("/players", createPlayerHandler).Methods("POST")
	r.HandleFunc("/players", listPlayersHandler).Methods("GET")
	r.HandleFunc("/players/{id}", getPlayerHandler).Methods("GET")

	r.HandleFunc("/locations", listLocationsHandler).Methods("GET")
	r.HandleFunc("/locations/{id}", getLocationHandler).Methods("GET")
	r.HandleFunc("/locations/{id}/players", getPlayersInLocationHandler).Methods("GET")

	r.HandleFunc("/player", getPlayerHandler).Methods("GET")
	r.HandleFunc("/player", updatePlayerHandler).Methods("PATCH")
	r.HandleFunc("/player", deletePlayerHandler).Methods("DELETE")
	r.HandleFunc("/player/move", movePlayerHandler).Methods("POST")
	r.HandleFunc("/player/travel", travelPlayerHandler).Methods("POST")
}

func notFoundHandler(w http.ResponseWriter, r *http.Request) {
	FromError(errors.ENotFound.NewError("endpoint does not exist")).Write(w)
}
