package server

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
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
		o = object{value: []string{}}
	}

	list, ok := o.value.([]string)
	if !ok {
		return errWrongType
	}

	list = append(list, req.args[1:]...)
	o.value = list

	s.store.set(key, o)
	req.touchedKeys = append(req.touchedKeys, key)
	return resp.Integer(len(list))
}

func (s *Server) handleLpush(req *request) []byte {
	if len(req.args) < 2 {
		return errNumArgs(req.command)
	}

	key := req.args[0]

	o, ok := s.store.get(key)
	if !ok {
		o = object{value: []string{}}
	}

	list, ok := o.value.([]string)
	if !ok {
		return errWrongType
	}

	elems := req.args[1:]
	slices.Reverse(elems)

	list = append(elems, list...) // TODO: consider using a linked-list to optimize this
	o.value = list

	s.store.set(key, o)
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

	o, ok := s.store.get(key)
	if !ok {
		return resp.Array(nil)
	}

	list, ok := o.value.([]string)
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
	o, ok := s.store.get(key)
	if !ok {
		return resp.Integer(0)
	}

	list, ok := o.value.([]string)
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

	o, ok := s.store.get(key)
	if !ok {
		return resp.NullBulkString
	}

	list, ok := o.value.([]string)
	if !ok {
		return errWrongType
	}

	count = min(count, len(list))
	ret := list[:count]
	o.value = list[count:]

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
		return errInvalidTimeout, true
	}

	req.dependency = key

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

	list, ok := o.value.([]string)
	if !ok {
		return nil, false
	}

	o.value = list[1:]
	s.store.set(key, o)

	req.touchedKeys = append(req.touchedKeys, key)
	return resp.Array([]string{key, list[0]}), true
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
	case []string:
		return resp.SimpleString("list")
	case []map[string]string:
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
		o = object{value: []map[string]string{}}
	}

	stream, ok := o.value.([]map[string]string)
	if !ok {
		return errWrongType
	}

	parseEntryID := func(entryID string) (int, int, error) {
		var milliSeconds, sequenceNumber int
		_, err := fmt.Sscanf(entryID, "%d-%d", &milliSeconds, &sequenceNumber)
		return milliSeconds, sequenceNumber, err
	}

	milliSeconds, sequenceNumber, err := parseEntryID(entryID)
	if err != nil {
		return resp.SimpleError("ERR Invalid stream ID specified as stream command argument")
	}

	if milliSeconds <= 0 && sequenceNumber <= 0 {
		return resp.SimpleError("ERR The ID specified in XADD must be greater than 0-0")
	}

	if len(stream) != 0 {
		lastEntry := stream[len(stream)-1]
		lastMilliSeconds, lastSequenceNumber, err := parseEntryID(lastEntry["id"])
		if err != nil {
			panic(err)
		}

		if milliSeconds < lastMilliSeconds ||
			milliSeconds == lastMilliSeconds && sequenceNumber <= lastSequenceNumber {
			return resp.SimpleError("ERR The ID specified in XADD is equal or smaller than the target stream top item")
		}
	}

	streamEntry := map[string]string{"id": entryID}
	for i := 0; i < len(kvs); i += 2 {
		k, v := kvs[i], kvs[i+1]
		streamEntry[k] = v
	}

	stream = append(stream, streamEntry)

	o.value = stream
	s.store.set(key, o)
	req.touchedKeys = append(req.touchedKeys, key)

	return resp.BulkString(entryID)
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
