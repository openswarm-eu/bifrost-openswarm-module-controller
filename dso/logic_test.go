package dso

import (
	"testing"

	"code.siemens.com/energy-community-controller/common"
)

func TestCalculateSensorLimit(t *testing.T) {
	state := &state{
		energyCommunities: make([]*energyCommunity, 0),
		topology: topology{
			Version: 1,
			Sensors: make(map[string]sensor),
		},
		localSenorInformations: make(map[string]*localSenorInformation),
	}
	state.energyCommunities = append(state.energyCommunities, &energyCommunity{Id: "ec1"})
	state.energyCommunities = append(state.energyCommunities, &energyCommunity{Id: "ec2"})
	state.energyCommunities = append(state.energyCommunities, &energyCommunity{Id: "ec3"})
	state.energyCommunities = append(state.energyCommunities, &energyCommunity{Id: "ec4"})
	state.topology.Sensors["sensor1"] = sensor{limit: 10}
	state.topology.Sensors["sensor2"] = sensor{limit: 5}
	state.topology.Sensors["sensor3"] = sensor{limit: 5}
	state.localSenorInformations["sensor1"] = &localSenorInformation{
		measurement: 9,
		sumECLimits: 5,
		ecFlowProposal: map[string]common.FlowProposal{
			"ec1": {Flow: 10, NumberOfNodes: 2},
			"ec2": {Flow: 4, NumberOfNodes: 1},
			"ec3": {Flow: -11, NumberOfNodes: 4},
		},
	}
	state.localSenorInformations["sensor2"] = &localSenorInformation{
		measurement: -5,
		sumECLimits: 3,
		ecFlowProposal: map[string]common.FlowProposal{
			"ec1": {Flow: -10, NumberOfNodes: 2},
			"ec2": {Flow: 3, NumberOfNodes: 1},
			"ec3": {Flow: -11, NumberOfNodes: 1},
		},
	}
	state.localSenorInformations["sensor3"] = &localSenorInformation{
		measurement: 5,
		sumECLimits: 4,
		ecFlowProposal: map[string]common.FlowProposal{
			"ec1": {Flow: -2, NumberOfNodes: 2},
			"ec2": {Flow: 6, NumberOfNodes: 1},
			"ec3": {Flow: 1, NumberOfNodes: 1},
		},
	}

	logic, _ := newLogic(Config{}, &connector{}, state)
	logic.calculateSensorLimits()

	if state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor1"] != 10 {
		t.Errorf("Expected sensor1 limit for ec1 to be 10, got %f", state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor1"])
	}
	if state.energyCommunitySensorLimits["ec2"].SensorLimits["sensor1"] != 4 {
		t.Errorf("Expected sensor1 limit for ec2 to be 4, got %f", state.energyCommunitySensorLimits["ec2"].SensorLimits["sensor1"])
	}
	if state.energyCommunitySensorLimits["ec3"].SensorLimits["sensor1"] != 11 {
		t.Errorf("Expected sensor1 limit for ec3 to be 11, got %f", state.energyCommunitySensorLimits["ec3"].SensorLimits["sensor1"])
	}
	if state.localSenorInformations["sensor1"].sumECLimits != 3 {
		t.Errorf("Expected sumECLimits for sensor1 to be 3, got %f", state.localSenorInformations["sensor1"].sumECLimits)
	}

	if state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor2"] != 4 {
		t.Errorf("Expected sensor2 limit for ec1 to be 4, got %f", state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor2"])
	}
	if state.energyCommunitySensorLimits["ec2"].SensorLimits["sensor2"] != 3 {
		t.Errorf("Expected sensor2 limit for ec2 to be 3, got %f", state.energyCommunitySensorLimits["ec2"].SensorLimits["sensor2"])
	}
	if state.energyCommunitySensorLimits["ec3"].SensorLimits["sensor2"] != 2 {
		t.Errorf("Expected sensor2 limit for ec3 to be 2, got %f", state.energyCommunitySensorLimits["ec3"].SensorLimits["sensor2"])
	}
	if state.localSenorInformations["sensor2"].sumECLimits != -3 {
		t.Errorf("Expected sumECLimits for sensor2 to be -3, got %f", state.localSenorInformations["sensor2"].sumECLimits)
	}

	if state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor3"] != 2 {
		t.Errorf("Expected sensor3 limit for ec1 to be 2, got %f", state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor3"])
	}
	if state.energyCommunitySensorLimits["ec2"].SensorLimits["sensor3"] != 5 {
		t.Errorf("Expected sensor3 limit for ec2 to be 5, got %f", state.energyCommunitySensorLimits["ec2"].SensorLimits["sensor3"])
	}
	if state.energyCommunitySensorLimits["ec3"].SensorLimits["sensor3"] != 1 {
		t.Errorf("Expected sensor3 limit for ec3 to be 1, got %f", state.energyCommunitySensorLimits["ec3"].SensorLimits["sensor3"])
	}
	if len(state.energyCommunitySensorLimits) != 3 {
		t.Errorf("Expected 3 energy community sensor limits, got %d", len(state.energyCommunitySensorLimits))
	}
	if state.localSenorInformations["sensor3"].sumECLimits != 4 {
		t.Errorf("Expected sumECLimits for sensor3 to be 4, got %f", state.localSenorInformations["sensor3"].sumECLimits)
	}

}
