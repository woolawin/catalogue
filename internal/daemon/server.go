package daemon

import (
	"errors"
	"log/slog"
	"net"
	"os"
	"os/user"
	"strconv"

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
	log      *internal.Log
	system   internal.System
	api      *ext.API
	registry reg.Registry

	listener net.Listener
}

func NewServer(log *internal.Log, system internal.System, api *ext.API, registry reg.Registry) *Server {
	log.Stage("server")
	return &Server{log: log, system: system, api: api, registry: registry}
}

func (server *Server) Start() error {
	err := os.RemoveAll(path)
	if err != nil {
		return internal.ErrOf(err, "can not clean up old socket '%s'", path)
	}

	listener, err := net.Listen("unix", path)
	if err != nil {
		return internal.ErrOf(err, "can not listen to socket '%s'", path)
	}
	server.listener = listener

	group, err := user.LookupGroup("catalogue")
	if err != nil {
		return internal.ErrOf(err, "can not look up ground 'catalogue'")
	}

	currentUser, err := user.Current()
	if err != nil {
		return internal.ErrOf(err, "can not get current user")
	}

	groupID, err := strconv.Atoi(group.Gid)
	if err != nil {
		return internal.ErrOf(err, "invalid group id '%s'", group.Gid)
	}

	currentUserID, err := strconv.Atoi(currentUser.Uid)
	if err != nil {
		return internal.ErrOf(err, "invalid user id '%s'", group.Gid)
	}

	err = os.Chown(path, currentUserID, groupID)
	if err != nil {
		return internal.ErrOf(err, "can not change socket ownership to '%d'/'%d'", currentUserID, groupID)
	}

	err = os.Chmod(path, 0770)
	if err != nil {
		return internal.ErrOf(err, "can not set socket permissions")
	}
	go func() {
		for {
			conn, err := server.listener.Accept()
			if err != nil {
				slog.Error("failed to accept connection", "error", err)
				continue
			}
			go server.handle(conn)
		}
	}()

	return nil
}

func (server *Server) Shutdown() {
	if server.listener == nil {
		return
	}
	server.listener.Close()
}

func (server *Server) handle(conn net.Conn) {
	defer conn.Close()

	reader := msgpacklib.NewDecoder(conn)
	writer := msgpacklib.NewEncoder(conn)

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
			server.add(msg, writer)
		case ListPackages:
			server.list(writer)
		}
	}

}

func (server *Server) add(msg Message, writer *msgpacklib.Encoder) {
	remote, ok, err := msg.Cmd.StringArg("remote")
	if err != nil {
		slog.Warn("can not get remote argument", "error", err)
		return
	}
	if !ok {
		slog.Error("missing required remote argument")
		return
	}

	protocol, ok, raw, err := msg.Cmd.IntArg("protocol")
	if err != nil {
		server.log.Msg(9, "invalid protocol argument").
			With("value", raw).
			Error()
		return
	}
	if !ok {
		slog.Error("missing required protocol argument")
		return
	}

	err = add.Add(clone.Protocol(protocol), remote, server.system, server.api, server.registry)
	ok = err == nil
	if err != nil {
		slog.Error("failed to add package", "remote", remote, "error", err)
		err = writer.Encode(&Message{Log: &Log{Value: err.Error()}})
		if err != nil {
			slog.Error("failed to write log reply", "error", err)
		}
	}

	err = writer.Encode(&Message{End: &End{Ok: ok}})
	if err != nil {
		slog.Error("failed to write end reply", "error", err)
	}
}

func (server *Server) list(writer *msgpacklib.Encoder) {
	packages, err := server.registry.ListPackages()
	if err != nil {
		slog.Error("failed to list packages", "error", err)
		msg := Message{Log: &Log{Value: err.Error()}}
		err = writer.Encode(&msg)
		if err != nil {
			slog.Error("failed to write log reply", "error", err)
		}
		return
	}
	reply := Message{End: &End{Ok: packages != nil, Value: packages}}
	err = writer.Encode(&reply)
	if err != nil {
		slog.Error("failed to write log reply", "error", err)
	}
}
