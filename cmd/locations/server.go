package main

import (
	"context"

	"github.com/carsonmyers/bublar-assignment/data"
	"github.com/carsonmyers/bublar-assignment/locations"
	"github.com/carsonmyers/bublar-assignment/proto"
	"go.uber.org/zap"
)

// Server - RPC server for the locations service
type Server struct {
	proto.UnimplementedLocationsServer
}

// Create - create a new location
func (s *Server) Create(ctx context.Context, req *proto.Location) (*proto.Location, error) {
	loc := &data.Location{
		Name: req.GetName(),
		X:    int(req.GetX()),
		Y:    int(req.GetY()),
	}

	newLoc, err := locations.CreateLocation(loc)
	if err != nil {
		return nil, err
	}

	return &proto.Location{
		Name: newLoc.Name,
		X:    int32(newLoc.X),
		Y:    int32(newLoc.Y),
	}, nil
}

// Get - get a location by name
func (s *Server) Get(ctx context.Context, req *proto.Location) (*proto.Location, error) {
	loc, err := locations.GetLocation(req.Name)
	if err != nil {
		return nil, err
	}

	return &proto.Location{
		Name: loc.Name,
		X:    int32(loc.X),
		Y:    int32(loc.Y),
	}, nil
}

// List - list all locations
func (s *Server) List(req *proto.Empty, srv proto.Locations_ListServer) error {
	res, err := locations.ListLocations()
	if err != nil {
		return err
	}

	log.Debug("Sending locations", zap.Int("locations", len(res)))
	for _, loc := range res {
		if err := srv.Send(&proto.Location{
			Name: loc.Name,
			X:    int32(loc.X),
			Y:    int32(loc.Y),
		}); err != nil {
			return err
		}
	}

	return nil
}

// ListPlayers - list all players in a location
func (s *Server) ListPlayers(req *proto.Location, srv proto.Locations_ListPlayersServer) error {
	res, err := locations.ListPlayers(req.GetName())
	if err != nil {
		return err
	}

	log.Debug("Sending players from location", zap.String("location", req.GetName()), zap.Int("players", len(res)))
	for _, p := range res {
		player := &proto.Player{
			Username: p.Username,
		}

		if p.Position != nil {
			player.Location = p.Position.Location
			player.X = int32(p.Position.X)
			player.Y = int32(p.Position.Y)
		}

		if err := srv.Send(player); err != nil {
			return err
		}
	}

	return nil
}

// Update - update a location's information
func (s *Server) Update(ctx context.Context, req *proto.LocationUpdate) (*proto.Location, error) {
	newLoc, err := locations.UpdateLocation(req.GetId(), &data.Location{
		Name: req.GetLocation().GetName(),
		X:    int(req.GetLocation().GetX()),
		Y:    int(req.GetLocation().GetY()),
	})
	if err != nil {
		return nil, err
	}

	return &proto.Location{
		Name: newLoc.Name,
		X:    int32(newLoc.X),
		Y:    int32(newLoc.Y),
	}, nil
}

func (s *Server) Delete(ctx context.Context, req *proto.Location) (*proto.Location, error) {
	err := locations.DeleteLocation(req.GetName())
	return req, err
}
