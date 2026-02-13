package containers

type LessFunc[T any] func(a, b T) bool

type PriorityQueue[T any] struct {
	data []T
	less LessFunc[T]
}

func NewPriorityQueue[T any](less LessFunc[T]) *PriorityQueue[T] {
	return &PriorityQueue[T]{less: less}
}

func (pq *PriorityQueue[T]) Len() int {
	return len(pq.data)
}

func (pq *PriorityQueue[T]) Peek() (T, bool) {
	if len(pq.data) == 0 {
		var zero T
		return zero, false
	}
	return pq.data[0], true
}

func (pq *PriorityQueue[T]) Push(x T) {
	pq.data = append(pq.data, x)
	pq.swim(len(pq.data) - 1)
}

func (pq *PriorityQueue[T]) Pop() (T, bool) {
	if len(pq.data) == 0 {
		var zero T
		return zero, false
	}

	top := pq.data[0]
	last := pq.data[len(pq.data)-1]
	pq.data = pq.data[:len(pq.data)-1]

	if len(pq.data) > 0 {
		pq.data[0] = last
		pq.sink(0)
	}

	return top, true
}

func (pq *PriorityQueue[T]) swim(i int) {
	for {
		parent := (i - 1) / 2
		if i == 0 || !pq.less(pq.data[i], pq.data[parent]) {
			break
		}
		pq.data[i], pq.data[parent] = pq.data[parent], pq.data[i]
		i = parent
	}
}

func (pq *PriorityQueue[T]) sink(i int) {
	n := len(pq.data)
	for {
		left := 2*i + 1
		right := 2*i + 2
		smallest := i

		if left < n && pq.less(pq.data[left], pq.data[smallest]) {
			smallest = left
		}
		if right < n && pq.less(pq.data[right], pq.data[smallest]) {
			smallest = right
		}
		if smallest == i {
			break
		}
		pq.data[i], pq.data[smallest] = pq.data[smallest], pq.data[i]
		i = smallest
	}
}
