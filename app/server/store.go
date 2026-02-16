package server

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/server/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/streams"
)

type store struct {
	data map[string]object

	// TODO:
	// - implement background GC in addition to lazy delete, e.g: by sending a "gc" request to the event loop
	// - shrink-to-fit?
}

func newStore() *store {
	return &store{data: map[string]object{}}
}

func (s *store) get(key string) (object, bool) {
	o, ok := s.data[key]
	if !ok {
		return object{}, false
	}

	if o.isExpired() {
		delete(s.data, key)
		return object{}, false
	}

	return o, true
}

func (s *store) set(key string, o object) {
	switch v := o.value.(type) {
	case lists.List:
		if len(v) == 0 {
			delete(s.data, key)
			return
		}
	case streams.Stream:
		if v.Len() == 0 {
			delete(s.data, key)
			return
		}
	}

	s.data[key] = o
}

type object struct {
	value     any
	expiresAt time.Time
}

func (e *object) isExpired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}
