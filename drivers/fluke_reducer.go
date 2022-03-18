// +build fluke windows 386

package drivers

import (
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/state"
)

var (
	FlukeInitialState = fmtrpc.RealTimeData{}
	FlukeReducer state.Reducer = func(s interface{}, a state.Action) (interface{}, error) {
		// assert type of s
		_, ok := s.(fmtrpc.RealTimeData)
		if !ok {
			return nil, state.ErrInvalidStateType
		}
		// switch case action
		switch a.Type {
		case "fluke/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			return newState, nil
		default:
			return nil, state.ErrInvalidAction
		} 
	}
)

