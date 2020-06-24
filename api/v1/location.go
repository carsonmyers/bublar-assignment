package v1

import (
	"net/http"

	"github.com/carsonmyers/bublar-assignment/connect"
	"github.com/carsonmyers/bublar-assignment/data"
	"github.com/carsonmyers/bublar-assignment/errors"
	"github.com/carsonmyers/bublar-assignment/proto"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func createLocationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var req data.Location
	if err := DecodeRequest(w, r, &req); err != nil {
		return
	}

	id, ok := vars["id"]
	if !ok || len(id) == 0 {
		id = req.Name
	}

	res := NewResponse()
	if id != req.Name && len(req.Name) > 0 {
		res.AddError(errors.EInvalidRequest.NewError("location name does not match route").WithContext("name"))
	}
	if len(req.Name) == 0 {
		res.AddError(errors.EInvalidRequest.NewError("location name is required").WithContext("name"))
	}

	if res.Status == StatusError {
		res.Write(w)
		return
	}

	locationSvc, err := connect.Locations()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	location, err := locationSvc.Create(&proto.Location{
		Name: req.Name,
		X:    int32(req.X),
		Y:    int32(req.Y),
	})
	if err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	FromData(&data.Location{
		Name: location.GetName(),
		X:    int(location.GetX()),
		Y:    int(location.GetY()),
	}).Write(w)
}

func updateLocationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var req data.Location
	if err := DecodeRequest(w, r, &req); err != nil {
		return
	}

	id, ok := vars["id"]
	if !ok || len(id) == 0 {
		id = req.Name
	}

	if len(id) == 0 {
		FromError(errors.EInvalidRequest.NewError("location name is required").WithContext("name")).Write(w)
		return
	}

	locationSvc, err := connect.Locations()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	location, err := locationSvc.Update(&proto.LocationUpdate{
		Id: id,
		Location: &proto.Location{
			Name: req.Name,
			X:    int32(req.X),
			Y:    int32(req.Y),
		},
	})

	if err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	FromData(&data.Location{
		Name: location.GetName(),
		X:    int(location.GetX()),
		Y:    int(location.GetY()),
	}).Write(w)
}

func getLocationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	id, ok := vars["id"]
	if !ok || len(id) == 0 {
		FromError(errors.EInvalidRequest.NewError("location name is required")).Write(w)
		return
	}

	locationSvc, err := connect.Locations()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	location, err := locationSvc.Get(&proto.Location{
		Name: id,
	})

	if err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	FromData(&data.Location{
		Name: location.GetName(),
		X:    int(location.GetX()),
		Y:    int(location.GetY()),
	}).Write(w)
}

func listLocationsHandler(w http.ResponseWriter, r *http.Request) {
	locationSvc, err := connect.Locations()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	locations, err := locationSvc.List()
	if err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	log.Debug("Received locations", zap.Int("locations", len(locations)))

	res := make([]*data.Location, len(locations))
	for i, l := range locations {
		res[i] = &data.Location{
			Name: l.GetName(),
			X:    int(l.GetX()),
			Y:    int(l.GetY()),
		}
	}

	FromData(res).Write(w)
}

func deleteLocationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok || len(id) == 0 {
		FromError(errors.EInvalidRequest.NewError("location name is required")).Write(w)
		return
	}

	locationSvc, err := connect.Locations()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	if err = locationSvc.Delete(id); err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	FromData(nil).Write(w)
}

func getPlayersInLocationHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, ok := vars["id"]
	if !ok || len(id) == 0 {
		FromError(errors.EInvalidRequest.NewError("location name is required")).Write(w)
		return
	}

	locationSvc, err := connect.Locations()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	players, err := locationSvc.ListPlayers(&proto.Location{
		Name: id,
	})

	if err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	res := make([]*data.Player, len(players))
	for i, p := range players {
		res[i] = &data.Player{
			Username: p.GetUsername(),
			Position: &data.Position{
				Location: p.GetLocation(),
				X:        int(p.GetX()),
				Y:        int(p.GetY()),
			},
		}
	}

	FromData(res).Write(w)
}
