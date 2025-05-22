package controller

import "github.com/coatyio/dda/services/com/api"

type state struct {
	leader          bool
	registeredAtDso bool
	sensors         map[string]*sensor
	chargers        map[string]*component
	pvs             map[string]*component
	rootSensor      *sensor

	registerCallbacks   map[string]api.ActionCallback
	deregisterCallbacks map[string]api.ActionCallback
	clusterMembers      int
}

type component struct {
	id string

	demand   float64
	setPoint float64
}
