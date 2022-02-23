package state

import (
	"sync"
)

type (
	Reducer func(interface{}, Action) (interface{}, error)

	Action struct {
		Type    string
		Payload interface{}
	}

	Store struct {
		mutex     sync.RWMutex
		state     interface{}
		reducer   Reducer
		listeners []func()
		unsub     func(*Store, int)
	}
)

// CreateStore creates a new state store object
func CreateStore(initialState interface{}, rootReducer Reducer) *Store {
	return &Store{
		state:   initialState,
		reducer: rootReducer,
		unsub: func(s *Store, i int) {
			var ls []func()
			for j, s := range s.listeners {
				if j != i {
					ls = append(ls, s)
				}
			}
			s.listeners = ls
		},
	}
}

// GetState returns the current state object
func (s *Store) GetState() interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.state
}

// Dispatch takes an action and returns an error. It is the only way to change the state
func (s *Store) Dispatch(action Action) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	newState, err := s.reducer(s.state, action)
	if err != nil {
		return err
	}
	s.state = newState
	// Update subscribers on state change
	for _, l := range s.listeners {
		l()
	}
	return nil
}

// Subscribe adds a callback function to the list of listeners which will be executed upon each Dispatch call.
// Returns the index in the listener slice belonging to callback and unsubscribe function
func (s *Store) Subscribe(f func()) (int, func(*Store, int)) {
	s.listeners = append(s.listeners, f)
	return len(s.listeners) - 1, s.unsub
}

// CombineReducers combines any number of reducers and returns one combined reducer
func CombineReducers(reducers ...Reducer) Reducer {
	var combinedReducer Reducer = func(s interface{}, a Action) (interface{}, error) {
		newState := make([]interface{}, len(reducers))
		for i, r := range reducers {
			newS, err := r(s, a)
			if err != nil {
				return nil, err
			}
			newState[i] = newS
		}
		return newState, nil
	}
	return combinedReducer
}