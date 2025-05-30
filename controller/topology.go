package controller

import (
	"sync"
)

type toplogy struct {
	sensors    map[string]*sensor
	chargers   map[string]*component
	pvs        map[string]*component
	rootSensor *sensor

	mutex sync.Mutex
}

type component struct {
	id string

	demand   float64
	setPoint float64
}

func newToplogy() *toplogy {
	return &toplogy{
		sensors:    make(map[string]*sensor),
		chargers:   make(map[string]*component),
		pvs:        make(map[string]*component),
		rootSensor: &sensor{id: "root", childSensors: make([]*sensor, 0)},
	}
}

func (t *toplogy) addPV(pvId string, sensorId string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, ok := t.sensors[sensorId]; !ok {
		t.sensors[sensorId] = &sensor{id: sensorId, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
	}

	pv := &component{id: pvId, demand: 0, setPoint: 0}
	t.sensors[sensorId].pvs = append(t.sensors[sensorId].pvs, pv)
	t.pvs[pvId] = pv
}

func (t *toplogy) addCharger(chargerId string, sensorId string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if _, ok := t.sensors[sensorId]; !ok {
		t.sensors[sensorId] = &sensor{id: sensorId, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
		t.rootSensor.childSensors = append(t.rootSensor.childSensors, t.sensors[sensorId])
	}

	charger := &component{id: chargerId, demand: 0, setPoint: 0}
	t.sensors[sensorId].chargers = append(t.sensors[sensorId].chargers, charger)
	t.chargers[chargerId] = charger
}

func (t *toplogy) removeNode(nodeId string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, sensor := range t.sensors {
		found := false
		for i, pv := range sensor.pvs {
			if pv.id == nodeId {
				sensor.pvs = append(sensor.pvs[:i], sensor.pvs[i+1:]...)
				found = true
				break
			}
		}

		if !found {
			for i, charger := range sensor.chargers {
				if charger.id == nodeId {
					sensor.chargers = append(sensor.chargers[:i], sensor.chargers[i+1:]...)
					break
				}
			}
		}
	}
}

func (t *toplogy) buildTopology(topology map[string][]string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	for _, s := range t.sensors {
		s.childSensors = make([]*sensor, 0)
	}

	sensorsWithoutParent := make(map[string]*sensor)
	for sensorId := range topology {
		if _, ok := t.sensors[sensorId]; !ok {
			t.sensors[sensorId] = &sensor{id: sensorId, childSensors: make([]*sensor, 0), pvs: make([]*component, 0), chargers: make([]*component, 0)}
		}
		sensorsWithoutParent[sensorId] = t.sensors[sensorId]
	}

	for parentId, childIds := range topology {
		for _, childId := range childIds {
			t.sensors[parentId].childSensors = append(t.sensors[parentId].childSensors, t.sensors[childId])
			delete(sensorsWithoutParent, childId)
		}
	}

	t.rootSensor.childSensors = make([]*sensor, 0)
	for _, sensor := range sensorsWithoutParent {
		t.rootSensor.childSensors = append(t.rootSensor.childSensors, sensor)
	}
}

func (t *toplogy) setSensorLimit(sensorId string, limit float64) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	if sensor, ok := t.sensors[sensorId]; ok {
		sensor.limit = limit
	}
}

func (t *toplogy) setAllSensorLimits(limit float64) {
	for _, sensor := range t.sensors {
		sensor.limit = limit
	}
}
