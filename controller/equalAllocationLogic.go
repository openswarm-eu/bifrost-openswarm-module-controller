package controller

import (
	"log"
	"time"

	"code.siemens.com/energy-community-controller/dda"
)

type equalAllocationAlgorithm struct{}

func (equalAllocationAlgorithm) calculateChargerPower(pvProductionValues []dda.Value, chargers []dda.Message) []dda.Value {
	log.Println(pvProductionValues)
	log.Println(chargers)

	var sumPvProduction float64
	for _, productionValue := range pvProductionValues {
		sumPvProduction += productionValue.Value
	}

	var chargingSetPoint float64
	numChargers := len(chargers)
	if numChargers > 0 {
		chargingSetPoint = sumPvProduction / float64(len(chargers))
	} else {
		chargingSetPoint = 0
	}

	result := make([]dda.Value, len(chargers))

	for i, charger := range chargers {
		result[i] = dda.Value{Message: dda.Message{Id: charger.Id, Timestamp: time.Now()}, Value: chargingSetPoint}
	}

	return result
}
