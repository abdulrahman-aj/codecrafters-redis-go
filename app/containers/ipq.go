package containers

type LessFunc[V any] func(a, b V) bool

type KeyFunc[V any, K comparable] func(v V) K

// TODO: support shrink-to-fit
type IndexedPriorityQueue[V any, K comparable] struct {
	heap  []V         // heap-ordered values
	pos   map[K]int   // key -> index in heap
	less  LessFunc[V] // heap comparator
	keyFn KeyFunc[V, K]
}

func NewIndexedPriorityQueue[V any, K comparable](keyFn KeyFunc[V, K], less LessFunc[V]) *IndexedPriorityQueue[V, K] {
	return &IndexedPriorityQueue[V, K]{
		pos:   make(map[K]int),
		less:  less,
		keyFn: keyFn,
	}
}

func (pq *IndexedPriorityQueue[V, K]) Len() int {
	return len(pq.heap)
}

func (pq *IndexedPriorityQueue[V, K]) Peek() (V, bool) {
	if len(pq.heap) == 0 {
		var zero V
		return zero, false
	}
	return pq.heap[0], true
}

func (pq *IndexedPriorityQueue[V, K]) Dequeue() (V, bool) {
	v, ok := pq.Peek()
	if !ok {
		return v, false
	}
	pq.Remove(v)
	return v, true
}

func (pq *IndexedPriorityQueue[V, K]) Enqueue(val V) {
	k := pq.keyFn(val)
	if i, ok := pq.pos[k]; ok {
		pq.heap[i] = val
		pq.fix(i)
		return
	}
	pq.pos[k] = len(pq.heap)
	pq.heap = append(pq.heap, val)
	pq.swim(len(pq.heap) - 1)
}

func (pq *IndexedPriorityQueue[V, K]) Remove(val V) {
	k := pq.keyFn(val)
	i, ok := pq.pos[k]
	if !ok {
		return
	}

	n := len(pq.heap) - 1
	if i < n {
		pq.swap(i, n)
		pq.heap = pq.heap[:n]
		pq.fix(i)
	} else {
		pq.heap = pq.heap[:n]
	}

	delete(pq.pos, k)
}

func (pq *IndexedPriorityQueue[V, K]) fix(i int) {
	pq.swim(i)
	pq.sink(i)
}

func (pq *IndexedPriorityQueue[V, K]) swap(i, j int) {
	pq.heap[i], pq.heap[j] = pq.heap[j], pq.heap[i]
	pq.pos[pq.keyFn(pq.heap[i])] = i
	pq.pos[pq.keyFn(pq.heap[j])] = j
}

func (pq *IndexedPriorityQueue[V, K]) swim(i int) {
	for i > 0 {
		p := (i - 1) / 2
		if !pq.less(pq.heap[i], pq.heap[p]) {
			break
		}
		pq.swap(i, p)
		i = p
	}
}

func (pq *IndexedPriorityQueue[V, K]) sink(i int) {
	for {
		l := 2*i + 1
		r := 2*i + 2
		s := i
		if l < len(pq.heap) && pq.less(pq.heap[l], pq.heap[s]) {
			s = l
		}
		if r < len(pq.heap) && pq.less(pq.heap[r], pq.heap[s]) {
			s = r
		}
		if s == i {
			break
		}
		pq.swap(i, s)
		i = s
	}
}
