package controller

import (
	"log"
	"time"

	"code.siemens.com/energy-community-controller/ddaConnector"
)

type equalAllocationAlgorithm struct{}

func (equalAllocationAlgorithm) calculateChargerPower(pvProductionValues []ddaConnector.Value, chargers []ddaConnector.Message) []ddaConnector.Value {
	log.Println(pvProductionValues)
	log.Println(chargers)

	var sumPvProduction int
	for _, productionValue := range pvProductionValues {
		sumPvProduction += productionValue.Value
	}

	var chargingSetPoint int
	numChargers := len(chargers)
	if numChargers > 0 {
		chargingSetPoint = sumPvProduction / len(chargers)
	} else {
		chargingSetPoint = 0
	}

	result := make([]ddaConnector.Value, len(chargers))

	for i, charger := range chargers {
		result[i] = ddaConnector.Value{Message: ddaConnector.Message{Id: charger.Id, Timestamp: time.Now()}, Value: chargingSetPoint}
	}

	return result
}
