package server

import (
	"maps"
	"slices"
	"time"

	"github.com/codecrafters-io/redis-starter-go/app/containers"
)

type waitQueue struct {
	byTime       *containers.IndexedPriorityQueue[*envelope, int64]
	byDependency map[string]map[int64]*envelope
}

func newWaitQueue() *waitQueue {
	return &waitQueue{
		byTime: containers.NewIndexedPriorityQueue(
			(*envelope).key,
			(*envelope).before,
		),
		byDependency: map[string]map[int64]*envelope{},
	}
}

func (q *waitQueue) enqueue(msg *envelope) {
	if msg.req.hasDeadline() {
		q.byTime.Enqueue(msg)
	}

	for _, dependency := range msg.req.dependencies {
		if _, ok := q.byDependency[dependency]; !ok {
			q.byDependency[dependency] = map[int64]*envelope{}
		}

		q.byDependency[dependency][msg.req.id] = msg
	}
}

func (q *waitQueue) remove(msg *envelope) {
	q.byTime.Remove(msg)

	for _, dependency := range msg.req.dependencies {
		if m, ok := q.byDependency[dependency]; ok {
			delete(m, msg.req.id)
			if len(m) == 0 {
				delete(q.byDependency, dependency)
			}
		}
	}
}

func (q *waitQueue) timeUntilNext() (time.Duration, bool) {
	msg, ok := q.byTime.Peek()
	if !ok {
		return 0, false
	}

	return time.Until(msg.req.deadline), true
}

func (q *waitQueue) dequeueWaiters(key string) []*envelope {
	waiters := slices.Collect(maps.Values(q.byDependency[key]))

	for _, w := range waiters {
		q.remove(w)
	}

	return waiters
}

func (q *waitQueue) dequeueExpired() []*envelope {
	var expired []*envelope

	now := time.Now()
	for q.byTime.Len() > 0 {
		msg, _ := q.byTime.Peek()
		if now.Before(msg.req.deadline) {
			break
		} else {
			expired = append(expired, msg)
			q.remove(msg)
		}
	}

	return expired
}
