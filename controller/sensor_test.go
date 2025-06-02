package controller

import (
	"math"
	"testing"
)

const tolerance = .00001

func TestConsumptionOutgoingLimit(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 2, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 10, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 7, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 2, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv1", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger4", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv2", demand: 5, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv3", demand: 5, setPoint: 0})

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

	//test if reset is working
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

func TestConsumptionRequestedPowerLimit(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 10, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 0.5, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 5, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv1", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger4", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv2", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv3", demand: 3, setPoint: 0})

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

func TestConsumptionIngoingLimit(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 1, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 10, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 7, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 1, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv2", demand: 2, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 2, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv2", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger4", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv3", demand: 2, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv4", demand: 3, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 2, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-2) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 2, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 1, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor1.pvs[1].setPoint-2) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 2, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 2, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 3, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 2, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-3) > tolerance {
		t.Errorf("Expected pv5 setPoint to be 3, got %f", sensor3.pvs[1].setPoint)
	}
}

func TestProductionOutgoingLimit(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 2, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 5, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 3, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 4, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 1, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv2", demand: 2, setPoint: 0})

	sensor3.pvs = append(sensor3.pvs, &component{id: "pv3", demand: 10, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv4", demand: 7, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 1, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-3) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 1, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-4) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 4, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 1, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 1, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 1, got %f", sensor3.pvs[1].setPoint)
	}
}

func TestProductionRequestedPowerLimit(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 4, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 3, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 4, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 1, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv2", demand: 5, setPoint: 0})

	sensor3.pvs = append(sensor3.pvs, &component{id: "pv4", demand: 10, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv5", demand: 0.5, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-4) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 4, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-3) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 3, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 2.5, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 1, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 2.5, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 2.5, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-0.5) > tolerance {
		t.Errorf("Expected pv5 setPoint to be 0.5, got %f", sensor3.pvs[1].setPoint)
	}
}

func TestProductionIngoingLimit(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 1, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 2, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 3, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 4, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 1, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv2", demand: 2, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger4", demand: 1, setPoint: 0})
	sensor3.chargers = append(sensor3.chargers, &component{id: "charger5", demand: 2, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv3", demand: 10, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv4", demand: 7, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 2, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-3) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 3, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 3, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 1, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 2, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 1, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.chargers[1].setPoint-2) > tolerance {
		t.Errorf("Expected charger5 setPoint to be 2, got %f", sensor3.chargers[1].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 2, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-2) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 2, got %f", sensor3.pvs[1].setPoint)
	}
}

func TestBalanced(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 2, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 3, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 5, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 1, setPoint: 0})
	sensor2.chargers = append(sensor2.chargers, &component{id: "charger4", demand: 2, setPoint: 0})

	sensor3.pvs = append(sensor3.pvs, &component{id: "pv2", demand: 2, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv3", demand: 1, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 2, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-3) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 3, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-5) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 5, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 1, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.chargers[1].setPoint-2) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 2, got %f", sensor2.chargers[1].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 2, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 1, got %f", sensor3.pvs[1].setPoint)
	}
}

func TestConsumptionWithChildren(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 1, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor4 := sensor{id: "sensor4", limit: 1, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor5 := sensor{id: "sensor5", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}

	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor3)
	root.childSensors = append(root.childSensors, &sensor5)
	sensor1.childSensors = append(sensor1.childSensors, &sensor2)
	sensor3.childSensors = append(sensor3.childSensors, &sensor4)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 10, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 7, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 1, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv2", demand: 2, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 5, setPoint: 0})
	sensor2.chargers = append(sensor2.chargers, &component{id: "charger4", demand: 5, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv3", demand: 2, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv4", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger5", demand: 3, setPoint: 0})
	sensor3.chargers = append(sensor3.chargers, &component{id: "charger6", demand: 1, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv5", demand: 1, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv6", demand: 2, setPoint: 0})

	sensor4.chargers = append(sensor4.chargers, &component{id: "charger7", demand: 2, setPoint: 0})
	sensor4.chargers = append(sensor4.chargers, &component{id: "charger8", demand: 1, setPoint: 0})
	sensor4.pvs = append(sensor4.pvs, &component{id: "pv7", demand: 0.5, setPoint: 0})
	sensor4.pvs = append(sensor4.pvs, &component{id: "pv8", demand: 0.5, setPoint: 0})

	sensor5.chargers = append(sensor5.chargers, &component{id: "charger9", demand: 5, setPoint: 0})
	sensor5.pvs = append(sensor5.pvs, &component{id: "pv9", demand: 2, setPoint: 0})
	sensor5.pvs = append(sensor5.pvs, &component{id: "pv10", demand: 4, setPoint: 0})
	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-1.75) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 1.75, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-1.75) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 1.75, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 1, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor1.pvs[1].setPoint-2) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 2, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-1.75) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 1.75, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.chargers[1].setPoint-1.75) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 1.75, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 2, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor2.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 1, got %f", sensor2.pvs[1].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger5 setPoint to be 3, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger6 setPoint to be 1, got %f", sensor3.chargers[1].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv5 setPoint to be 1, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-2) > tolerance {
		t.Errorf("Expected pv6 setPoint to be 2, got %f", sensor3.pvs[1].setPoint)
	}
	if math.Abs(sensor4.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger7 setPoint to be 1, got %f", sensor4.chargers[0].setPoint)
	}
	if math.Abs(sensor4.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger8 setPoint to be 1, got %f", sensor4.chargers[1].setPoint)
	}
	if math.Abs(sensor4.pvs[0].setPoint-0.5) > tolerance {
		t.Errorf("Expected pv7 setPoint to be 0.5, got %f", sensor4.pvs[0].setPoint)
	}
	if math.Abs(sensor4.pvs[1].setPoint-0.5) > tolerance {
		t.Errorf("Expected pv8 setPoint to be 0.5, got %f", sensor4.pvs[1].setPoint)
	}
	if math.Abs(sensor5.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger9 setPoint to be 3, got %f", sensor5.chargers[0].setPoint)
	}
	if math.Abs(sensor5.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv9 setPoint to be 2, got %f", sensor5.pvs[0].setPoint)
	}
	if math.Abs(sensor5.pvs[1].setPoint-4) > tolerance {
		t.Errorf("Expected pv10 setPoint to be 4, got %f", sensor5.pvs[1].setPoint)
	}
}

func TestProductionWithChildren(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 1, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor4 := sensor{id: "sensor4", limit: 1, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor5 := sensor{id: "sensor5", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}

	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor4)
	sensor2.childSensors = append(sensor2.childSensors, &sensor3)
	sensor4.childSensors = append(sensor4.childSensors, &sensor5)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 2, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 4, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 5, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 1, setPoint: 0})
	sensor2.chargers = append(sensor2.chargers, &component{id: "charger4", demand: 2, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv2", demand: 3, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv3", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger5", demand: 0.5, setPoint: 0})
	sensor3.chargers = append(sensor3.chargers, &component{id: "charger6", demand: 0.5, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv4", demand: 2, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv5", demand: 1, setPoint: 0})

	sensor4.chargers = append(sensor4.chargers, &component{id: "charger7", demand: 1, setPoint: 0})
	sensor4.chargers = append(sensor4.chargers, &component{id: "charger8", demand: 2, setPoint: 0})
	sensor4.pvs = append(sensor4.pvs, &component{id: "pv6", demand: 10, setPoint: 0})
	sensor4.pvs = append(sensor4.pvs, &component{id: "pv7", demand: 7, setPoint: 0})

	sensor5.chargers = append(sensor5.chargers, &component{id: "charger9", demand: 2, setPoint: 0})
	sensor5.chargers = append(sensor5.chargers, &component{id: "charger10", demand: 1, setPoint: 0})
	sensor5.pvs = append(sensor5.pvs, &component{id: "pv8", demand: 5, setPoint: 0})
	sensor5.pvs = append(sensor5.pvs, &component{id: "pv9", demand: 5, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 2, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-4) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 4, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 3, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 1, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.chargers[1].setPoint-2) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 2, got %f", sensor2.chargers[1].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 3, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor2.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 1, got %f", sensor2.pvs[1].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-0.5) > tolerance {
		t.Errorf("Expected charger5 setPoint to be 0.5, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.chargers[1].setPoint-0.5) > tolerance {
		t.Errorf("Expected charger6 setPoint to be 0.51, got %f", sensor3.chargers[1].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 1, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv5 setPoint to be 1, got %f", sensor3.pvs[1].setPoint)
	}
	if math.Abs(sensor4.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger7 setPoint to be 1, got %f", sensor4.chargers[0].setPoint)
	}
	if math.Abs(sensor4.chargers[1].setPoint-2) > tolerance {
		t.Errorf("Expected charger8 setPoint to be 2, got %f", sensor4.chargers[1].setPoint)
	}
	if math.Abs(sensor4.pvs[0].setPoint-1.75) > tolerance {
		t.Errorf("Expected pv6 setPoint to be 1.75, got %f", sensor4.pvs[0].setPoint)
	}
	if math.Abs(sensor4.pvs[1].setPoint-1.75) > tolerance {
		t.Errorf("Expected pv7 setPoint to be 1.75, got %f", sensor4.pvs[1].setPoint)
	}
	if math.Abs(sensor5.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger9 setPoint to be 2, got %f", sensor5.chargers[0].setPoint)
	}
	if math.Abs(sensor5.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger10 setPoint to be 1, got %f", sensor5.chargers[1].setPoint)
	}
	if math.Abs(sensor5.pvs[0].setPoint-1.75) > tolerance {
		t.Errorf("Expected pv8 setPoint to be 1.75, got %f", sensor5.pvs[0].setPoint)
	}
	if math.Abs(sensor5.pvs[1].setPoint-1.75) > tolerance {
		t.Errorf("Expected pv9 setPoint to be 1.75, got %f", sensor5.pvs[1].setPoint)
	}

}

func TestConsumptionWithProductionChildren(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor4 := sensor{id: "sensor4", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor3)
	sensor1.childSensors = append(sensor1.childSensors, &sensor2)
	sensor3.childSensors = append(sensor3.childSensors, &sensor4)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 5, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 4, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 2, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv2", demand: 1, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 3, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv3", demand: 2, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger4", demand: 1, setPoint: 0})
	sensor3.chargers = append(sensor3.chargers, &component{id: "charger5", demand: 1, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv4", demand: 3, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv5", demand: 1, setPoint: 0})

	sensor4.chargers = append(sensor4.chargers, &component{id: "charger6", demand: 1, setPoint: 0})
	sensor4.pvs = append(sensor4.pvs, &component{id: "pv6", demand: 2, setPoint: 0})
	sensor4.pvs = append(sensor4.pvs, &component{id: "pv7", demand: 1, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 3, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-3) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 3, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 2, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor1.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 1, got %f", sensor1.pvs[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 3, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 2, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 1, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger5 setPoint to be 1, got %f", sensor3.chargers[1].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 3, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv5 setPoint to be 1, got %f", sensor3.pvs[1].setPoint)
	}
	if math.Abs(sensor4.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger6 setPoint to be 1, got %f", sensor4.chargers[0].setPoint)
	}
	if math.Abs(sensor4.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv6 setPoint to be 1, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor4.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv7 setPoint to be 1, got %f", sensor3.pvs[1].setPoint)
	}
}

func TestProductionWithConsumtionChildren(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor4 := sensor{id: "sensor4", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor3)
	sensor1.childSensors = append(sensor1.childSensors, &sensor2)
	sensor3.childSensors = append(sensor3.childSensors, &sensor4)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 3, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 1, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 1, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv2", demand: 1, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 2, setPoint: 0})
	sensor2.chargers = append(sensor2.chargers, &component{id: "charger4", demand: 1, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv3", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger5", demand: 2, setPoint: 0})
	sensor3.chargers = append(sensor3.chargers, &component{id: "charger6", demand: 1, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv4", demand: 3, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv5", demand: 3, setPoint: 0})

	sensor4.chargers = append(sensor4.chargers, &component{id: "charger7", demand: 2, setPoint: 0})
	sensor4.pvs = append(sensor4.pvs, &component{id: "pv6", demand: 3, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 3, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 1, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 1, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor1.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 1, got %f", sensor1.pvs[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 2, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 1, got %f", sensor2.chargers[1].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 1, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger5 setPoint to be 2, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger6 setPoint to be 1, got %f", sensor3.chargers[1].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 3, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-3) > tolerance {
		t.Errorf("Expected pv5 setPoint to be 3, got %f", sensor3.pvs[1].setPoint)
	}
	if math.Abs(sensor4.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger7 setPoint to be 2, got %f", sensor4.chargers[0].setPoint)
	}
	if math.Abs(sensor4.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv6 setPoint to be 3, got %f", sensor3.pvs[0].setPoint)
	}
}

func TestConsumptionChildrenWithDifferntFlow(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 1, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor4 := sensor{id: "sensor4", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor5 := sensor{id: "sensor5", limit: 1, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor3)
	sensor1.childSensors = append(sensor1.childSensors, &sensor2)
	sensor3.childSensors = append(sensor3.childSensors, &sensor4)
	sensor3.childSensors = append(sensor3.childSensors, &sensor5)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 15, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 0.5, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 1, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 2, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv2", demand: 2, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv3", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger4", demand: 3, setPoint: 0})
	sensor3.chargers = append(sensor3.chargers, &component{id: "charger5", demand: 2, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv4", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv5", demand: 6, setPoint: 0})

	sensor4.chargers = append(sensor4.chargers, &component{id: "charger6", demand: 3, setPoint: 0})
	sensor4.chargers = append(sensor4.chargers, &component{id: "charger7", demand: 1, setPoint: 0})
	sensor4.pvs = append(sensor4.pvs, &component{id: "pv6", demand: 2, setPoint: 0})

	sensor5.chargers = append(sensor5.chargers, &component{id: "charger8", demand: 4, setPoint: 0})
	sensor5.chargers = append(sensor5.chargers, &component{id: "charger9", demand: 2, setPoint: 0})
	sensor5.pvs = append(sensor5.pvs, &component{id: "pv7", demand: 2, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 2.5, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.chargers[1].setPoint-0.5) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 0.5, got %f", sensor1.chargers[1].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-1) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 1, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 2, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 2, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor2.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 1, got %f", sensor2.pvs[1].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 3, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.chargers[1].setPoint-2) > tolerance {
		t.Errorf("Expected charger5 setPoint to be 2, got %f", sensor3.chargers[1].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-4) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 4, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-5) > tolerance {
		t.Errorf("Expected pv5 setPoint to be 5, got %f", sensor3.pvs[1].setPoint)
	}
	if math.Abs(sensor4.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger6 setPoint to be 3, got %f", sensor4.chargers[0].setPoint)
	}
	if math.Abs(sensor4.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger7 setPoint to be 1, got %f", sensor4.chargers[1].setPoint)
	}
	if math.Abs(sensor4.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv6 setPoint to be 2, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor5.chargers[0].setPoint-1.5) > tolerance {
		t.Errorf("Expected charger8 setPoint to be 1.5, got %f", sensor5.chargers[0].setPoint)
	}
	if math.Abs(sensor5.chargers[1].setPoint-1.5) > tolerance {
		t.Errorf("Expected charger9 setPoint to be 1.5, got %f", sensor5.chargers[1].setPoint)
	}
	if math.Abs(sensor5.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv7 setPoint to be 2, got %f", sensor5.pvs[0].setPoint)
	}
}

func TestProductionChildrenWithDifferntFlow(t *testing.T) { // TODO
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 1, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor4 := sensor{id: "sensor4", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor5 := sensor{id: "sensor5", limit: 1, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor3)
	sensor1.childSensors = append(sensor1.childSensors, &sensor2)
	sensor3.childSensors = append(sensor3.childSensors, &sensor4)
	sensor3.childSensors = append(sensor3.childSensors, &sensor5)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 1, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv1", demand: 15, setPoint: 0})
	sensor1.pvs = append(sensor1.pvs, &component{id: "pv2", demand: 0.5, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger2", demand: 2, setPoint: 0})
	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 1, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv3", demand: 2, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger4", demand: 4, setPoint: 0})
	sensor3.chargers = append(sensor3.chargers, &component{id: "charger5", demand: 6, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv4", demand: 3, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv5", demand: 2, setPoint: 0})

	sensor4.chargers = append(sensor4.chargers, &component{id: "charger6", demand: 2, setPoint: 0})
	sensor4.pvs = append(sensor4.pvs, &component{id: "pv6", demand: 3, setPoint: 0})
	sensor4.pvs = append(sensor4.pvs, &component{id: "pv7", demand: 1, setPoint: 0})

	sensor5.chargers = append(sensor5.chargers, &component{id: "charger7", demand: 2, setPoint: 0})
	sensor5.pvs = append(sensor5.pvs, &component{id: "pv8", demand: 4, setPoint: 0})
	sensor5.pvs = append(sensor5.pvs, &component{id: "pv9", demand: 2, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-1) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 1, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor1.pvs[0].setPoint-2.5) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 2.5, got %f", sensor1.pvs[0].setPoint)
	}
	if math.Abs(sensor1.pvs[1].setPoint-0.5) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 0.5, got %f", sensor1.pvs[1].setPoint)
	}
	if math.Abs(sensor2.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 2, got %f", sensor2.chargers[0].setPoint)
	}
	if math.Abs(sensor2.chargers[1].setPoint-1) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 1, got %f", sensor2.chargers[1].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-2) > tolerance {
		t.Errorf("Expected pv3 setPoint to be 2, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-4) > tolerance {
		t.Errorf("Expected charger4 setPoint to be 4, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor3.chargers[1].setPoint-5) > tolerance {
		t.Errorf("Expected charger5 setPoint to be 5, got %f", sensor3.chargers[1].setPoint)
	}
	if math.Abs(sensor3.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv4 setPoint to be 3, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor3.pvs[1].setPoint-2) > tolerance {
		t.Errorf("Expected pv5 setPoint to be 2, got %f", sensor3.pvs[1].setPoint)
	}
	if math.Abs(sensor4.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger6 setPoint to be 2, got %f", sensor4.chargers[0].setPoint)
	}
	if math.Abs(sensor4.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv6 setPoint to be 3, got %f", sensor3.pvs[0].setPoint)
	}
	if math.Abs(sensor4.pvs[1].setPoint-1) > tolerance {
		t.Errorf("Expected pv7 setPoint to be 1, got %f", sensor3.pvs[1].setPoint)
	}
	if math.Abs(sensor5.chargers[0].setPoint-2) > tolerance {
		t.Errorf("Expected charger7 setPoint to be 2, got %f", sensor5.chargers[0].setPoint)
	}
	if math.Abs(sensor5.pvs[0].setPoint-1.5) > tolerance {
		t.Errorf("Expected pv8 setPoint to be 1.5, got %f", sensor5.pvs[0].setPoint)
	}
	if math.Abs(sensor5.pvs[1].setPoint-1.5) > tolerance {
		t.Errorf("Expected pv9 setPoint to be 1.5, got %f", sensor5.pvs[1].setPoint)
	}
}

func TestReset(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 20, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 2, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor2)
	root.childSensors = append(root.childSensors, &sensor3)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 10, setPoint: 0})
	sensor1.chargers = append(sensor1.chargers, &component{id: "charger2", demand: 7, setPoint: 0})

	sensor2.chargers = append(sensor2.chargers, &component{id: "charger3", demand: 2, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv1", demand: 1, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger4", demand: 4, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv2", demand: 5, setPoint: 0})
	sensor3.pvs = append(sensor3.pvs, &component{id: "pv3", demand: 5, setPoint: 0})

	root.setSetPoints()
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

func TestZeroLimit(t *testing.T) {
	root := sensor{id: "root", limit: math.MaxFloat64, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor1 := sensor{id: "sensor1", limit: 0, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor2 := sensor{id: "sensor2", limit: 10, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor3 := sensor{id: "sensor3", limit: 0, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor4 := sensor{id: "sensor4", limit: 0, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	sensor5 := sensor{id: "sensor5", limit: 10, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	root.childSensors = append(root.childSensors, &sensor1)
	root.childSensors = append(root.childSensors, &sensor3)
	root.childSensors = append(root.childSensors, &sensor4)
	sensor1.childSensors = append(sensor1.childSensors, &sensor2)
	sensor4.childSensors = append(sensor4.childSensors, &sensor5)

	sensor1.chargers = append(sensor1.chargers, &component{id: "charger1", demand: 5, setPoint: 0})
	sensor2.pvs = append(sensor2.pvs, &component{id: "pv1", demand: 5, setPoint: 0})

	sensor3.chargers = append(sensor3.chargers, &component{id: "charger2", demand: 4, setPoint: 0})

	sensor4.pvs = append(sensor4.pvs, &component{id: "pv2", demand: 3, setPoint: 0})
	sensor5.chargers = append(sensor5.chargers, &component{id: "charger3", demand: 3, setPoint: 0})

	root.setSetPoints()

	if math.Abs(sensor1.chargers[0].setPoint-5) > tolerance {
		t.Errorf("Expected charger1 setPoint to be 5, got %f", sensor1.chargers[0].setPoint)
	}
	if math.Abs(sensor2.pvs[0].setPoint-5) > tolerance {
		t.Errorf("Expected pv1 setPoint to be 5, got %f", sensor2.pvs[0].setPoint)
	}
	if math.Abs(sensor3.chargers[0].setPoint-0) > tolerance {
		t.Errorf("Expected charger2 setPoint to be 0, got %f", sensor3.chargers[0].setPoint)
	}
	if math.Abs(sensor4.pvs[0].setPoint-3) > tolerance {
		t.Errorf("Expected pv2 setPoint to be 3, got %f", sensor4.pvs[0].setPoint)
	}
	if math.Abs(sensor5.chargers[0].setPoint-3) > tolerance {
		t.Errorf("Expected charger3 setPoint to be 3, got %f", sensor5.chargers[0].setPoint)
	}
}
