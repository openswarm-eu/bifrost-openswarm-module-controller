package controller

import "math"

type sensor struct {
	id           string
	sensorLimit  float64
	parent       *sensor
	childSensors []*sensor
	pvs          []*component
	chargers     []*component

	virtualComponent virtualComponent
}

func (s *sensor) setSetPoints() {
	if s.isGlobalOverConsumptionPresent() {
		// distribute available pv demand equally to chargers
		for {
			demandMet := false

			for _, child1 := range s.childSensors {
				numberOfChargers := s.numChargersWithCapacity()
				if numberOfChargers == 0 {
					continue
				}

				chargerSetPoint := child1.getChargerSetPoint(numberOfChargers)
				if chargerSetPoint == 0 {
					continue
				}

				remoteChargerDemandMet := 0.0
				localChargerDemandMet := 0.0
				for _, child2 := range s.childSensors {
					if child1 == child2 {
						localChargerDemandMet += child2.updateLocalChargerSetPoint(chargerSetPoint)
					} else {
						chargerDemandMet := child2.updateRemoteChargerSetPoints(chargerSetPoint)
						child2.setConsumption(chargerDemandMet) //set virtual node
						remoteChargerDemandMet += chargerDemandMet
					}
				}

				if remoteChargerDemandMet+localChargerDemandMet > 0 {
					demandMet = true
				}

				child1.distributeMetDemandToPVs(remoteChargerDemandMet + localChargerDemandMet)
				child1.setProduction(remoteChargerDemandMet) //set virtual node
			}

			if !demandMet {
				break
			}
		}
	} else {
		// distribute available charger demand equally to pvs
		for {
			demandMet := false

			for _, child1 := range s.childSensors {
				numberOfPVs := s.numPVsWithCapacity()
				if numberOfPVs == 0 {
					continue
				}

				pvSetPoint := child1.getPVSetPoint(numberOfPVs)
				if pvSetPoint == 0 {
					continue
				}

				remotePVDemandMet := 0.0
				localPVDemandMet := 0.0
				for _, child2 := range s.childSensors {
					if child1 == child2 {
						localPVDemandMet += child2.updateLocalPVSetPoints(pvSetPoint)
					} else {
						pvDemandMet := child2.updateRemotePVSetPoints(pvSetPoint)
						child2.setProduction(pvDemandMet) //set virtual node
						remotePVDemandMet += pvDemandMet
					}
				}

				if remotePVDemandMet+localPVDemandMet > 0 {
					demandMet = true
				}

				child1.distributeMetDemandToChargers(remotePVDemandMet + localPVDemandMet)
				child1.setConsumption(remotePVDemandMet) //set virtual node
			}

			if !demandMet {
				break
			}
		}
	}

	s.distributeOpenDemandLocally()

	// virtual components are set, distribute energy
	for _, child := range s.childSensors {
		child.setSetPoints()
	}
}

func (s *sensor) isGlobalOverConsumptionPresent() bool {
	globalFlow := 0.0
	for _, childSensor := range s.childSensors {
		globalFlow += childSensor.calculateLocalFlow()
	}

	return globalFlow >= 0
}

func (s *sensor) calculateLocalFlow() float64 {
	localFlow := 0.0
	for _, childSensor := range s.childSensors {
		localFlow += childSensor.calculateLocalFlow()
	}

	for _, pv := range s.pvs {
		localFlow -= pv.demand
	}

	for _, charger := range s.chargers {
		localFlow += charger.demand
	}

	if localFlow > 0 {
		return math.Min(s.sensorLimit, localFlow)
	} else {
		return math.Max(-s.sensorLimit, localFlow)
	}
}

func (s *sensor) getChargerSetPoint(numberOfChargers int) float64 {
	openDemand := s.getOpenPVsDemand()
	numberLocalChargers := s.numLocalChargersWithCapacity()

	fairChargerSetPoint := openDemand / float64(numberOfChargers)
	availableDemand := openDemand - fairChargerSetPoint*float64(numberLocalChargers)
	if s.sensorLimit-s.virtualComponent.getFlexibility() > availableDemand {
		return fairChargerSetPoint
	} else {
		return (s.sensorLimit - s.virtualComponent.getFlexibility()) / float64(numberOfChargers-numberLocalChargers)
	}
}

func (s *sensor) getPVSetPoint(numberOfPVs int) float64 {
	openDemand := s.getOpenChargersDemand()
	numberLocalPVs := s.numLocalPVsWithCapacity()

	fairPVSetPoint := openDemand / float64(numberOfPVs)
	availableDemand := openDemand - fairPVSetPoint*float64(numberLocalPVs)
	if s.sensorLimit+s.virtualComponent.getFlexibility() > availableDemand {
		return fairPVSetPoint
	} else {
		return (s.sensorLimit + s.virtualComponent.getFlexibility()) / float64(numberOfPVs-numberLocalPVs)
	}
}

func (s *sensor) distributeOpenDemandLocally() {
	production := s.getOpenPVsDemand()
	consumption := s.getOpenChargersDemand()

	if production >= consumption {
		for _, charger := range s.chargers {
			charger.setPoint = charger.demand
		}

		s.distributeMetDemandToPVs(consumption)
	} else {
		for _, pv := range s.pvs {
			pv.setPoint = pv.demand
		}

		s.distributeMetDemandToChargers(production)
	}
}

func (s *sensor) getOpenPVsDemand() float64 {
	openDemand := 0.0

	/*for _, childSensor := range s.childSensors {
		production += childSensor.getProduction()
	}*/

	for _, pv := range s.pvs {
		openDemand += pv.demand - pv.setPoint
	}
	return openDemand
}

func (s *sensor) getOpenChargersDemand() float64 {
	openDemand := 0.0

	/*for _, childSensor := range s.childSensors {
		consumption += childSensor.getConsumption()
	}*/

	for _, charger := range s.chargers {
		openDemand += charger.demand - charger.setPoint
	}
	return openDemand
}

func (s *sensor) numChargersWithCapacity() int {
	numberOfChargers := 0

	for _, childSensor := range s.childSensors {
		numberOfChargers += childSensor.numChargersWithCapacity()
	}

	if s.sensorLimit+s.virtualComponent.getFlexibility() == 0 {
		return numberOfChargers
	}

	for _, charger := range s.chargers {
		if charger.demand > charger.setPoint {
			numberOfChargers++
		}
	}

	return numberOfChargers
}

func (s *sensor) numLocalChargersWithCapacity() int {
	numberOfChargers := 0

	for _, childSensor := range s.childSensors {
		numberOfChargers += childSensor.numLocalChargersWithCapacity()
	}

	for _, charger := range s.chargers {
		if charger.demand > charger.setPoint {
			numberOfChargers++
		}
	}

	return numberOfChargers
}

func (s *sensor) numPVsWithCapacity() int {
	numberOfPVs := 0

	for _, childSensor := range s.childSensors {
		numberOfPVs += childSensor.numPVsWithCapacity()
	}

	for _, pv := range s.pvs {
		if pv.demand > pv.setPoint {
			numberOfPVs++
		}
	}

	return numberOfPVs
}

func (s *sensor) numLocalPVsWithCapacity() int {
	numberOfPVs := 0

	for _, childSensor := range s.childSensors {
		numberOfPVs += childSensor.numLocalPVsWithCapacity()
	}

	for _, pv := range s.pvs {
		if pv.demand > pv.setPoint {
			numberOfPVs++
		}
	}

	return numberOfPVs
}

func (s *sensor) setProduction(flow float64) {
	s.virtualComponent.consumeFlexibility(-flow)
}

func (s *sensor) setConsumption(flow float64) {
	s.virtualComponent.consumeFlexibility(flow)
}

func (s *sensor) updateRemoteChargerSetPoints(setPoint float64) float64 {
	consumption := 0.0
	for _, charger := range s.chargers {
		consumption += math.Min(setPoint, charger.demand-charger.setPoint)
	}

	consumption = math.Min(s.sensorLimit+s.virtualComponent.getFlexibility(), consumption)
	s.distributeMetDemandToChargers(consumption)

	return consumption
}

func (s *sensor) updateRemotePVSetPoints(setPoint float64) float64 {
	metDemand := 0.0
	for _, pv := range s.pvs {
		metDemand += math.Min(setPoint, pv.demand-pv.setPoint)
	}

	metDemand = math.Min(s.sensorLimit+s.virtualComponent.getFlexibility(), metDemand)
	s.distributeMetDemandToPVs(metDemand)

	return metDemand
}

func (s *sensor) updateLocalChargerSetPoint(setPoint float64) float64 {
	metDemand := 0.0
	for _, charger := range s.chargers {
		metDemand += math.Min(setPoint, charger.demand-charger.setPoint)
		charger.setPoint = math.Min(charger.demand, charger.setPoint+setPoint)
	}
	return metDemand
}

func (s *sensor) updateLocalPVSetPoints(setPoint float64) float64 {
	metDemand := 0.0
	for _, pv := range s.pvs {
		metDemand += math.Min(setPoint, pv.demand-pv.setPoint)
		pv.setPoint = math.Min(pv.demand, pv.setPoint+setPoint)
	}
	return metDemand
}

func (s *sensor) distributeMetDemandToPVs(demand float64) {
	pvCount := s.numLocalPVsWithCapacity()
	for pvCount > 0 && demand > 0 {
		setPoint := demand / float64(pvCount)
		for _, pv := range s.pvs {
			if pv.demand-pv.setPoint >= setPoint {
				pv.setPoint += setPoint
				demand -= setPoint
			} else {
				demand -= pv.demand - pv.setPoint
				pv.setPoint += pv.demand - pv.setPoint
				pvCount--
			}
		}
	}
}

func (s *sensor) distributeMetDemandToChargers(demand float64) {
	chargerCount := s.numLocalChargersWithCapacity()
	for chargerCount > 0 && demand > 0 {
		setPoint := demand / float64(chargerCount)
		for _, charger := range s.chargers {
			if charger.demand-charger.setPoint >= setPoint {
				charger.setPoint += setPoint
				demand -= setPoint
			} else {
				demand -= charger.demand - charger.setPoint
				charger.setPoint += charger.demand - charger.setPoint
				chargerCount--
			}
		}
	}
}

func (s *sensor) reset() {
	for _, childSensor := range s.childSensors {
		childSensor.reset()
	}

	s.virtualComponent.possibleFlexibility = 0

	for _, pv := range s.pvs {
		pv.setPoint = 0
	}
	for _, charger := range s.chargers {
		charger.setPoint = 0
	}
}
