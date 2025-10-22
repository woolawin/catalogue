package daemon

import (
	"errors"
	"log/slog"
	"net"
	"os"

	msgpacklib "github.com/vmihailenco/msgpack/v5"
	"github.com/woolawin/catalogue/internal"
	"github.com/woolawin/catalogue/internal/add"
	"github.com/woolawin/catalogue/internal/clone"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
)

var ErrInvalidMessage = errors.New("invalid command syntax")
var ErrUnknownAction = errors.New("unknown action")

const path = "/var/run/catalogue.sock"

type Server struct {
	system   internal.System
	api      *ext.API
	registry *reg.DiskRegistry
}

func NewServer(system internal.System, api *ext.API, registry *reg.DiskRegistry) *Server {
	return &Server{system: system, api: api, registry: registry}
}

func (server *Server) start() error {
	err := os.RemoveAll(path)
	if err != nil {
		return internal.ErrOf(err, "can not clean up old socket '%s'", path)
	}

	listener, err := net.Listen("unix", path)
	if err != nil {
		return internal.ErrOf(err, "can not listen to socket '%s'", path)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("failed to accept connection", "error", err)
			continue
		}
		go server.handle(conn)
	}
}

func (server *Server) handle(conn net.Conn) {
	defer conn.Close()

	reader := msgpacklib.NewDecoder(conn)

	for {
		msg := Message{}
		err := reader.Decode(&msg)
		if err != nil {
			slog.Error("can not read message", "error", err)
			return
		}

		if msg.Cmd == nil {
			slog.Error("received non command message from client")
			continue
		}

		switch msg.Cmd.Command {
		case Add:
			server.add(msg)
		}
	}

}

func (server *Server) add(msg Message) {
	remote, ok, err := msg.Cmd.StringArg("remote")
	if err != nil {
		slog.Warn("can not get remote argument", "error", err)
		return
	}
	if !ok {
		slog.Error("missing required remote argument")
		return
	}

	protocol, ok, err := msg.Cmd.IntArg("protocol")
	if err != nil {
		slog.Warn("can not get protocol argument", "error", err)
		return
	}
	if !ok {
		slog.Error("missing required protocol argument")
		return
	}

	add.Add(clone.Protocol(protocol), remote, server.system, server.api, server.registry)
	slog.Info("added")
}
