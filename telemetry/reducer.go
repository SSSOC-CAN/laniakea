/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package telemetry

import (
	"github.com/SSSOC-CAN/fmtd/errors"
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOCPaulCote/gux"
)

var (
	TelemetryInitialState = fmtrpc.RealTimeData{}
	TelemetryReducer gux.Reducer = func(s interface{}, a gux.Action) (interface{}, error) {
		// assert type of s
		_, ok := s.(fmtrpc.RealTimeData)
		if !ok {
			return nil, errors.ErrInvalidStateType
		}
		// switch case action
		switch a.Type {
		case "telemetry/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, errors.ErrInvalidPayloadType
			}
			return newState, nil
		default:
			return nil, errors.ErrInvalidAction
		} 
	}
)

