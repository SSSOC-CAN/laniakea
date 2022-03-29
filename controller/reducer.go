package controller

import (
	"github.com/SSSOC-CAN/fmtd/fmtrpc"
	"github.com/SSSOC-CAN/fmtd/state"
)

type ControllerInitialState struct {
	PressureSetPoint	float64
	TemperatureSetPoint	float64
	RealTimeData		fmtrpc.RealTimeData
}

var (
	InitialState := ControllerInitialState{
		PressureSetPoint: float64(760.2095),
		TemperatureSetPoint: float64(25.0),
		RealTimeData: fmtrpc.RealTimeData{},
	}
	ControllerReducer state.Reducer = func(s interface{}, a state.Action) (interface{}, error) {
		// assert type of s
		oldState, ok := s.(ControllerInitialState)
		if !ok {
			return nil, state.ErrInvalidStateType
		}
		// switch case action
		switch a.Type {
		case "telemetry/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.RealTimeData = newState
			return oldState, nil
		case "rga/update":
			// assert type of payload
			newState, ok := a.Payload.(fmtrpc.RealTimeData)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.RealTimeData = newState
			return oldState, nil
		case "setpoint/temperature/update":
			// assert type of payload
			newTemp, ok := a.Payload.(float64)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.TemperatureSetPoint = newTemp
			return oldState, nil
		case "setpoint/pressure/update":
			// assert type of payload
			newPres, ok := a.Payload.(float64)
			if !ok {
				return nil, state.ErrInvalidPayloadType
			}
			oldState.PressureSetPoint = newPres
			return oldState, nil
		default:
			return nil, state.ErrInvalidAction
		} 
	}
	updateTempSetPointAction = func(newTempSetPoint float64) state.Action {
		return state.Action{
			Type: "setpoint/temperature/update",
			Payload: newTempSetPoint,
		}
	}
	updatePresSetPointAction = func(newPresSetPoint float64) state.Action {
		return state.Action{
			Type: "setpoint/pressure/update",
			Payload: newPresSetPoint,
		}
	}
)
