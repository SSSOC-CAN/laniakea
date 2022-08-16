package queue

import (
	"sync"

	bg "github.com/SSSOCPaulCote/blunderguard"
	"github.com/rs/zerolog"
)

const (
	ErrSourceAlreadyRegistered   = bg.Error("source already registered")
	ErrListenerAlreadyRegistered = bg.Error("listener already registered")
	queueManagerName             = "queue-manager"
	defaultSrcQueueSize          = 10
	defaultLisQueueSize          = 20
)

type QueueManager struct {
	incomingQueues map[string]*Queue
	outgoingQueues map[string]map[string]*Queue
	quitChans      map[string]chan struct{}
	Logger         *zerolog.Logger
	wgs            map[string]*sync.WaitGroup
	sync.RWMutex
}

// NewQueueManager initializes a new QueueManager instance
func NewQueueManager(logger *zerolog.Logger) *QueueManager {
	return &QueueManager{
		incomingQueues: make(map[string]*Queue),
		outgoingQueues: make(map[string]map[string]*Queue),
		quitChans:      make(map[string]chan struct{}),
		Logger:         logger,
		wgs:            make(map[string]*sync.WaitGroup),
	}
}

// RegisterSource adds a new entry in the incomingQueues map and returns a newly created queue as well as deregistering function
func (m *QueueManager) RegisterSource(name string) (*Queue, func(), error) {
	if _, ok := m.incomingQueues[name]; ok {
		return nil, nil, ErrSourceAlreadyRegistered
	}
	newQ := NewQueue(defaultSrcQueueSize)
	quitChan := make(chan struct{})
	var wg sync.WaitGroup
	m.Lock()
	m.incomingQueues[name] = newQ
	m.quitChans[name] = quitChan
	m.wgs[name] = &wg
	m.Unlock()
	wg.Add(1)
	go func() {
		defer wg.Done()
		sigChan, unsub, err := newQ.Subscribe(queueManagerName)
		if err != nil {
			m.Logger.Error().Msgf("could not subscribe to %v queue: %v", name, err)
			return
		}
		defer unsub()
		for {
			select {
			case qLength := <-sigChan:
				if qLength == 0 {
					m.Logger.Info().Msgf("closing %v queue manager", name)
					return
				}
				for i := 0; i < qLength-1; i++ {
					frame := newQ.Pop()
					for _, outQ := range m.outgoingQueues[name] {
						outQ.Push(frame)
					}
				}
			case <-quitChan:
				m.Logger.Info().Msgf("closing %v queue manager", name)
				return
			}
		}
	}()
	deregister := func() {
		m.Lock()
		defer m.Unlock()
		close(m.quitChans[name])
		wg.Wait()
		delete(m.wgs, name)
		delete(m.quitChans, name)
		delete(m.incomingQueues, name)
	}
	return newQ, deregister, nil
}

// RegisterListener creates a new queue, registers a new listener in the list of outgoing queues, and returns that queue
func (m *QueueManager) RegisterListener(source, name string) (*Queue, func(), error) {
	if _, ok := m.outgoingQueues[source]; !ok {
		m.outgoingQueues[source] = map[string]*Queue{}
	}
	if _, ok := m.outgoingQueues[source][name]; ok {
		return nil, nil, ErrListenerAlreadyRegistered
	}
	newQ := NewQueue(defaultLisQueueSize)
	m.Lock()
	m.outgoingQueues[source][name] = newQ
	m.Unlock()
	deregister := func() {
		m.Lock()
		m.Unlock()
		delete(m.outgoingQueues[source], name)
	}
	return newQ, deregister, nil
}

// Shutdown shuts down all the queue listener go routines, closes all quit channels
func (m *QueueManager) Shutdown() {
	for name, quitChan := range m.quitChans {
		close(quitChan)
		wg := m.wgs[name]
		wg.Wait()
	}
}
