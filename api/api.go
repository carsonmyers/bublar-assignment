package api

import (
	v1 "github.com/carsonmyers/bublar-assignment/api/v1"
	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/gorilla/mux"
)

var router *mux.Router

// GetAPI - initialize or get a previously initialized API router
func GetAPI() *mux.Router {
	if router == nil {
		conf := configure.GetAPI()

		router = mux.NewRouter()
		v1.Init(router.PathPrefix(conf.BasePath).Subrouter())
	}

	return router
}
