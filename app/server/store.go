package server

import "time"

type store struct {
	data map[string]entry

	// TODO:
	// - implement background GC in addition to lazy delete, e.g: by sending a "gc" request to the event loop
	// - shrink-to-fit?
}

func newStore() *store {
	return &store{data: map[string]entry{}}
}

func (s *store) get(key string) (entry, bool) {
	v, ok := s.data[key]
	if !ok {
		return entry{}, false
	}

	if v.isExpired() {
		delete(s.data, key)
		return entry{}, false
	}

	return v, true
}

func (s *store) set(key string, e entry) {
	if l, ok := e.value.([]string); ok && len(l) == 0 {
		delete(s.data, key)
	} else {
		s.data[key] = e
	}
}

type entry struct {
	value     any
	expiresAt time.Time
}

func (e *entry) isExpired() bool {
	return !e.expiresAt.IsZero() && time.Now().After(e.expiresAt)
}
