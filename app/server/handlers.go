package server

import (
	"fmt"
	"slices"
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

	key, value := args[0], args[1]
	s.storage[key] = entry{value: value, expiresAt: expiresAt}
	return resp.SimpleString("OK")
}

func (s *Server) handleGet(command string, args []string) []byte {
	if len(args) != 1 {
		return errNumArgs(command)
	}
	key := args[0]
	e, ok := s.storage[key]
	if !ok {
		return resp.NullBulkString
	}
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		delete(s.storage, key)
		return resp.NullBulkString
	}

	valueStr, ok := e.value.(string)
	if !ok {
		return errWrongType
	}

	return resp.BulkString(valueStr)
}

func (s *Server) handleRpush(command string, args []string) []byte {
	if len(args) < 2 {
		return errNumArgs(command)
	}

	key := args[0]

	e, ok := s.storage[key]
	if !ok {
		e = entry{value: []string{}}
	}

	list, ok := e.value.([]string)
	if !ok {
		return errWrongType
	}

	list = append(list, args[1:]...)
	e.value = list
	s.storage[key] = e

	return resp.Integer(len(list))
}

func (s *Server) handleLpush(command string, args []string) []byte {
	if len(args) < 2 {
		return errNumArgs(command)
	}

	key := args[0]

	e, ok := s.storage[key]
	if !ok {
		e = entry{value: []string{}}
	}

	list, ok := e.value.([]string)
	if !ok {
		return errWrongType
	}

	elems := args[1:]
	slices.Reverse(elems)

	list = append(elems, list...) // TODO: consider using a linked-list to optimize this
	e.value = list
	s.storage[key] = e

	return resp.Integer(len(list))
}

func (s *Server) handleLrange(command string, args []string) []byte {
	if len(args) != 3 {
		return errNumArgs(command)
	}

	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return errInvalidInteger
	}
	end, err := strconv.Atoi(args[2])
	if err != nil {
		return errInvalidInteger
	}

	e, ok := s.storage[key]
	if !ok {
		return resp.Array(nil)
	}

	list, ok := e.value.([]string)
	if !ok {
		return errWrongType
	}

	n := len(list)
	normalizeIndex := func(x int) int {
		if x < 0 {
			x += n
		}

		return max(0, x)
	}

	start = normalizeIndex(start)
	end = normalizeIndex(end)

	if start > end || start >= n || n == 0 {
		return resp.Array(nil)
	}

	return resp.Array(list[start:min(end+1, n)])
}

func (s *Server) handleLlen(command string, args []string) []byte {
	if len(args) != 1 {
		return errNumArgs(command)
	}

	key := args[0]

	e, ok := s.storage[key]
	if !ok {
		return resp.Integer(0)
	}

	list, ok := e.value.([]string)
	if !ok {
		return errWrongType
	}

	return resp.Integer(len(list))
}

func (s *Server) handleLpop(command string, args []string) []byte {
	if len(args) == 0 || len(args) > 2 {
		return errNumArgs(command)
	}

	key := args[0]
	count := 1
	if len(args) == 2 {
		var err error
		count, err = strconv.Atoi(args[1])
		if err != nil {
			return errMustBePositive
		}
	}

	if count < 0 {
		return errMustBePositive
	}

	e, ok := s.storage[key]
	if !ok {
		return resp.Integer(0)
	}

	list, ok := e.value.([]string)
	if !ok {
		return errWrongType
	}

	if len(list) == 0 {
		return resp.NullBulkString
	}

	ret := list[:count]
	e.value = list[count:]
	s.storage[key] = e

	if len(args) == 1 {
		return resp.BulkString(ret[0])
	} else {
		return resp.Array(ret)
	}
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
	errMustBePositive = resp.SimpleError("ERR value is out of range, must be positive")
	errWrongType      = resp.SimpleError("WRONGTYPE Operation against a key holding the wrong kind of value")
)
