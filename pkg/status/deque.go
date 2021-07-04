package status
/*Simple deque written in Go*/
type node struct {
	prev  *node
	next  *node
	value uint64
}

type deque struct {
	head    *node
	tail    *node
	size    int
	maxsize int
	sum     uint64
}

func newDeque(maxsize int) *deque {
	return &deque{
		nil,
		nil,
		0,
		maxsize,
		0,
	}
}



func (q *deque) isEmpty() bool {
	return q.size == 0
}

func (q *deque) isFull() bool {
	return q.size == q.maxsize
}

func (q *deque) push(value uint64) {
	if q.isFull() {
		q.pop()
	}
	if q.isEmpty() {
		item := &node{nil, nil, value}
		q.head = item
		q.tail = item
	} else {
		item := &node{q.tail, nil, value}
		q.tail.next = item
		q.tail = item
	}
	q.size++
	if q.size >= 2 {
		q.sum += q.tail.value - q.tail.prev.value
	}
}

func (q *deque) pop() {
	q.sum -= q.head.next.value - q.head.value
	q.head = q.head.next
	q.size--
}

func (q *deque) popBack() {
	q.sum -= q.tail.value - q.tail.prev.value
	q.tail = q.tail.prev
	q.size--
}

func (q *deque) avg() float64 {
	return float64(q.sum) / float64(q.size)
}
