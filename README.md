# Bublar Technical Assignment
Carson Myers

## Overview

This project models a simple game in which players can join into rooms ("locations") which are located geometrically relative to each other, as well as moving around inside the room.

The player and game data is stored using Postgres and Redis, and is managed by a pair of microservices:

* `locations`, which manages the game world, and
* `players`, which manages the player data.

These services overlap somewhat, in that the players service is responsible for executing player movements, while the locations service needs to list the current locations of players in the world.

These services have no access controls, and are together managed by a pair of identical API services (one exposed, and an admin copy which is accessible only locally) which manage authentication and feature access.

Finally, there is a simple client program which is a scaffold for an actual game client.

### Usage

All the services can be started through `docker-compose`:

   > docker-compose up -d

It's useful to monitor the services through logging:

   > docker-compose logs -f api api_admin locations players

There is no pre-initialized data, so the client program can be used to create some locations and players:

   > docker-compose run client locations create -n level1
   > docker-compose run client locations create -n coolzone -x 1
   > docker-compose run client players create -u test1
   > docker-compose run client players create -u test2

The client program is made up of a command-tree accompanied by help messages for each command and their arguments. The help menu appears if you add `-h` after a command, `help` before or after a command, or by omitting a following command (e.g., `client locations help list`, `client locations list help`, and `client locations list -h` will print the `list` command's help. `client locations`, `client help locations`, etc. will print the `locations` command's help).

For both `create` and `update` commands will use default values for any omitted arguments.

The `players create` command will prompt for a password if one isn't provided on the command-line.

The client program setup in docker-compose is communicating with the `api_admin` service. To use the client in a non-administrative role, just build and use it outside of the docker networks:

   > go build -o ./client ./cmd/client

Alternatively, the `client` service can be copied or modified in `docker-compose.yml` such that it uses the `public` network and the `API_HOST=api` environment variable, and restarted.

Many functions don't require authorization:

   > ./client players list
   > ./client locations list

However, to interact with the game, login is required:

   > ./client login -u test1
   > ./client logout

The user session will be persisted in ~/.client-session. Logging out will delete it.

A player can travel to a location, and then move within that location

   > ./client players travel -l level1
   > ./client players move -x 1 -y 2

The admin can move players around as well:

   > docker-compose run client players travel -u test2 -l coolzone
   > docker-compose run client players move -u test2 -x 11 -y 22

Now that the players are in rooms, that will become visible to everyone:

   > ./client players list
   > ./client locations list players -n coolzone

## Communication

The client program is designed to communicate over an HTTP API, although with the shared configuration and connection packages, as well as a common env configuration scheme, it can easily communicate with HTTPS as well.

There are two API services, one of which has the `API_ENABLEADMIN=true` environment variable set, and is only accessible from within a docker network. This could be realized as a private API running inside a VPN, or that only allows connections from localhost - this setup is in lieu of admin accounts existing side-by-side with player accounts on a single API service. The other service is exposed to the host but does not permit access to any administrative endpoints.

The public (and private, since they are identical apart from an environment variable) API does utilize a user auth system; player accounts are created with a hashed password, and users are issued a JSON web token on login (in the form of a base64-encoded cookie). This token is available in the router context to all routes on the API.

The API service communicates with the locations and players services over grpc with the protocol and messages compiled from a `.proto` file. The RPC interface is relatively simplistic, and the `players` and `locations` packages (or parts of their functionality) could be packaged directly into the API, bypassing the RPC layer altogether or in part with very little effort, since their functionality is separate from both the API and the grpc server binaries.

The binaries can be deployed to a wide variety of environments due to the shared configuration and communication packages - each uses the same code to communicate, and the same sets of environment variables, and are otherwise decoupled. It's simple to run the API with the `./run-api.sh` script (which is little more than some environment variables and a go command) alongside the other services running in docker-compose.

## Missing parts and next steps

* [ ] Makefile: make targets for things like generating the protobuf code, building dev and prod containers, and running the services on the host
* [ ] Security improvements:
   * [ ] TLS support: The APIs should have the ability to accept a key-file and operate over secure connection
   * [ ] Token signing: The authentication tokens are not cryptographically signed and so could be modified by the user to take over another account or extend the token's validity
   * [ ] Token invalidation: The auth token cannot be revoked by the API, and its expiration time is not observed
* [ ] Player updates: the player update endpoint is not implemented, in part because changing the username would cause the auth token to stop working (and possibly _start_ working on a new account created in the former name).
* [ ] Realtime updates: A websocket or other streaming protocol could be setup between the client program and the API (or another service) to make the communication more realtime and game-like
* [ ] Game interface: A simple visual display of the rooms that the player can move around in, and see other players in.