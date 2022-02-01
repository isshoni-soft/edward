package network

import (
	"container/list"
	"fmt"
)

// GO 1.18 PLEASE COME
// This is basically a rip off of the stack code for the isshoni-client but the push method pushes back and not front
type messageQueue struct {
	queue *list.List
}

func newMessageQueue() *messageQueue {
	return &messageQueue{
		queue: list.New(),
	}
}

func (m *messageQueue) Push(str string) {
	m.queue.PushBack(str)
}

func (m *messageQueue) Pop() (string, error) {
	if m.Empty() {
		return "", fmt.Errorf("pop error: empty stack")
	}

	ele := m.queue.Front()

	m.queue.Remove(ele)

	return m.castError(ele.Value)
}

func (m *messageQueue) Peep() (string, error) {
	if m.Empty() {
		return "", fmt.Errorf("peep error: empty stack")
	}

	return m.castError(m.queue.Front().Value)
}

func (m *messageQueue) Size() int {
	return m.queue.Len()
}

func (m *messageQueue) Empty() bool {
	return m.Size() == 0
}

func (m messageQueue) castError(i interface{}) (string, error) {
	if v, ok := i.(string); ok {
		return v, nil
	} else {
		return "", fmt.Errorf("conversion error: datatype in stack doesn't match high level type")
	}
}
