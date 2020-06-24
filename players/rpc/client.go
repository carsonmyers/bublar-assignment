package rpc

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/carsonmyers/bublar-assignment/configure"
	"github.com/carsonmyers/bublar-assignment/logger"
	"github.com/carsonmyers/bublar-assignment/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

var log = logger.GetLogger()

// Client - RPC client for players service
type Client struct {
	client proto.PlayersClient
}

// NewClient - create a new RPC client
func NewClient() (*Client, error) {
	var opts = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	conf := configure.GetPlayers()
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", conf.Host, conf.Port), opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: proto.NewPlayersClient(conn),
	}, nil
}

// Create - send a player create request
func (c *Client) Create(player *proto.Player) (*proto.Player, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	return c.client.Create(ctx, player)
}

// Get - send a get player request
func (c *Client) Get(player *proto.Player) (*proto.Player, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	return c.client.Get(ctx, player)
}

// Auth - send an auth player request
func (c *Client) Auth(player *proto.Player) (*proto.AuthResponse, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	return c.client.Auth(ctx, player)
}

// List - send a list players request
func (c *Client) List() ([]*proto.Player, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	src, err := c.client.List(ctx, &proto.Empty{})
	if err != nil {
		log.Error("Error listing players", zap.Error(err))
		return nil, err
	}

	res := make([]*proto.Player, 0)
	for {
		var msg proto.Player
		if err := src.RecvMsg(&msg); err != nil {
			if err == io.EOF {
				return res, nil
			}

			log.Error("Error receiving player", zap.Error(err))
			return nil, err
		}

		res = append(res, &msg)
	}
}

// Update - send an update player request
func (c *Client) Update(player *proto.PlayerUpdate) (*proto.Player, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	return c.client.Update(ctx, player)
}

// Travel - send a player travel request
func (c *Client) Travel(username, location string) (*proto.TravelResponse, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	return c.client.Travel(ctx, &proto.TravelRequest{
		Username: username,
		Location: location,
	})
}

// Move - send a move player request
func (c *Client) Move(username string, x, y int32) (*proto.Position, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	return c.client.Move(ctx, &proto.MoveRequest{
		Username: username,
		X:        x,
		Y:        y,
	})
}

// Delete - send a delete player request
func (c *Client) Delete(username string) error {
	ctx, cancel := c.ctx()
	defer cancel()
	_, err := c.client.Delete(ctx, &proto.Player{
		Username: username,
	})

	return err
}

func (c *Client) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
