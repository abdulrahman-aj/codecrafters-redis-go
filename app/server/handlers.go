package server

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func (s *Server) handlePing(command string, args []string) []byte {
	if len(args) != 0 {
		return errNumArgs(command)
	}
	return resp.SimpleString("PONG")
}

func (s *Server) handleEcho(command string, args []string) []byte {
	if len(args) != 1 {
		return errNumArgs(command)
	}
	return resp.BulkString(args[0])
}

func (s *Server) handleSet(command string, args []string) []byte {
	if len(args) != 2 && len(args) != 4 {
		return errNumArgs(command)
	}

	var expiresAt time.Time

	if len(args) == 4 {
		ttl, err := strconv.Atoi(args[3])
		if err != nil {
			return errInvalidInteger
		}

		switch args[2] {
		case "PX":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Millisecond)
		case "EX":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
		default:
			return errSyntaxError
		}
	}

	s.storage[args[0]] = entry{value: args[1], expiresAt: expiresAt}
	return resp.SimpleString("OK")
}

func (s *Server) handleGet(command string, args []string) []byte {
	if len(args) != 1 {
		return errNumArgs(command)
	}
	entry, ok := s.storage[args[0]]
	if !ok {
		return resp.NullBulkString
	}
	if !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt) {
		delete(s.storage, args[0])
		return resp.NullBulkString
	}
	return resp.BulkString(entry.value)
}

func errNumArgs(command string) []byte {
	msg := fmt.Sprintf("ERR wrong number of arguments for '%s' command", command)
	return resp.SimpleError(msg)
}

func errUnknownCommand(command string) []byte {
	msg := fmt.Sprintf("ERR unknown command '%s'", command)
	return resp.SimpleError(msg)
}

var (
	errSyntaxError    = resp.SimpleError("ERR syntax error")
	errInvalidInteger = resp.SimpleError("ERR value is not an integer or out of range")
)
