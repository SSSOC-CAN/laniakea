package queue

import (
	"sync"

	"github.com/SSSOC-CAN/laniakea-plugin-sdk/proto"
	bg "github.com/SSSOCPaulCote/blunderguard"
)

const (
	ErrAlreadySubscribed = bg.Error("subscriber with given name already subscribed")
)

type (
	QueueListener struct {
		IsConnected bool
		Signal      chan int
	}
	Queue struct {
		queue     []*proto.Frame
		listeners map[string]*QueueListener
		sync.RWMutex
	}
)

// NewQueue instantiates a new Queue struct
func NewQueue() *Queue {
	return &Queue{
		queue:     []*proto.Frame{},
		listeners: make(map[string]*QueueListener),
	}
}

// Pop returns the first item in the queue and deletes it from the queue
func (q *Queue) Pop() *proto.Frame {
	q.Lock()
	defer q.Unlock()
	var item *proto.Frame
	if len(q.queue) > 0 {
		item = q.queue[0]
		q.queue = q.queue[1:]
	} else {
		q.queue = []*proto.Frame{}
	}
	return item
}

// Push adds a new item to the back of the queue
func (q *Queue) Push(v *proto.Frame) {
	q.Lock()
	defer q.Unlock()
	q.queue = append(q.queue, v)
	newListenerMap := make(map[string]*QueueListener)
	for n, l := range q.listeners {
		if !l.IsConnected {
			close(l.Signal)
			continue
		}
		l.Signal <- len(q.queue) + 1 // + 1 because then the subscriber can know when the channel is closed (if they receive 0)
		newListenerMap[n] = l
	}
	q.listeners = newListenerMap
}

// Subscribe returns a channel which will have signals sent when a new item is pushed as well as an unsub function
func (q *Queue) Subscribe(name string) (chan int, func(), error) {
	q.Lock()
	defer q.Unlock()
	if _, ok := q.listeners[name]; ok {
		return nil, nil, ErrAlreadySubscribed
	}
	q.listeners[name] = &QueueListener{IsConnected: true, Signal: make(chan int, 2)}
	unsub := func() {
		q.Lock()
		defer q.Unlock()
		q.listeners[name].IsConnected = false
	}
	return q.listeners[name].Signal, unsub, nil
}
