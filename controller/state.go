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

type virtualComponent struct {
	possibleFlexibility float64
}

func (v *virtualComponent) getFlexibility() float64 {
	return v.possibleFlexibility
}

// positive flexility = charger, negative flexibility = pv
func (v *virtualComponent) consumeFlexibility(flexibility float64) {
	v.possibleFlexibility -= flexibility
}

func (v *virtualComponent) reset() {
	v.possibleFlexibility = 0
}
