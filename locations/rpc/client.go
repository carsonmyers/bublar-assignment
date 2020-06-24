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

// Client - RPC client for locations service
type Client struct {
	client proto.LocationsClient
}

// NewClient - create a new RPC client
func NewClient() (*Client, error) {
	var opts = []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}

	conf := configure.GetLocations()
	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", conf.Host, conf.Port), opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		client: proto.NewLocationsClient(conn),
	}, nil
}

// Create - send a create location request
func (c *Client) Create(location *proto.Location) (*proto.Location, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	return c.client.Create(ctx, location)
}

// Get - send a get location request
func (c *Client) Get(location *proto.Location) (*proto.Location, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	return c.client.Get(ctx, location)
}

// List - send a list locations request
func (c *Client) List() ([]*proto.Location, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	src, err := c.client.List(ctx, &proto.Empty{})
	if err != nil {
		return nil, err
	}

	res := make([]*proto.Location, 0)
	for {
		var msg proto.Location
		if err := src.RecvMsg(&msg); err != nil {
			if err == io.EOF {
				return res, nil
			}

			log.Error("Error receiving locations", zap.Error(err))
			return nil, err
		}

		log.Debug("Receiving location", zap.String("name", msg.GetName()))
		res = append(res, &msg)
	}
}

// ListPlayers - send a list players request
func (c *Client) ListPlayers(location *proto.Location) ([]*proto.Player, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	src, err := c.client.ListPlayers(ctx, location)
	if err != nil {
		return nil, err
	}

	res := make([]*proto.Player, 0)
	for {
		var msg proto.Player
		if err := src.RecvMsg(&msg); err != nil {
			if err == io.EOF {
				return res, nil
			}

			log.Error("Error receiving players", zap.Error(err))
			return nil, err
		}

		log.Debug("Receiving player from location", zap.String("location", location.GetName()), zap.String("username", msg.GetUsername()))
		res = append(res, &msg)
	}
}

// Update - send an update location request
func (c *Client) Update(location *proto.LocationUpdate) (*proto.Location, error) {
	ctx, cancel := c.ctx()
	defer cancel()
	return c.client.Update(ctx, location)
}

// Delete - send a delete location request
func (c *Client) Delete(name string) error {
	ctx, cancel := c.ctx()
	defer cancel()
	_, err := c.client.Delete(ctx, &proto.Location{
		Name: name,
	})

	return err
}

func (c *Client) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 10*time.Second)
}
