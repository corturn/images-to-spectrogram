package spectro2

import (
	"sync/atomic"
)

type node struct {
	data *JsonData
	next atomic.Pointer[node]
	prev atomic.Pointer[node]
}

type DeQueue struct {
	head atomic.Pointer[node]
	tail atomic.Pointer[node]
}

// NewQueue creates and initializes a DeQueue
func NewDeQueue() *DeQueue {
	q := &DeQueue{head: atomic.Pointer[node]{}, tail: atomic.Pointer[node]{}}
	return q
}

func (dequeue *DeQueue) PushBottom(data *JsonData) {
	newNode := &node{data: data, next: atomic.Pointer[node]{}, prev: atomic.Pointer[node]{}}
	tail := dequeue.tail.Load()
	if tail == nil {
		// The dequeue is empty: set both head and tail
		dequeue.head.Store(newNode)
		dequeue.tail.Store(newNode)
	} else {
		// Append the new node after the current tail
		newNode.prev.Store(tail)
		tail.next.Store(newNode)
		dequeue.tail.Store(newNode)
	}
}

func (dequeue *DeQueue) PopBottom() *JsonData {

	tail := dequeue.tail.Load()
	if tail == nil {
		// dequeue is empty
		return nil
	}

	head := dequeue.head.Load()
	// If there is only one node, we must compete with thief
	if tail == head {
		// Try to remove the only node via CAS on head
		if dequeue.head.CompareAndSwap(tail, nil) {
			dequeue.tail.Store(nil)
			return tail.data
		}
		// A thief must have stolen the node
		return nil
	}

	// More than one element exists, remove tail node
	newTail := tail.prev.Load()
	if newTail != nil {
		newTail.next.Store(nil) // detach the last node
	}
	dequeue.tail.Store(newTail)
	return tail.data
}

func (dequeue *DeQueue) PopTop() *JsonData {
	oldHead := dequeue.head.Load()

	if oldHead == nil {
		// DeQueue is empty
		return nil
	}

	newHead := oldHead.next.Load()

	// Otherwise try to steal data
	r := oldHead.data

	if dequeue.head.CompareAndSwap(oldHead, newHead) {
		// Return if stealing head is successful
		if newHead != nil {
			newHead.prev.Store(nil)
		} else {
			// The deque has become empty; update tail as well.
			dequeue.tail.Store(nil)
		}
		return r
	}

	// Maybe something happened in the meantime, if so give up and return nil
	return nil
}
