package dso

import (
	"code.siemens.com/energy-community-controller/common"
)

type state struct {
	// exchanged via raft
	energyCommunities map[string]int
	topology          topology

	// local only
	leader                      bool
	newTopology                 topology
	localSenorInformations      map[string]*localSenorInformation                   // sensorId -> localSenorInformation
	energyCommunitySensorLimits map[string]common.EnergyCommunitySensorLimitMessage // energyCommunityId -> FlowSetPointsMessage
}

type topology struct {
	Version int
	Sensors map[string]*sensor
}

type sensor struct {
	Limit            float64
	ChildrenSensorId []string
	parentSensorId   string // parentSensorId is used to remove the sensor from the topology
}

type localSenorInformation struct {
	measurement    float64
	sumECLimits    float64
	ecFlowProposal map[string]common.FlowProposal // energyCommunityId --> FlowProposal
}

func (s *state) updateLocalSensorInformation() {
	for sensorId := range s.topology.Sensors {
		if _, ok := s.localSenorInformations[sensorId]; !ok {
			s.localSenorInformations[sensorId] = &localSenorInformation{
				measurement:    0,
				ecFlowProposal: make(map[string]common.FlowProposal),
			}
		}
	}
}

func (s *state) resetEnergyCommunitySensorLimits() {
	for sensorId := range s.topology.Sensors {
		if sensor, ok := s.localSenorInformations[sensorId]; !ok {
			sensor.sumECLimits = 0
		}
	}
}

func (s *state) addNodeToTopology(sensorId string, parentSensorId string, limit float64) {
	if _, ok := s.newTopology.Sensors[sensorId]; !ok {
		s.newTopology.Sensors[sensorId] = &sensor{ChildrenSensorId: make([]string, 0), parentSensorId: parentSensorId}
	}

	snsr := s.newTopology.Sensors[sensorId]
	snsr.Limit = limit

	if parentSensorId == "" {
		return
	}

	if _, ok := s.newTopology.Sensors[parentSensorId]; !ok {
		s.newTopology.Sensors[parentSensorId] = &sensor{ChildrenSensorId: make([]string, 0)}
	}

	parentSensor := s.newTopology.Sensors[parentSensorId]
	parentSensor.ChildrenSensorId = append(parentSensor.ChildrenSensorId, sensorId)
	s.newTopology.Sensors[parentSensorId] = parentSensor
}

func (s *state) removeNodeFromTopology(sensorId string) {
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
	childrenSensorId := parentSensor.ChildrenSensorId
	for i, childId := range childrenSensorId {
		if childId == sensorId {
			parentSensor.ChildrenSensorId = append(childrenSensorId[:i], childrenSensorId[i+1:]...)
			break
		}
	}
}
