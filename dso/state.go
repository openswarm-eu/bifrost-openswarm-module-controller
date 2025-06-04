package dso

import (
	"sync"

	"code.siemens.com/energy-community-controller/common"
)

type state struct {
	leader                      bool
	energyCommunities           map[string]int
	newEnergyCommunities        map[string]int
	topology                    topology
	newTopology                 topology
	energyCommunityUpdate       bool
	energyCommunitySensorLimits map[string]common.EnergyCommunitySensorLimitMessage // energyCommunityId -> FlowSetPointsMessage

	mutex sync.Mutex
}

type topology struct {
	Version int
	Sensors map[string]*sensor
}

type sensor struct {
	limit            float64
	childrenSensorId []string
	measurement      float64
	sumECLimits      float64
	ecFlowProposal   map[string]common.FlowProposal
	parentSensorId   string // parentSensorId is used to remove the sensor from the topology
}

func (s *state) resetSensorInformation() {
	for _, sensor := range s.topology.Sensors {
		sensor.measurement = 0
		sensor.ecFlowProposal = make(map[string]common.FlowProposal)
	}
}

func (s *state) resetEnergyCommunitySensorLimits() {
	for _, sensor := range s.topology.Sensors {
		sensor.sumECLimits = 0
	}
}

func (s *state) copyNewTopology() {
	s.topology.Version = s.newTopology.Version

	s.topology.Sensors = make(map[string]*sensor)
	for sensorId, snsr := range s.newTopology.Sensors {
		s.topology.Sensors[sensorId] = &sensor{
			limit:            snsr.limit,
			childrenSensorId: make([]string, len(snsr.childrenSensorId)),
			parentSensorId:   snsr.parentSensorId,
			measurement:      0,
			ecFlowProposal:   make(map[string]common.FlowProposal),
			sumECLimits:      0,
		}
		copy(s.topology.Sensors[sensorId].childrenSensorId, snsr.childrenSensorId)
	}
}

func (s *state) copyNewEnergyCommunities() {
	s.energyCommunities = make(map[string]int)
	for energyCommunityId, version := range s.newEnergyCommunities {
		s.energyCommunities[energyCommunityId] = version
	}
}

func (s *state) addNodeToTopology(sensorId string, parentSensorId string, limit float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.newTopology.Sensors[sensorId]; ok {
		return
	}

	s.newTopology.Sensors[sensorId] = &sensor{childrenSensorId: make([]string, 0), parentSensorId: parentSensorId, ecFlowProposal: make(map[string]common.FlowProposal), limit: limit}

	if parentSensorId == "" {
		return
	}

	if _, ok := s.newTopology.Sensors[parentSensorId]; !ok {
		s.newTopology.Sensors[parentSensorId] = &sensor{childrenSensorId: make([]string, 0)}
	}

	parentSensor := s.newTopology.Sensors[parentSensorId] //todo: can this fail??
	parentSensor.childrenSensorId = append(parentSensor.childrenSensorId, sensorId)
	s.newTopology.Sensors[parentSensorId] = parentSensor
}

func (s *state) removeNodeFromTopology(sensorId string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if _, ok := s.newTopology.Sensors[sensorId]; !ok {
		return
	}

	parentSensorId := s.newTopology.Sensors[sensorId].parentSensorId
	delete(s.newTopology.Sensors, sensorId)

	if parentSensorId == "" {
		return
	}

	if _, ok := s.newTopology.Sensors[parentSensorId]; !ok {
		return
	}

	parentSensor := s.newTopology.Sensors[parentSensorId]
	childrenSensorId := parentSensor.childrenSensorId
	for i, childId := range childrenSensorId {
		if childId == sensorId {
			parentSensor.childrenSensorId = append(childrenSensorId[:i], childrenSensorId[i+1:]...)
			break
		}
	}
}

func (s *state) addEnergyCommunity(energyCommunityId string, version int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.newEnergyCommunities[energyCommunityId] = version
}

func (s *state) removeEnergyCommunity(energyCommunityId string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.newEnergyCommunities, energyCommunityId)
}
