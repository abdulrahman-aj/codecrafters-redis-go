package server

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/resp"
	"github.com/codecrafters-io/redis-starter-go/app/server/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/streams"
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

		switch strings.ToLower(req.args[2]) {
		case "px":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Millisecond)
		case "ex":
			expiresAt = time.Now().Add(time.Duration(ttl) * time.Second)
		default:
			return errSyntaxError
		}
	}

	key, value := req.args[0], req.args[1]
	s.store.set(key, object{value: value, expiresAt: expiresAt})
	req.touchedKeys = append(req.touchedKeys, key)

	return resp.SimpleString("OK")
}

func (s *Server) handleGet(req *request) []byte {
	if len(req.args) != 1 {
		return errNumArgs(req.command)
	}

	key := req.args[0]
	o, ok := s.store.get(key)
	if !ok {
		return resp.NullBulkString
	}

	valueStr, ok := o.value.(string)
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

	o, ok := s.store.get(key)
	if !ok {
		o = object{value: lists.List{}}
	}

	l, ok := o.value.(lists.List)
	if !ok {
		return errWrongType
	}

	l = append(l, req.args[1:]...)
	o.value = l

	s.store.set(key, o)
	req.touchedKeys = append(req.touchedKeys, key)
	return resp.Integer(len(l))
}

func (s *Server) handleLpush(req *request) []byte {
	if len(req.args) < 2 {
		return errNumArgs(req.command)
	}

	key := req.args[0]

	o, ok := s.store.get(key)
	if !ok {
		o = object{value: lists.List{}}
	}

	l, ok := o.value.(lists.List)
	if !ok {
		return errWrongType
	}

	elems := req.args[1:]
	slices.Reverse(elems)

	l = append(elems, l...)
	o.value = l

	s.store.set(key, o)
	req.touchedKeys = append(req.touchedKeys, key)
	return resp.Integer(len(l))
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

	o, ok := s.store.get(key)
	if !ok {
		return resp.Array(nil)
	}

	l, ok := o.value.(lists.List)
	if !ok {
		return errWrongType
	}

	n := len(l)
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

	return resp.Array(l[start:min(end+1, n)])
}

func (s *Server) handleLlen(req *request) []byte {
	if len(req.args) != 1 {
		return errNumArgs(req.command)
	}

	key := req.args[0]
	o, ok := s.store.get(key)
	if !ok {
		return resp.Integer(0)
	}

	l, ok := o.value.(lists.List)
	if !ok {
		return errWrongType
	}

	return resp.Integer(len(l))
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

	o, ok := s.store.get(key)
	if !ok {
		return resp.NullBulkString
	}

	l, ok := o.value.(lists.List)
	if !ok {
		return errWrongType
	}

	count = min(count, len(l))
	ret := l[:count]
	o.value = l[count:]

	s.store.set(key, o)
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
		return errTimeoutNotFloat, true
	}

	req.dependencies = []string{key}

	if timeout != 0 {
		duration := time.Duration(timeout * float64(time.Second))
		req.deadline = req.requestedAt.Add(duration)
	}

	if req.isExpired() {
		return resp.NullArray, true
	}

	o, ok := s.store.get(key)
	if !ok {
		return nil, false
	}

	l, ok := o.value.(lists.List)
	if !ok {
		return nil, false
	}

	o.value = l[1:]
	s.store.set(key, o)
	req.touchedKeys = append(req.touchedKeys, key)

	return resp.Array([]string{key, l[0]}), true
}

func (s *Server) handleType(req *request) []byte {
	if len(req.args) != 1 {
		return errNumArgs(req.command)
	}

	key := req.args[0]
	o, ok := s.store.get(key)
	if !ok {
		return resp.SimpleString("none")
	}

	switch o.value.(type) {
	case string:
		return resp.SimpleString("string")
	case lists.List:
		return resp.SimpleString("list")
	case streams.Stream:
		return resp.SimpleString("stream")
	default:
		panic("unknown type?")
	}
}

func (s *Server) handleXadd(req *request) []byte {
	if len(req.args) < 4 {
		return errNumArgs(req.command)
	}

	var (
		key     = req.args[0]
		entryID = req.args[1]
		kvs     = req.args[2:]
	)

	if len(kvs)%2 != 0 {
		return errNumArgs(req.command)
	}

	o, ok := s.store.get(key)
	if !ok {
		o = object{value: streams.Stream{}}
	}

	stream, ok := o.value.(streams.Stream)
	if !ok {
		return errWrongType
	}

	var fields []streams.Field
	for i := 0; i < len(kvs); i += 2 {
		fields = append(fields, streams.Field{
			Key:   kvs[i],
			Value: kvs[i+1],
		})
	}

	id, err := stream.Append(entryID, fields, req.requestedAt)
	if err != nil {
		return err
	}

	o.value = stream
	s.store.set(key, o)
	req.touchedKeys = append(req.touchedKeys, key)

	return resp.BulkString(id)
}

func (s *Server) handleXrange(req *request) []byte {
	if len(req.args) != 3 {
		return errNumArgs(req.command)
	}

	var (
		key   = req.args[0]
		start = req.args[1]
		end   = req.args[2]
	)

	o, ok := s.store.get(key)
	if !ok {
		o = object{value: streams.Stream{}}
	}

	stream, ok := o.value.(streams.Stream)
	if !ok {
		return errWrongType
	}

	entries, err := stream.Between(start, end)
	if err != nil {
		return err
	}

	ret := []any{}
	for _, e := range entries {
		ret = append(ret, e.Format())
	}

	return resp.Array(ret)
}

func (s *Server) handleXread(req *request) ([]byte, bool) {
	if len(req.args) < 3 {
		return errNumArgs(req.command), true
	}

	// TODO: parsing the same blocking requests multiple times is a bit wasteful
	var (
		isBlocking = strings.EqualFold(req.args[0], "block")
		keysAndIDs []string
	)

	if isBlocking {
		if len(req.args) < 5 { // block ms streams k1 v1
			return errNumArgs(req.command), true
		}

		if !strings.EqualFold(req.args[2], "streams") {
			return errSyntaxError, true
		}

		timeoutMs, err := strconv.Atoi(req.args[1])
		if err != nil {
			return errTimeoutNotInt, true
		}

		if timeoutMs != 0 {
			req.deadline = req.requestedAt.Add(time.Duration(timeoutMs) * time.Millisecond)
		}

		keysAndIDs = req.args[3:]
	} else {
		if !strings.EqualFold(req.args[0], "streams") {
			return errSyntaxError, true
		}

		keysAndIDs = req.args[1:]
	}

	if req.isExpired() {
		return resp.NullArray, true
	}

	if len(keysAndIDs)%2 != 0 {
		return errUnbalancedXread, true
	}

	var (
		numKeys = len(keysAndIDs) / 2
		keys    = keysAndIDs[:numKeys]
		ids     = keysAndIDs[numKeys:]
		ret     []any
	)

	req.dependencies = keys

	for i := range numKeys {
		o, ok := s.store.get(keys[i])
		if !ok {
			continue
		}

		stream, ok := o.value.(streams.Stream)
		if !ok {
			return errWrongType, true
		}

		var (
			entries []streams.Entry
			err     []byte
		)

		if ids[i] == "$" {
			entries = stream.AfterTime(req.requestedAt)
		} else {
			entries, err = stream.After(ids[i])
		}

		if err != nil {
			return err, true
		}

		if len(entries) != 0 {
			var formatted []any
			for _, e := range entries {
				formatted = append(formatted, e.Format())
			}
			ret = append(ret, []any{keys[i], formatted})
		}
	}

	if isBlocking && len(ret) == 0 {
		return nil, false
	}

	return resp.Array(ret), true
}

func errNumArgs(command string) []byte {
	msg := fmt.Sprintf("ERR wrong number of arguments for '%s' command", command)
	return resp.SimpleError(msg)
}

func errUnknownCommand(command string) []byte {
	msg := fmt.Sprintf("ERR unknown command '%s'", command)
	return resp.SimpleError(msg)
}

// TODO: cleanup/centralize error handling
var (
	errSyntaxError     = resp.SimpleError("ERR syntax error")
	errInvalidInteger  = resp.SimpleError("ERR value is not an integer or out of range")
	errTimeoutNotFloat = resp.SimpleError("ERR timeout is not a float or out of range")
	errTimeoutNotInt   = resp.SimpleError("ERR timeout is not an integer or out of range")
	errMustBePositive  = resp.SimpleError("ERR value is out of range, must be positive")
	errWrongType       = resp.SimpleError("WRONGTYPE Operation against a key holding the wrong kind of value")
	errUnbalancedXread = resp.SimpleError("ERR Unbalanced 'xread' list of streams: for each stream key an ID, '+', or '$' must be specified.")
)
