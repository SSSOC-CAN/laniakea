package drivers

import (
	"fmt"
	"reflect"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/state"
)

var (
	RGAInitialState = fmtrpc.RealTimeData{}
	RGAReducer state.Reducer = func(s interface{}, a state.Action) (interface{}, error) {
		// assert type of s
		_, ok := s.(fmtrpc.RealTimeData)
		if !ok {
			return nil, fmt.Errorf("Invalid state type %v: expected type %v", reflect.TypeOf(s), reflect.TypeOf(fmtrpc.RealTimeData{}))
		}
		// switch case action
		switch a.Type {
		case "rga/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, fmt.Errorf("Invalid payload type %v: expected type %v", reflect.TypeOf(a.Payload), reflect.TypeOf(fmtrpc.RealTimeData{}))
			}
			return newState, nil
		default:
			return nil, fmt.Errorf("Invalid action: %v", a.Type)
		} 
	}
)

