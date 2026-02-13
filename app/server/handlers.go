package server

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func (s *Server) handlePing(command string, args []string) response {
	if len(args) != 0 {
		return response{bytes: errNumArgs(command)}
	}
	return response{bytes: resp.SimpleString("PONG")}
}

func (s *Server) handleEcho(command string, args []string) response {
	if len(args) != 1 {
		return response{bytes: errNumArgs(command)}
	}
	return response{bytes: resp.BulkString(args[0])}
}

func (s *Server) handleSet(command string, args []string) response {
	if len(args) != 2 && len(args) != 4 {
		return response{bytes: errNumArgs(command)}
	}

	var expiresAt time.Time

	if len(args) == 4 {
		ttl, err := strconv.Atoi(args[3])
		if err != nil {
			return response{bytes: errInvalidInteger}
		}

		switch args[2] {
		case "PX":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Millisecond)
		case "EX":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
		default:
			return response{bytes: errSyntaxError}
		}
	}

	key, value := args[0], args[1]
	s.storage[key] = entry{value: value, expiresAt: expiresAt}

	return response{
		bytes:       resp.SimpleString("OK"),
		touchedKeys: []string{key},
	}
}

func (s *Server) handleGet(command string, args []string) response {
	if len(args) != 1 {
		return response{bytes: errNumArgs(command)}
	}
	key := args[0]
	e, ok := s.storage[key]
	if !ok {
		return response{bytes: resp.NullBulkString}
	}
	if !e.expiresAt.IsZero() && time.Now().After(e.expiresAt) {
		delete(s.storage, key)
		return response{bytes: resp.NullBulkString}
	}

	valueStr, ok := e.value.(string)
	if !ok {
		return response{bytes: errWrongType}
	}

	return response{bytes: resp.BulkString(valueStr)}
}

func (s *Server) handleRpush(command string, args []string) response {
	if len(args) < 2 {
		return response{bytes: errNumArgs(command)}
	}

	key := args[0]

	e, ok := s.storage[key]
	if !ok {
		e = entry{value: []string{}}
	}

	list, ok := e.value.([]string)
	if !ok {
		return response{bytes: errWrongType}
	}

	list = append(list, args[1:]...)
	e.value = list
	s.storage[key] = e

	return response{
		bytes:       resp.Integer(len(list)),
		touchedKeys: []string{key},
	}
}

func (s *Server) handleLpush(command string, args []string) response {
	if len(args) < 2 {
		return response{bytes: errNumArgs(command)}
	}

	key := args[0]

	e, ok := s.storage[key]
	if !ok {
		e = entry{value: []string{}}
	}

	list, ok := e.value.([]string)
	if !ok {
		return response{bytes: errWrongType}
	}

	elems := args[1:]
	slices.Reverse(elems)

	list = append(elems, list...) // TODO: consider using a linked-list to optimize this
	e.value = list
	s.storage[key] = e

	return response{
		bytes:       resp.Integer(len(list)),
		touchedKeys: []string{key},
	}
}

func (s *Server) handleLrange(command string, args []string) response {
	if len(args) != 3 {
		return response{bytes: errNumArgs(command)}
	}

	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return response{bytes: errInvalidInteger}
	}
	end, err := strconv.Atoi(args[2])
	if err != nil {
		return response{bytes: errInvalidInteger}
	}

	e, ok := s.storage[key]
	if !ok {
		return response{bytes: resp.Array(nil)}
	}

	list, ok := e.value.([]string)
	if !ok {
		return response{bytes: errWrongType}
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
		return response{bytes: resp.Array(nil)}
	}

	return response{bytes: resp.Array(list[start:min(end+1, n)])}
}

func (s *Server) handleLlen(command string, args []string) response {
	if len(args) != 1 {
		return response{bytes: errNumArgs(command)}
	}

	key := args[0]

	e, ok := s.storage[key]
	if !ok {
		return response{bytes: resp.Integer(0)}
	}

	list, ok := e.value.([]string)
	if !ok {
		return response{bytes: errWrongType}
	}

	return response{bytes: resp.Integer(len(list))}
}

func (s *Server) handleLpop(command string, args []string) response {
	if len(args) == 0 || len(args) > 2 {
		return response{bytes: errNumArgs(command)}
	}

	key := args[0]
	count := 1
	if len(args) == 2 {
		var err error
		count, err = strconv.Atoi(args[1])
		if err != nil {
			return response{bytes: errMustBePositive}
		}
	}

	if count < 0 {
		return response{bytes: errMustBePositive}
	}

	e, ok := s.storage[key]
	if !ok {
		return response{bytes: resp.NullBulkString}
	}

	list, ok := e.value.([]string)
	if !ok {
		return response{bytes: errWrongType}
	}

	ret := list[:count]
	if len(list[count:]) == 0 {
		delete(s.storage, key)
	} else {
		e.value = list[count:]
		s.storage[key] = e
	}

	res := response{touchedKeys: []string{key}}
	if len(args) == 1 {
		res.bytes = resp.BulkString(ret[0])
	} else {
		res.bytes = resp.Array(ret)
	}
	return res
}

func (s *Server) handleBlpop(command string, args []string) response {
	if len(args) != 2 {
		return response{bytes: errNumArgs(command)}
	}

	key := args[0]
	timeout := args[1]

	if timeout != "0" {
		panic("ha?") // TODO: temp for this stage
	}

	e, ok := s.storage[key]
	if !ok {
		return response{waitingOn: key}
	}

	list, ok := e.value.([]string)
	if !ok || len(list) == 0 {
		return response{waitingOn: key}
	}

	ret := list[0]
	e.value = list[1:]
	s.storage[key] = e

	return response{
		bytes:       resp.Array([]string{key, ret}),
		touchedKeys: []string{key},
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
