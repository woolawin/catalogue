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
	"github.com/woolawin/catalogue/internal/registry"
	"github.com/woolawin/catalogue/internal/update"
)

var ErrInvalidMessage = errors.New("invalid command syntax")
var ErrUnknownAction = errors.New("unknown action")

const path = "/var/run/catalogue.sock"

type Server struct {
	system internal.System
	api    *ext.API
	logger internal.Logger
	log    *internal.Log

	listener net.Listener
}

func NewServer(logger internal.Logger, system internal.System, api *ext.API) *Server {
	log := internal.NewLog(logger)
	log.Stage("server")
	return &Server{log: log, logger: logger, system: system, api: api}
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
	log    *internal.Log
}

func (session *Session) Log(stmt *internal.LogStatement) {
	withoutLogger := internal.NewLogStatement(stmt.Stage, stmt.Level, stmt.Timestamp, stmt.Message, stmt.IsErr)
	err := session.writer.Encode(&Message{Log: &Log{Statement: &withoutLogger}})
	if err != nil {
		slog.Error("failed to write log reply", "error", err)
	}
}

func (session *Session) end(ok bool, value any) {
	err := session.writer.Encode(&Message{End: &End{Ok: ok, Value: value}})
	if err != nil {
		slog.Error("failed to write end reply", "error", err)
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

	log := internal.NewLog(internal.NewMultiLogger(&session, server.logger))
	session.log = log

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
	session.log.Stage("server")
	if err != nil {
		session.log.Err(err, "can not get remote argument")
		session.end(false, nil)
		return
	}
	if !ok {
		session.log.Err(nil, "missing required remote argument")
		session.end(false, nil)
		return
	}

	protocol, ok, raw, err := session.msg.Cmd.IntArg("protocol")
	if err != nil {
		session.log.Err(err, "invalid protocol value '%s'", raw)
		session.end(false, nil)
		return
	}
	if !ok {
		session.log.Err(nil, "missing required protocol argument")
		session.end(false, nil)
		return
	}

	ok = add.Add(config.Protocol(protocol), remote, server.log, server.system, server.api)
	session.end(ok, nil)
}

func (server *Server) list(session *Session) {
	session.log.Stage("server")
	packages, err := registry.ListPackages()
	if err != nil {
		session.log.Err(err, "failed to list packages")
		return
	}
	session.end(true, packages)
}

func (server *Server) update(session *Session) {
	session.log.Stage("server")
	component, found, err := session.msg.Cmd.StringArg("component")
	if err != nil {
		session.log.Err(err, "failed to get component argument from client")
		session.end(false, nil)
		return
	}

	if !found {
		// support updating all later on
		session.end(true, nil)
		return
	}

	record, found, err := registry.GetPackageRecord(component)
	if err != nil {
		session.log.Err(err, "failed to get package record")
		session.end(false, nil)
		return
	}

	if !found {
		session.log.Err(nil, "could not find package '%s'", component)
		session.end(false, nil)
		return
	}

	ok := update.Update(record, session.log, server.system, server.api)
	session.end(ok, nil)
}
