package main

import (
	"context"
	"encoding/json"

	"github.com/carsonmyers/bublar-assignment/data"
	"github.com/carsonmyers/bublar-assignment/errors"
	"github.com/carsonmyers/bublar-assignment/players"
	"github.com/carsonmyers/bublar-assignment/proto"
	"go.uber.org/zap"
)

// Server - RPC server for the player service
type Server struct {
	proto.UnimplementedPlayersServer
}

// Create - create a new player
func (s *Server) Create(ctx context.Context, req *proto.Player) (*proto.Player, error) {
	pw := req.GetPassword()
	player := &data.Player{
		Username: req.GetUsername(),
		Password: &pw,
	}

	newPlayer, err := players.CreatePlayer(player)
	if err != nil {
		return nil, err
	}

	return &proto.Player{
		Username: newPlayer.Username,
	}, nil
}

// Get - retrieve a player's details
func (s *Server) Get(ctx context.Context, req *proto.Player) (*proto.Player, error) {
	player, err := players.GetPlayer(req.Username)
	if err != nil {
		return nil, err
	}

	p := &proto.Player{
		Username: player.Username,
	}

	if player.Position != nil {
		p.Location = player.Position.Location
		p.X = int32(player.Position.X)
		p.Y = int32(player.Position.Y)
	}

	return p, nil
}

// Auth - authenticate a user via their password and return an auth token
func (s *Server) Auth(ctx context.Context, req *proto.Player) (*proto.AuthResponse, error) {
	token, err := players.AuthPlayer(req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	tokenStr, err := json.Marshal(token)
	if err != nil {
		return nil, err
	}

	return &proto.AuthResponse{
		Username: req.GetUsername(),
		Token:    string(tokenStr),
	}, nil
}

// List - list all players in the system
func (s *Server) List(req *proto.Empty, srv proto.Players_ListServer) error {
	res, err := players.ListPlayers()
	if err != nil {
		log.Error("Error listing players", zap.Error(err))
		return err
	}

	for _, player := range res {
		log.Debug("Serving player", zap.String("username", player.Username))

		p := &proto.Player{
			Username: player.Username,
		}

		if player.Position != nil {
			p.Location = player.Position.Location
			p.X = int32(player.Position.X)
			p.Y = int32(player.Position.Y)
		}

		if err := srv.Send(p); err != nil {
			log.Error("Error sending player", zap.String("username", player.Username), zap.Error(err))
			return err
		}
	}

	return nil
}

// Update - update a player's information
func (s *Server) Update(ctx context.Context, req *proto.PlayerUpdate) (*proto.Player, error) {
	return nil, errors.ENotImplemented.NewErrorf("players don't contain enough information to justify updates")
}

// Delete - delete a player
func (s *Server) Delete(ctx context.Context, req *proto.Player) (*proto.Player, error) {
	err := players.DeletePlayer(req.GetUsername())
	return req, err
}

// Travel - move a player to a new location
func (s *Server) Travel(ctx context.Context, req *proto.TravelRequest) (*proto.TravelResponse, error) {
	player, err := players.GetPlayer(req.GetUsername())
	if err != nil {
		return nil, err
	}

	position, err := players.Travel(player, req.GetLocation())
	if err != nil {
		return nil, err
	}

	return &proto.TravelResponse{
		Player: &proto.Player{
			Username: player.Username,
			Location: position.Location,
		},
		Position: &proto.Position{
			Location: position.Location,
			X:        int32(position.X),
			Y:        int32(position.Y),
		},
	}, nil
}

// Move - move a player within their current location
func (s *Server) Move(ctx context.Context, req *proto.MoveRequest) (*proto.Position, error) {
	player, err := players.GetPlayer(req.GetUsername())
	if err != nil {
		return nil, err
	}

	if err := players.Move(player, int(req.GetX()), int(req.GetY())); err != nil {
		return nil, err
	}

	return &proto.Position{
		Location: player.Position.Location,
		X:        req.GetX(),
		Y:        req.GetY(),
	}, nil
}
