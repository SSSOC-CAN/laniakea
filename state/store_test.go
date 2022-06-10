/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package state

import (
	"reflect"
	"sync"
	"testing"
	"time"
)

var (
	testReducer Reducer = func(v interface{}, a Action) (interface{}, error) {
		// assert v and a.Payload are integers
		cState, ok := v.(int)
		if !ok {
			return nil, ErrInvalidStateType
		}
		payload, ok := a.Payload.(int)
		if !ok {
			return nil, ErrInvalidPayloadType
		}
		switch a.Type {
		case "increment":
			return cState+payload, nil
		case "decrement":
			return cState-payload, nil
		default:
			return nil, ErrInvalidAction
		}
	} 
	testInitialState int = 0
	testIncrementAction = Action{
		Type: "increment",
		Payload: 1,
	}
	testDecrementAction = Action{
		Type: "decrement",
		Payload: 1,
	}
	listenerName string = "test"
)

// TestStore will test the state store against many different scenarios
func TestStore(t *testing.T) {
	store := CreateStore(testInitialState, testReducer)
	// increment and assert state change
	t.Run("valid increment", func(t *testing.T) {
		err := store.Dispatch(testIncrementAction)
		if err != nil {
			t.Errorf("Unexpected error when dispatching: %v", err)
		}
		currentState := store.GetState()
		cState, ok := currentState.(int)
		if !ok {
			t.Errorf("Invalid type, expected int, got: %v", reflect.TypeOf(cState))
		}
		if cState != 1 {
			t.Errorf("Unexpected state, expected 1 got: %v", cState)
		}
	})
	// valid decrement
	t.Run("valid decrement", func(t *testing.T) {
		err := store.Dispatch(testDecrementAction)
		if err != nil {
			t.Errorf("Unexpected error when dispatching: %v", err)
		}
		currentState := store.GetState()
		cState, ok := currentState.(int)
		if !ok {
			t.Errorf("Invalid type, expected int, got: %v", reflect.TypeOf(cState))
		}
		if cState != 0 {
			t.Errorf("Unexpected state, expected 1 got: %v", cState)
		}
	})
	// invalid action
	t.Run("invalid action", func(t *testing.T) {
		err := store.Dispatch(Action{Type: "invalid", Payload: 2000})
		if err != ErrInvalidAction {
			t.Errorf("Unexpected error: %v", err)
		}
	})
	// invalid payload type
	t.Run("invalid payload", func(t *testing.T) {
		err := store.Dispatch(Action{Type: "increment", Payload: int64(1)})
		if err != ErrInvalidPayloadType {
			t.Errorf("Unexpected error: %v", err)
		}
	})
	// testing subscribe and unsubscribe
	t.Run("subscribe unsubscribe", func(t *testing.T) {
		var wg sync.WaitGroup
		stateChangeChan, unsub := store.Subscribe(listenerName)
		ticker := time.NewTicker(time.Duration(int64(2)) * time.Second)
		defer ticker.Stop()
		// We should get a state change update before the first tick
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ticker.C:
				t.Errorf("Timeout error: received ticker signal before stateChange")
			case <-stateChangeChan:
			}
			return
		}()
		err := store.Dispatch(testIncrementAction)
		if err != nil {
			t.Errorf("Unexpected error when dispatching: %v", err)
		}
		wg.Wait()
		// now we shouldn't get any state change updates and instead get a tick signal
		unsub(store, listenerName)
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-ticker.C:
			case <-stateChangeChan:
				t.Errorf("Error: received stateChange signal")
			}
			return
		}()
		wg.Wait()
	})
	// test for concurrent write/read
	t.Run("concurrency", func(t *testing.T) {
		var wg sync.WaitGroup
		for i := 0; i < 100; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := store.Dispatch(testIncrementAction)
				if err != nil {
					t.Errorf("Unexpected error when dispatching: %v", err)
				}
				currentState := store.GetState()
				cState, ok := currentState.(int)
				if !ok {
					t.Errorf("Invalid type, expected int, got: %v", reflect.TypeOf(cState))
				}
				t.Logf("State: %v", cState)
			}()
		}
		wg.Wait()
		currentState := store.GetState()
		cState, ok := currentState.(int)
		if !ok {
			t.Errorf("Invalid type, expected int, got: %v", reflect.TypeOf(cState))
		}
		if cState != 101 {
			t.Errorf("Unexpected state value: %v", cState)
		}
	})
}