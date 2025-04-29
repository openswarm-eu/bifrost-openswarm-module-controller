package controller

import (
	"math"
	"testing"
)

const tolerance = .00001

func TestConsumptionOutgoingLimit(t *testing.T) {
	root := sensor{id: "root", sensorLimit: 0, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor1 := sensor{id: "sensor1", sensorLimit: 20, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor2 := sensor{id: "sensor2", sensorLimit: 20, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor3 := sensor{id: "sensor3", sensorLimit: 2, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, component{id: "charger1", demand: 10, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, component{id: "charger2", demand: 7, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, component{id: "charger3", demand: 2, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, component{id: "pv1", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, component{id: "charger4", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, component{id: "pv2", demand: 5, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, component{id: "pv3", demand: 5, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 1, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 1, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 1, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-4) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 4, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 3, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-3) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 3, got %f", sensor3.pvs[1].setPoint)
	}

	// flipped order of sensors
	root = sensor{sensorLimit: 0, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	root.childSensors = append(root.childSensors, &sensor3)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor1)
	root.reset()

	root.setSetPoints()
	if math.Abs(sensor1.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 1, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 1, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 1, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-4) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 4, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 3, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-3) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 3, got %f", sensor3.pvs[1].setPoint)
	}

}

func TestGlobalOverConsumptionRequestedPowerLimit(t *testing.T) {
	root := sensor{id: "root", sensorLimit: 0, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor1 := sensor{id: "sensor1", sensorLimit: 20, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor2 := sensor{id: "sensor2", sensorLimit: 20, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor3 := sensor{id: "sensor3", sensorLimit: 20, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, component{id: "charger1", demand: 10, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, component{id: "charger2", demand: 0.5, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, component{id: "charger3", demand: 5, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, component{id: "pv1", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, component{id: "charger4", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, component{id: "pv2", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, component{id: "pv3", demand: 3, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 2.5, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-0.5) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 0.5, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 2.5, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 2.5, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-4) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 4, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-3) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 3, got %f", sensor3.pvs[1].setPoint)
	}

	// flipped order of sensors
	root = sensor{sensorLimit: 0, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	root.childSensors = append(root.childSensors, &sensor3)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor1)
	root.reset()

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 2.5, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-0.5) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 0.5, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 2.5, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 2.5, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-4) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 4, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-3) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 3, got %f", sensor3.pvs[1].setPoint)
	}
}

func TestGlobalOverConsumptionIngoingLimit(t *testing.T) {
	root := sensor{id: "root", sensorLimit: 0, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor1 := sensor{id: "sensor1", sensorLimit: 1, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor2 := sensor{id: "sensor2", sensorLimit: 20, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor3 := sensor{id: "sensor3", sensorLimit: 20, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, component{id: "charger1", demand: 10, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, component{id: "charger2", demand: 7, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, component{id: "charger3", demand: 2, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, component{id: "pv1", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, component{id: "charger4", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, component{id: "pv2", demand: 2, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, component{id: "pv3", demand: 3, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-0.5) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 0.5, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-0.5) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 0.5, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 2, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 3, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 2, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-3) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 3, got %f", sensor3.pvs[1].setPoint)
	}

	// flipped order of sensors
	root = sensor{sensorLimit: 0, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)
	root.reset()

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-0.5) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 0.5, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-0.5) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 0.5, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 2, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 3, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 2, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-3) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 3, got %f", sensor3.pvs[1].setPoint)
	}
}

func TestGlobalOverProductionIngoingLimit(t *testing.T) {
	root := sensor{id: "root", sensorLimit: 0, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor1 := sensor{id: "sensor2", sensorLimit: 20, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor2 := sensor{id: "sensor3", sensorLimit: 20, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor3 := sensor{id: "sensor4", sensorLimit: 2, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, component{id: "charger1", demand: 2, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, component{id: "pv1", demand: 5, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, component{id: "pv2", demand: 5, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, component{id: "charger2", demand: 2, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, component{id: "pv3", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, component{id: "charger3", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, component{id: "pv4", demand: 1, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 2, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 2.5, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor1.pvs[1].setPoint-2.5) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 2.5, got %f", sensor1.pvs[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 2, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 3, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 1, got %f", sensor3.pvs[0].setPoint)
	}

	// flipped order of sensors
	root = sensor{sensorLimit: 0, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	root.childSensors = append(root.childSensors, &sensor3)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor1)
	root.reset()

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 2, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 2.5, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor1.pvs[1].setPoint-2.5) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 2.5, got %f", sensor1.pvs[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 2, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 3, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 1, got %f", sensor3.pvs[0].setPoint)
	}

}

func TestEqualGlobalProductionConsumption(t *testing.T) {
	root := sensor{id: "root", sensorLimit: 0, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor1 := sensor{id: "sensor2", sensorLimit: 4, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor2 := sensor{id: "sensor3", sensorLimit: 20, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	sensor3 := sensor{id: "sensor4", sensorLimit: 4, childSensors: make([]*sensor, 0), pvs: make([]component, 0), chargers: make([]component, 0), virtualComponent: virtualComponent{possibleFlexibility: 0}}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.pvs = append(sensor1.pvs, component{id: "pv1", demand: 5, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, component{id: "pv2", demand: 5, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, component{id: "charger1", demand: 1, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, component{id: "pv3", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, component{id: "charger2", demand: 5, setPoint: 0})
	sensor3.chargers = append(sensor3.chargers, component{id: "charger3", demand: 5, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 2, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor1.pvs[1].setPoint-2) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 2, got %f", sensor1.pvs[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 1, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 2, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.chargers[1].setPoint-2) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 2, got %f", sensor3.chargers[1].setPoint)
	}
}
