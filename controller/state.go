package controller

import (
	"code.siemens.com/energy-community-controller/common"
)

type state struct {
	pvProductionValues []common.Value
	chargerRequests    []common.Value
	setPoints          []common.Value
	sensors            map[string]*sensor
	rootSensor         *sensor
}

type component struct {
	id string

	demand   float64
	setPoint float64
}
