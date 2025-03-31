package sct

import (
	"slices"
)

type supervisor struct {
	currentState state
	events       []event
}

func (su *supervisor) changeState(event event) {
	if newState, ok := su.currentState.transitions[event]; ok {
		su.currentState = newState
	}
}

func (su *supervisor) getActiveEvents() []event {
	var controllableEvents []event
	for event := range su.currentState.transitions {
		if event.controllable {
			controllableEvents = append(controllableEvents, event)
		}
	}
	return controllableEvents
}

func (su *supervisor) isEventPresent(event event) bool {
	return slices.Contains(su.events, event)
}
