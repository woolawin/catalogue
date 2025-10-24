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
	"github.com/woolawin/catalogue/internal/config"
	"github.com/woolawin/catalogue/internal/ext"
	reg "github.com/woolawin/catalogue/internal/registry"
	"github.com/woolawin/catalogue/internal/update"
)

var ErrInvalidMessage = errors.New("invalid command syntax")
var ErrUnknownAction = errors.New("unknown action")

const path = "/var/run/catalogue.sock"

type Server struct {
	system   internal.System
	api      *ext.API
	registry reg.Registry
	log      *internal.Log

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

type Session struct {
	reader *msgpacklib.Decoder
	writer *msgpacklib.Encoder
	msg    *Message
}

func (session *Session) end(ok bool, value any) {
	err := session.writer.Encode(&Message{End: &End{Ok: ok, Value: value}})
	if err != nil {
		slog.Error("failed to write end reply", "error", err)
	}
}

func (session *Session) log(message string) {
	err := session.writer.Encode(&Message{Log: &Log{Value: message}})
	if err != nil {
		slog.Error("failed to write log reply", "error", err)
	}
}

func (server *Server) handle(conn net.Conn) {
	defer conn.Close()

	reader := msgpacklib.NewDecoder(conn)
	writer := msgpacklib.NewEncoder(conn)

	msg := Message{}
	err := reader.Decode(&msg)
	if err != nil {
		slog.Error("can not read message", "error", err)
		return
	}

	if msg.Cmd == nil {
		slog.Error("received non command message from client")
		return
	}

	session := Session{msg: &msg, reader: reader, writer: writer}

	switch msg.Cmd.Command {
	case Add:
		server.add(&session)
	case Update:
		server.update(&session)
	case ListPackages:
		server.list(&session)
	}

}

func (server *Server) add(session *Session) {
	remote, ok, err := session.msg.Cmd.StringArg("remote")
	if err != nil {
		slog.Warn("can not get remote argument", "error", err)
		return
	}
	if !ok {
		slog.Error("missing required remote argument")
		return
	}

	protocol, ok, raw, err := session.msg.Cmd.IntArg("protocol")
	if err != nil {
		slog.Error("invalid protocol value", "value", raw)
		return
	}
	if !ok {
		slog.Error("missing required protocol argument")
		return
	}

	err = add.Add(config.Protocol(protocol), remote, server.log, server.system, server.api, server.registry)
	if err != nil {
		slog.Error("failed to add package", "remote", remote, "error", err)
		session.log(err.Error())
	}
	session.end(err == nil, nil)
}

func (server *Server) list(session *Session) {
	packages, err := server.registry.ListPackages()
	if err != nil {
		slog.Error("failed to list packages", "error", err)
		session.log(err.Error())
		session.end(false, nil)
		return
	}
	session.end(true, packages)
}

func (server *Server) update(session *Session) {
	component, found, err := session.msg.Cmd.StringArg("component")
	if err != nil {
		slog.Error("failed to get component arg", "error", err)
		session.log("daemon could not get component argument")
		session.end(false, nil)
		return
	}

	if !found {
		// support updating all later on
		session.end(true, nil)
		return
	}

	record, found, err := server.registry.GetPackageRecord(component)
	if err != nil {
		slog.Error("failed to get package record", "error", err)
		session.log(err.Error())
		session.end(false, nil)
		return
	}

	if !found {
		slog.Error("could not find package", "name", component)
		session.log("could not find package")
		session.end(false, nil)
		return
	}

	ok := update.Update(record, server.log, server.system, server.api, server.registry)
	session.end(ok, nil)
}
