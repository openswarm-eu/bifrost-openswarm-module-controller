package sct

import (
	"context"
	"io"
	"log"
	"math/rand"
	"slices"
)

type SCT struct {
	supervisors       []*supervisor
	callbacks         map[string]func()
	eventsLookupTable map[string]event

	eventChannel chan string
}

func NewSCT(xmlDefinitions []io.Reader, callbacks map[string]func()) (*SCT, error) {
	sct := &SCT{
		supervisors:       []*supervisor{},
		callbacks:         callbacks,
		eventsLookupTable: make(map[string]event),
		eventChannel:      make(chan string, 10),
	}

	for _, xmlDefinition := range xmlDefinitions {
		model, err := parseXML(xmlDefinition)
		if err != nil {
			return nil, err
		}

		eventList := make([]event, 0)
		eventIdLookupTable := make(map[string]event)
		for _, e := range model.Data.Events {
			event := event{name: e.Name, controllable: e.Controllable == "True"}
			eventIdLookupTable[e.ID] = event
			sct.eventsLookupTable[e.Name] = event
			eventList = append(eventList, event)
		}

		intialState := sct.createInitialState(model.Data, eventIdLookupTable)

		sct.supervisors = append(sct.supervisors, &supervisor{intialState, eventList})
	}

	return sct, nil
}

func (sct *SCT) Start(context context.Context) {
	go func() {
		for {
			select {
			case event := <-sct.eventChannel:
				sct.processEvent(event)
			case <-context.Done():
				log.Printf("SCT - shutdown event channel observer")
				return
			}
		}
	}()
}

func (sct *SCT) AddEvent(event string) {
	sct.eventChannel <- event
}

func (sct *SCT) processEvent(event string) {
	ev, ok := sct.eventsLookupTable[event]
	if !ok {
		log.Println("SCT - Unknown event:", event)
		return
	}

	log.Println("SCT - Processing event:", event)

	for _, su := range sct.supervisors {
		su.changeState(ev)
	}

	for {
		controllableEvent, found := sct.getNextControllableEvent()

		if !found {
			return
		}

		log.Println("SCT - Controllable event:", controllableEvent.name)

		for _, su := range sct.supervisors {
			su.changeState(controllableEvent)
		}

		if cb, ok := sct.callbacks[controllableEvent.name]; ok {
			cb()
		} else {
			log.Printf("SCT - Event %s not found in callback map", controllableEvent.name)
		}
	}
}

func (sct *SCT) createInitialState(data Data, eventIdLookupTable map[string]event) state {
	for _, s := range data.States {
		if s.Initial == "True" {
			return sct.createState(s.ID, data.Transitions, make(map[string]state), eventIdLookupTable)
		}
	}

	return state{}
}

func (sct *SCT) createState(id string, transitions []Transition, existingStates map[string]state, eventIdLookupTable map[string]event) state {
	if _, present := existingStates[id]; present {
		return existingStates[id]
	}

	newState := state{transitions: make(map[event]state)}
	existingStates[id] = newState

	for _, transition := range transitions {
		if transition.Source == id {
			evt := eventIdLookupTable[transition.Event]
			newState.transitions[evt] = sct.createState(transition.Target, transitions, existingStates, eventIdLookupTable)
		}
	}

	return newState
}

func (sct *SCT) getNextControllableEvent() (event, bool) {
	activeEvents := make([]event, 0)

	supervisorActiveEvents := make(map[*supervisor][]event)

	for _, su := range sct.supervisors {
		supervisorActiveEvents[su] = su.getActiveEvents()
	}

	for _, event := range sct.eventsLookupTable {
		if !event.controllable {
			continue
		}

		isActive := true

		for _, su := range sct.supervisors {
			if !su.isEventPresent(event) {
				continue
			}

			if slices.Contains(supervisorActiveEvents[su], event) {
				continue
			}

			isActive = false
			break
		}

		if isActive {
			activeEvents = append(activeEvents, event)
		}
	}

	switch len(activeEvents) {
	case 0:
		return event{}, false
	case 1:
		return activeEvents[0], true
	default:
		return activeEvents[rand.Intn(len(activeEvents))], true
	}
}
