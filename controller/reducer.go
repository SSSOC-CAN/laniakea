/*
Author: Paul Côté
Last Change Author: Paul Côté
Last Date Changed: 2022/06/10
*/

package controller

import (
	"github.com/SSSOC-CAN/fmtd/data"
	"github.com/SSSOC-CAN/fmtd/state"
)

var (
	InitialState = data.InitialCtrlState{
		PressureSetPoint: float64(760.0),
		TemperatureSetPoint: float64(25.0),
	}
	ControllerReducer state.Reducer = func(s interface{}, a state.Action) (interface{}, error) {
		// assert type of s
		oldState, ok := s.(data.InitialCtrlState)
		if !ok {
			return nil, state.ErrInvalidStateType
		}
		// switch case action
		switch a.Type {
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
