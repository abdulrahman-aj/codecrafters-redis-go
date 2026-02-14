package server

import (
	"fmt"
	"slices"
	"strconv"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
)

func (s *Server) handlePing(req *request) []byte {
	if len(req.args) != 0 {
		return errNumArgs(req.command)
	}
	return resp.SimpleString("PONG")
}

func (s *Server) handleEcho(req *request) []byte {
	if len(req.args) != 1 {
		return errNumArgs(req.command)
	}
	return resp.BulkString(req.args[0])
}

func (s *Server) handleSet(req *request) []byte {
	if len(req.args) != 2 && len(req.args) != 4 {
		return errNumArgs(req.command)
	}

	var expiresAt time.Time

	if len(req.args) == 4 {
		ttl, err := strconv.Atoi(req.args[3])
		if err != nil {
			return errInvalidInteger
		}

		switch req.args[2] {
		case "PX":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Millisecond)
		case "EX":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
		default:
			return errSyntaxError
		}
	}

	key, value := req.args[0], req.args[1]
	s.storage[key] = entry{value: value, expiresAt: expiresAt}

	req.touchedKeys = append(req.touchedKeys, key)
	return resp.SimpleString("OK")
}

func (s *Server) handleGet(req *request) []byte {
	if len(req.args) != 1 {
		return errNumArgs(req.command)
	}
	key := req.args[0]
	e, ok := s.storage[key]
	if !ok {
		return resp.NullBulkString
	}

	if e.isExpired() {
		delete(s.storage, key)
		return resp.NullBulkString
	}

	valueStr, ok := e.value.(string)
	if !ok {
		return errWrongType
	}

	return resp.BulkString(valueStr)
}

func (s *Server) handleRpush(req *request) []byte {
	if len(req.args) < 2 {
		return errNumArgs(req.command)
	}

	key := req.args[0]

	e, ok := s.storage[key]
	if !ok {
		e = entry{value: []string{}}
	}

	list, ok := e.value.([]string)
	if !ok {
		return errWrongType
	}

	list = append(list, req.args[1:]...)
	e.value = list
	s.storage[key] = e

	req.touchedKeys = append(req.touchedKeys, key)
	return resp.Integer(len(list))
}

func (s *Server) handleLpush(req *request) []byte {
	if len(req.args) < 2 {
		return errNumArgs(req.command)
	}

	key := req.args[0]

	e, ok := s.storage[key]
	if !ok {
		e = entry{value: []string{}}
	}

	list, ok := e.value.([]string)
	if !ok {
		return errWrongType
	}

	elems := req.args[1:]
	slices.Reverse(elems)

	list = append(elems, list...) // TODO: consider using a linked-list to optimize this
	e.value = list
	s.storage[key] = e

	req.touchedKeys = append(req.touchedKeys, key)
	return resp.Integer(len(list))
}

func (s *Server) handleLrange(req *request) []byte {
	if len(req.args) != 3 {
		return errNumArgs(req.command)
	}

	key := req.args[0]
	start, err := strconv.Atoi(req.args[1])
	if err != nil {
		return errInvalidInteger
	}
	end, err := strconv.Atoi(req.args[2])
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

func (s *Server) handleLlen(req *request) []byte {
	if len(req.args) != 1 {
		return errNumArgs(req.command)
	}

	key := req.args[0]

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

func (s *Server) handleLpop(req *request) []byte {
	if len(req.args) == 0 || len(req.args) > 2 {
		return errNumArgs(req.command)
	}

	key := req.args[0]
	count := 1
	if len(req.args) == 2 {
		var err error
		count, err = strconv.Atoi(req.args[1])
		if err != nil {
			return errMustBePositive
		}
	}

	if count < 0 {
		return errMustBePositive
	}

	e, ok := s.storage[key]
	if !ok {
		return resp.NullBulkString
	}

	list, ok := e.value.([]string)
	if !ok {
		return errWrongType
	}

	ret := list[:count]
	if len(list[count:]) == 0 {
		delete(s.storage, key)
	} else {
		e.value = list[count:]
		s.storage[key] = e
	}

	req.touchedKeys = append(req.touchedKeys, key)

	if len(req.args) == 1 {
		return resp.BulkString(ret[0])
	}
	return resp.Array(ret)
}

func (s *Server) handleBlpop(req *request) ([]byte, bool) {
	if len(req.args) != 2 {
		return errNumArgs(req.command), true
	}

	key := req.args[0]
	timeout, err := strconv.ParseFloat(req.args[1], 64)
	if err != nil {
		return errInvalidTimeout, true
	}

	if timeout != 0 {
		duration := time.Duration(timeout * float64(time.Second))
		req.deadline = req.requestedAt.Add(duration)
	}

	if !req.deadline.IsZero() && time.Now().After(req.deadline) {
		return resp.NullArray, true
	}

	if list, ok := s.storage[key].value.([]string); !ok || len(list) == 0 {
		req.dependency = key
		return nil, false
	}

	e := s.storage[key]
	list := e.value.([]string)
	e.value = list[1:]
	s.storage[key] = e

	req.touchedKeys = append(req.touchedKeys, key)
	return resp.Array([]string{key, list[0]}), true
}

func (s *Server) handleType(req *request) []byte {
	if len(req.args) != 1 {
		return errNumArgs(req.command)
	}

	key := req.args[0]
	e, ok := s.storage[key]
	if !ok {
		return resp.SimpleString("none")
	}

	if e.isExpired() {
		delete(s.storage, key)
		return resp.NullBulkString
	}

	switch e.value.(type) {
	case string:
		return resp.SimpleString("string")
	case []string:
		return resp.SimpleString("list")
	default:
		panic("unknown type?")
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
	errInvalidTimeout = resp.SimpleError("ERR timeout is not a float or out of range")
	errMustBePositive = resp.SimpleError("ERR value is out of range, must be positive")
	errWrongType      = resp.SimpleError("WRONGTYPE Operation against a key holding the wrong kind of value")
)
