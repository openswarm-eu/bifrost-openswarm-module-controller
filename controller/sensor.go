package controller

import (
	"math"
)

type sensor struct {
	id            string
	limit         float64
	maximumFlow   float64
	pvDemand      float64
	chargerDemand float64
	flow          float64
	childSensors  []*sensor

	pvs          []*component
	numGlobalPVs int

	chargers          []*component
	numGlobalChargers int

	ignoreDuringDistribution bool
}

func (s *sensor) setSetPoints() {
	s.calculateMaximumFlow()
	s.iterateThroughChildren()
}

func (s *sensor) calculateMaximumFlow() {
	maximumFlow := 0.0

	for _, child := range s.childSensors {
		child.calculateMaximumFlow()
		maximumFlow += child.maximumFlow
		s.pvDemand += math.Min(child.limit, child.pvDemand)
		s.chargerDemand += math.Min(child.limit, child.chargerDemand)
	}

	localPVDemand := s.getLocalPVDemand(false)
	localChargerDemand := s.getLocalChargerDemand(false)
	maximumFlow -= localPVDemand
	s.pvDemand += localPVDemand
	maximumFlow += localChargerDemand
	s.chargerDemand += localChargerDemand

	if maximumFlow > 0 {
		s.maximumFlow = math.Min(s.limit, maximumFlow)
	} else {
		s.maximumFlow = math.Max(-s.limit, maximumFlow)
	}
}

func (s *sensor) iterateThroughChildren() {
	if s.maximumFlow > 0 {
		s.overconsumption()
	} else if s.maximumFlow < 0 {
		s.overproduction()
	} else {
		if s.chargerDemand > s.pvDemand {
			s.overconsumption()
		} else {
			s.overproduction()
		}
	}

}

func (s *sensor) overconsumption() {
	pvProduction := 0.0
	for _, childSensor := range s.childSensors {
		if childSensor.maximumFlow <= 0 {
			if -childSensor.maximumFlow < childSensor.limit {
				childSensor.flow = -childSensor.getPVDemand()
				pvProduction += -childSensor.flow
			} else {
				// maximum pv production is limited by sensor limit
				childSensor.ignoreDuringDistribution = true
				childSensor.flow = childSensor.maximumFlow
				pvProduction += -childSensor.flow
			}
		} else {
			if childSensor.maximumFlow < childSensor.limit {
				pvProduction += childSensor.getPVDemand()
			}
		}
	}

	if s.maximumFlow >= s.limit {
		for _, pv := range s.pvs {
			pvProduction += pv.demand - pv.setPoint
			pv.setPoint = pv.demand
		}
	}

	if s.ignoreDuringDistribution {
		pvProduction += s.flow
	}

	s.distributePVProduction(pvProduction)

	for _, child := range s.childSensors {
		child.iterateThroughChildren()
	}
}

func (s *sensor) overproduction() {
	chargerConsumption := 0.0
	for _, childSensor := range s.childSensors {
		if childSensor.maximumFlow >= 0 {
			if childSensor.maximumFlow < childSensor.limit {
				childSensor.flow = childSensor.getChargerDemand()
				chargerConsumption += childSensor.flow
			} else {
				// maximum charger consumption is limited by sensor limit
				childSensor.ignoreDuringDistribution = true
				childSensor.flow = childSensor.maximumFlow
				chargerConsumption += childSensor.flow
			}
		} else {
			if -childSensor.maximumFlow < childSensor.limit {
				chargerConsumption += childSensor.getChargerDemand()
			}
		}
	}

	if -s.maximumFlow >= s.limit {
		for _, charger := range s.chargers {
			chargerConsumption += charger.demand - charger.setPoint
			charger.setPoint = charger.demand
		}
	}

	if s.ignoreDuringDistribution {
		chargerConsumption += -s.flow
	}

	s.distributeChargerConsumption(chargerConsumption)

	for _, child := range s.childSensors {
		child.iterateThroughChildren()
	}
}

func (s *sensor) getPVDemand() float64 {
	if math.Abs(s.maximumFlow) >= s.limit {
		return 0
	}

	demand := 0.0

	for _, childSensor := range s.childSensors {
		demand += childSensor.getPVDemand()
	}

	return demand + s.getLocalPVDemand(true)
}

func (s *sensor) getLocalPVDemand(setSetpoint bool) float64 {
	demand := 0.0
	if setSetpoint {
		for _, pv := range s.pvs {
			demand += pv.demand - pv.setPoint
			pv.setPoint = pv.demand
		}
	} else {
		for _, pv := range s.pvs {
			demand += pv.demand - pv.setPoint
		}
	}

	return demand
}

func (s *sensor) getChargerDemand() float64 {
	if math.Abs(s.maximumFlow) >= s.limit {
		return 0
	}

	demand := 0.0
	for _, childSensor := range s.childSensors {
		demand += childSensor.getChargerDemand()
	}

	return demand + s.getLocalChargerDemand(true)
}

func (s *sensor) getLocalChargerDemand(setSetpoint bool) float64 {
	demand := 0.0
	if setSetpoint {
		for _, charger := range s.chargers {
			demand += charger.demand - charger.setPoint
			charger.setPoint = charger.demand
		}
	} else {
		for _, charger := range s.chargers {
			demand += charger.demand - charger.setPoint
		}
	}

	return demand
}

func (s *sensor) distributePVProduction(production float64) {
	for production > 0 {
		numberOfGlobalChargers := s.getNumberOfGlobalChargers()

		if numberOfGlobalChargers == 0 {
			// due to rounding errors, small production can be left over
			// in this case, we can just return
			return
		}

		chargerSetPoint := production / float64(numberOfGlobalChargers)
		for _, childSensor := range s.childSensors {
			if childSensor.maximumFlow <= 0 {
				production -= childSensor.setChargerSetPointsOverproduction(chargerSetPoint)
			} else {
				production -= childSensor.setChargerSetPointsOverconsumption(chargerSetPoint)
			}
		}

		for _, charger := range s.chargers {
			if charger.demand-charger.setPoint > chargerSetPoint {
				production -= chargerSetPoint
				charger.setPoint += chargerSetPoint
			} else {
				production -= charger.demand - charger.setPoint
				charger.setPoint += charger.demand - charger.setPoint
			}
		}
	}
}

func (s *sensor) distributeChargerConsumption(consumption float64) {
	for consumption > 0 {
		numberOfGlobalPVs := s.getNumberOfGlobalPVs()

		if numberOfGlobalPVs == 0 {
			// due to rounding errors, small consumption can be left over
			// in this case, we can just return
			return
		}

		chargerSetPoint := consumption / float64(numberOfGlobalPVs)
		for _, childSensor := range s.childSensors {
			if childSensor.maximumFlow >= 0 {
				consumption -= childSensor.setPVSetPointsOverconsumption(chargerSetPoint)
			} else {
				consumption -= childSensor.setPVSetPointsOverproduction(chargerSetPoint)
			}
		}

		for _, pv := range s.pvs {
			if pv.demand-pv.setPoint > chargerSetPoint {
				consumption -= chargerSetPoint
				pv.setPoint += chargerSetPoint
			} else {
				consumption -= pv.demand - pv.setPoint
				pv.setPoint += pv.demand - pv.setPoint
			}
		}
	}
}

func (s *sensor) getNumberOfGlobalChargers() int {
	numberOfGlobalChargers := 0
	for _, childSensor := range s.childSensors {
		numberOfGlobalChargers += childSensor.updateNumberOfGlobalChargers()
	}
	for _, charger := range s.chargers {
		if charger.demand > charger.setPoint {
			numberOfGlobalChargers++
		}
	}
	return numberOfGlobalChargers
}

func (s *sensor) updateNumberOfGlobalChargers() int {
	s.numGlobalChargers = 0

	if s.limit-s.flow == 0 || s.ignoreDuringDistribution {
		return 0
	}

	for _, childSensor := range s.childSensors {
		s.numGlobalChargers += childSensor.updateNumberOfGlobalChargers()
	}

	for _, charger := range s.chargers {
		if charger.demand > charger.setPoint {
			s.numGlobalChargers++
		}
	}

	return s.numGlobalChargers
}

func (s *sensor) getNumberOfGlobalPVs() int {
	numberOfGlobalPVs := 0
	for _, childSensor := range s.childSensors {
		numberOfGlobalPVs += childSensor.updateNumberOfGlobalPVs()
	}
	for _, pv := range s.pvs {
		if pv.demand > pv.setPoint {
			numberOfGlobalPVs++
		}
	}
	return numberOfGlobalPVs
}

func (s *sensor) updateNumberOfGlobalPVs() int {
	s.numGlobalPVs = 0

	if s.limit+s.flow == 0 || s.ignoreDuringDistribution {
		return 0
	}

	for _, childSensor := range s.childSensors {
		s.numGlobalPVs += childSensor.updateNumberOfGlobalPVs()
	}

	for _, pv := range s.pvs {
		if pv.demand > pv.setPoint {
			s.numGlobalPVs++
		}
	}

	return s.numGlobalPVs
}

func (s *sensor) setChargerSetPointsOverconsumption(chargerSetPoint float64) float64 {
	usedProduction := 0.0
	if s.numGlobalChargers == 0 {
		return 0
	}

	maxPossibleFlow := s.limit - s.flow
	childChargerSetPoint := math.Min(chargerSetPoint, maxPossibleFlow/float64(s.numGlobalChargers))

	for _, childSensor := range s.childSensors {
		usedProduction += childSensor.setChargerSetPointsOverconsumption(childChargerSetPoint)
		s.flow += usedProduction
	}

	for _, charger := range s.chargers {
		if charger.demand-charger.setPoint > childChargerSetPoint {
			usedProduction += childChargerSetPoint
			charger.setPoint += childChargerSetPoint
			s.flow += childChargerSetPoint
		} else {
			usedProduction += charger.demand - charger.setPoint
			charger.setPoint += charger.demand - charger.setPoint
			s.flow += charger.demand - charger.setPoint
		}
	}
	return usedProduction
}

func (s *sensor) setChargerSetPointsOverproduction(chargerSetPoint float64) float64 {
	usedProduction := 0.0

	if s.ignoreDuringDistribution {
		return 0
	}

	for _, childSensor := range s.childSensors {
		usedProduction += childSensor.setChargerSetPointsOverproduction(chargerSetPoint)
	}

	for _, charger := range s.chargers {
		if charger.demand-charger.setPoint > chargerSetPoint {
			usedProduction += chargerSetPoint
			charger.setPoint += chargerSetPoint
		} else {
			usedProduction += charger.demand - charger.setPoint
			charger.setPoint += charger.demand - charger.setPoint
		}
	}

	return usedProduction
}

func (s *sensor) setPVSetPointsOverproduction(pvSetPoint float64) float64 {
	usedConsumption := 0.0
	if s.numGlobalPVs == 0 {
		return 0
	}

	maxPossibleFlow := s.limit + s.flow
	childPVSetPoint := math.Min(pvSetPoint, maxPossibleFlow/float64(s.numGlobalPVs))

	for _, childSensor := range s.childSensors {
		usedConsumption += childSensor.setPVSetPointsOverproduction(childPVSetPoint)
		s.flow -= usedConsumption
	}

	for _, pv := range s.pvs {
		if pv.demand-pv.setPoint > childPVSetPoint {
			usedConsumption += childPVSetPoint
			pv.setPoint += childPVSetPoint
			s.flow -= childPVSetPoint
		} else {
			usedConsumption += pv.demand - pv.setPoint
			pv.setPoint += pv.demand - pv.setPoint
			s.flow -= (pv.demand - pv.setPoint)
		}
	}
	return usedConsumption
}

func (s *sensor) setPVSetPointsOverconsumption(pvSetPoint float64) float64 {
	usedConsumption := 0.0

	if s.ignoreDuringDistribution {
		return 0
	}

	for _, childSensor := range s.childSensors {
		usedConsumption += childSensor.setPVSetPointsOverconsumption(pvSetPoint)
	}

	for _, pv := range s.pvs {
		if pv.demand-pv.setPoint > pvSetPoint {
			usedConsumption += pvSetPoint
			pv.setPoint += pvSetPoint
		} else {
			usedConsumption += pv.demand - pv.setPoint
			pv.setPoint += pv.demand - pv.setPoint
		}
	}

	return usedConsumption
}

func (s *sensor) reset() {
	for _, childSensor := range s.childSensors {
		childSensor.reset()
	}

	s.flow = 0
	s.maximumFlow = 0
	s.ignoreDuringDistribution = false

	for _, pv := range s.pvs {
		pv.setPoint = 0
	}
	for _, charger := range s.chargers {
		charger.setPoint = 0
	}
}

func (s *sensor) updateNumberOfGlobalPVsForFlowProposal() int {
	numChildPvs := 0
	for _, childSensor := range s.childSensors {
		numChildPvs += childSensor.updateNumberOfGlobalPVsForFlowProposal()
	}

	if s.flow >= 0 {
		s.numGlobalPVs = 0
		return 0
	}

	s.numGlobalPVs = numChildPvs
	for _, pv := range s.pvs {
		if pv.setPoint > 0 {
			s.numGlobalPVs++
		}
	}

	return s.numGlobalPVs
}

func (s *sensor) updateNumberOfGlobalChargerssForFlowProposal() int {
	numChildChargers := 0
	for _, childSensor := range s.childSensors {
		numChildChargers += childSensor.updateNumberOfGlobalChargerssForFlowProposal()
	}

	if s.flow <= 0 {
		s.numGlobalChargers = 0
		return 0
	}

	s.numGlobalChargers = numChildChargers
	for _, charger := range s.chargers {
		if charger.setPoint > 0 {
			s.numGlobalChargers++
		}
	}

	return s.numGlobalChargers
}
