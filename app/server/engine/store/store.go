package store

import (
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/server/engine/context"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/lists"
	"github.com/codecrafters-io/redis-starter-go/app/server/engine/store/streams"
)

type Store struct {
	data map[string]Object
	// TODO:
	// - implement background GC in addition to lazy delete, e.g: by sending a "gc" request to the event loop
	// - shrink-to-fit?
}

func New() *Store {
	return &Store{data: map[string]Object{}}
}

func (s *Store) Get(key string) (Object, bool) {
	o, ok := s.data[key]
	if !ok {
		return Object{}, false
	}

	if o.IsExpired() {
		delete(s.data, key)
		return Object{}, false
	}

	return o, true
}

func (s *Store) Set(ctx *context.Request, key string, o Object) {
	ctx.TouchedKeys[key] = true

	switch v := o.Value.(type) {
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

type Object struct {
	Value     any
	ExpiresAt time.Time
}

func (e *Object) IsExpired() bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}
