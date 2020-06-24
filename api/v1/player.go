package v1

import (
	"encoding/json"
	"net/http"

	"github.com/carsonmyers/bublar-assignment/connect"
	"github.com/carsonmyers/bublar-assignment/data"
	"github.com/carsonmyers/bublar-assignment/errors"
	"github.com/carsonmyers/bublar-assignment/proto"
	"github.com/gbrlsnchs/jwt/v2"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func createPlayerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)

	var req data.Player
	if err := DecodeRequest(w, r, &req); err != nil {
		return
	}

	id, ok := vars["id"]
	if !ok || len(id) == 0 {
		id = req.Username
	}

	res := NewResponse()
	if id != req.Username && len(req.Username) > 0 {
		res.AddError(errors.EInvalidRequest.NewError("username does not match route").WithContext("username"))
	}
	if len(req.Username) == 0 {
		res.AddError(errors.EInvalidRequest.NewError("username is required").WithContext("username"))
	}
	if req.Password == nil || len(*req.Password) == 0 {
		res.AddError(errors.EInvalidRequest.NewError("password is required").WithContext("password"))
	}

	if res.Status == StatusError {
		res.Write(w)
		return
	}

	playerSvc, err := connect.Players()
	if err != nil {
		res.AddError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	player, err := playerSvc.Create(&proto.Player{
		Username: req.Username,
		Password: *req.Password,
	})
	if err != nil {
		res.AddError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	res.SetData(player).Write(w)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	var req data.Player
	if err := DecodeRequest(w, r, &req); err != nil {
		return
	}

	res := NewResponse()
	if len(req.Username) == 0 {
		res.AddError(errors.EInvalidRequest.NewError("username is required").WithContext("username"))
	}
	if req.Password == nil || len(*req.Password) == 0 {
		res.AddError(errors.EInvalidRequest.NewError("password is required").WithContext("password"))
	}

	if res.Status == StatusError {
		res.Write(w)
		return
	}

	playerSvc, err := connect.Players()
	if err != nil {
		res.AddError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	tokenResponse, err := playerSvc.Auth(&proto.Player{
		Username: req.Username,
		Password: *req.Password,
	})
	if err != nil {
		res.AddError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	var token *jwt.JWT
	if err := json.Unmarshal([]byte(tokenResponse.Token), &token); err != nil {
		res.AddError(errors.EInternal.NewError(err)).Write(w)
		return
	}

	if tokenResponse.GetUsername() != req.Username {
		res.AddError(errors.EInternal.NewErrorf("token does not match user")).Write(w)
		return
	}

	if err := SetAuth(w, r, token); err != nil {
		res.AddError(err).Write(w)
		return
	}

	res.Write(w)
}

func getPlayerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := GetAuth(r)

	id, ok := vars["id"]
	if !ok || len(id) == 0 {
		if auth != nil {
			id = auth.Audience
		} else {
			FromError(errors.EInvalidRequest.NewError("id or auth token is required")).Write(w)
			return
		}
	}

	if len(id) == 0 {
		FromError(errors.EAuth.NewError("not logged in")).Write(w)
		return
	}

	playerSvc, err := connect.Players()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	player, err := playerSvc.Get(&proto.Player{
		Username: id,
	})
	if err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	result := &data.Player{
		Username: player.GetUsername(),
	}

	if len(player.GetLocation()) > 0 {
		result.Position = &data.Position{
			Location: player.GetLocation(),
			X:        int(player.GetX()),
			Y:        int(player.GetY()),
		}
	}

	FromData(result).Write(w)
}

func listPlayersHandler(w http.ResponseWriter, r *http.Request) {
	playerSvc, err := connect.Players()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	players, err := playerSvc.List()
	if err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	GetLogger(r).Debug("Fetched players from service", zap.Int("players", len(players)))

	results := make([]*data.Player, len(players))
	for i, p := range players {
		results[i] = &data.Player{
			Username: p.GetUsername(),
		}

		if len(p.GetLocation()) > 0 {
			results[i].Position = &data.Position{
				Location: p.GetLocation(),
				X:        int(p.GetX()),
				Y:        int(p.GetY()),
			}
		}
	}

	FromData(results).Write(w)
}

func updatePlayerHandler(w http.ResponseWriter, r *http.Request) {
	FromError(errors.ENotImplemented.NewError("update not implemented")).Write(w)
}

func deletePlayerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := GetAuth(r)

	id, ok := vars["id"]
	if !ok || len(id) == 0 {
		if auth != nil {
			id = auth.Audience
		} else {
			FromError(errors.EInvalidRequest.NewError("id or auth token is required")).Write(w)
			return
		}
	}

	if len(id) == 0 {
		FromError(errors.EAuth.NewError("not logged in")).Write(w)
		return
	}

	playerSvc, err := connect.Players()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	if err := playerSvc.Delete(id); err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	FromData(nil).Write(w)
}

type moveRequest struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func movePlayerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := GetAuth(r)

	var req moveRequest
	if err := DecodeRequest(w, r, &req); err != nil {
		return
	}

	id, ok := vars["id"]
	if !ok || len(id) == 0 {
		if auth != nil {
			id = auth.Audience
		} else {
			FromError(errors.EInvalidRequest.NewError("id or auth token is required")).Write(w)
			return
		}
	}

	if len(id) == 0 {
		FromError(errors.EAuth.NewError("not logged in")).Write(w)
		return
	}

	playerSvc, err := connect.Players()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	position, err := playerSvc.Move(id, int32(req.X), int32(req.Y))
	if err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	FromData(&data.Player{
		Username: id,
		Position: &data.Position{
			Location: position.GetLocation(),
			X:        int(position.GetX()),
			Y:        int(position.GetY()),
		},
	}).Write(w)
}

type travelRequest struct {
	Location string `json:"location"`
}

func travelPlayerHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	auth := GetAuth(r)

	var req travelRequest
	if err := DecodeRequest(w, r, &req); err != nil {
		return
	}

	id, ok := vars["id"]
	if !ok || len(id) == 0 {
		if auth != nil {
			id = auth.Audience
		} else {
			FromError(errors.EInvalidRequest.NewError("id or auth token required")).Write(w)
			return
		}
	}

	if len(id) == 0 {
		FromError(errors.EAuth.NewError("not logged in")).Write(w)
		return
	}

	playerSvc, err := connect.Players()
	if err != nil {
		FromError(errors.ERPCConnection.NewError(err)).Write(w)
		return
	}

	res, err := playerSvc.Travel(id, req.Location)
	if err != nil {
		FromError(errors.ERPC.NewError(err)).Write(w)
		return
	}

	position := res.GetPosition()

	FromData(&data.Player{
		Username: res.GetPlayer().GetUsername(),
		Position: &data.Position{
			Location: position.GetLocation(),
			X:        int(position.GetX()),
			Y:        int(position.GetY()),
		},
	}).Write(w)
}
