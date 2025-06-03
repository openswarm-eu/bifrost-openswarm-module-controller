package dso

import (
	"testing"

	"code.siemens.com/energy-community-controller/common"
)

func TestCalculateSensorLimit(t *testing.T) {
	state := &state{
		energyCommunities: make(map[string]int),
		topology: topology{
			Version: 1,
			Sensors: make(map[string]*sensor),
		},
		localSenorInformations: make(map[string]*localSenorInformation),
	}
	state.energyCommunities["ec1"] = 0
	state.energyCommunities["ec2"] = 0
	state.energyCommunities["ec3"] = 0
	state.energyCommunities["ec4"] = 0
	state.topology.Sensors["sensor1"] = &sensor{Limit: 10}
	state.topology.Sensors["sensor2"] = &sensor{Limit: 5}
	state.topology.Sensors["sensor3"] = &sensor{Limit: 5}
	state.localSenorInformations["sensor1"] = &localSenorInformation{
		measurement: 9,
		sumECLimits: 5,
		ecFlowProposal: map[string]common.FlowProposal{
			"ec1": {Flow: 10, NumberOfNodes: 2},
			"ec2": {Flow: 4, NumberOfNodes: 1},
			"ec3": {Flow: -11, NumberOfNodes: 4},
			"ec4": {Flow: 0, NumberOfNodes: 1},
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

	calculator := newSensorLimitsCalculator(state)
	calculator.calculateSensorLimits()

	if state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor1"] != 10 {
		t.Errorf("Expected sensor1 limit for ec1 to be 10, got %f", state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor1"])
	}
	if state.energyCommunitySensorLimits["ec2"].SensorLimits["sensor1"] != 4 {
		t.Errorf("Expected sensor1 limit for ec2 to be 4, got %f", state.energyCommunitySensorLimits["ec2"].SensorLimits["sensor1"])
	}
	if state.energyCommunitySensorLimits["ec3"].SensorLimits["sensor1"] != 11 {
		t.Errorf("Expected sensor1 limit for ec3 to be 11, got %f", state.energyCommunitySensorLimits["ec3"].SensorLimits["sensor1"])
	}
	if state.energyCommunitySensorLimits["ec4"].SensorLimits["sensor1"] != 0 {
		t.Errorf("Expected sensor1 limit for ec4 to be 0, got %f", state.energyCommunitySensorLimits["ec4"].SensorLimits["sensor1"])
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
	if state.localSenorInformations["sensor3"].sumECLimits != 4 {
		t.Errorf("Expected sumECLimits for sensor3 to be 4, got %f", state.localSenorInformations["sensor3"].sumECLimits)
	}

	if len(state.energyCommunitySensorLimits) != 4 {
		t.Errorf("Expected 4 energy community sensor limits, got %d", len(state.energyCommunitySensorLimits))
	}
}

func TestCalculateSensorLimitZero(t *testing.T) {
	state := &state{
		energyCommunities: make(map[string]int),
		topology: topology{
			Version: 1,
			Sensors: make(map[string]*sensor),
		},
		localSenorInformations: make(map[string]*localSenorInformation),
	}
	state.energyCommunities["ec1"] = 0
	state.topology.Sensors["sensor1"] = &sensor{Limit: 10}
	state.localSenorInformations["sensor1"] = &localSenorInformation{
		measurement: 9,
		sumECLimits: 5,
		ecFlowProposal: map[string]common.FlowProposal{
			"ec1": {Flow: 0, NumberOfNodes: 0},
		},
	}

	calculator := newSensorLimitsCalculator(state)
	calculator.calculateSensorLimits()

	if state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor1"] != 0 {
		t.Errorf("Expected sensor1 limit for ec1 to be 0, got %f", state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor1"])
	}

	if len(state.energyCommunitySensorLimits) != 1 {
		t.Errorf("Expected 1 energy community sensor limits, got %d", len(state.energyCommunitySensorLimits))
	}
}

func TestCalculatOverload(t *testing.T) {
	state := &state{
		energyCommunities: make(map[string]int),
		topology: topology{
			Version: 1,
			Sensors: make(map[string]*sensor),
		},
		localSenorInformations: make(map[string]*localSenorInformation),
	}
	state.energyCommunities["ec1"] = 0
	state.topology.Sensors["sensor1"] = &sensor{Limit: 10}
	state.localSenorInformations["sensor1"] = &localSenorInformation{
		measurement: 15,
		sumECLimits: 3,
		ecFlowProposal: map[string]common.FlowProposal{
			"ec1": {Flow: 0, NumberOfNodes: 0},
		},
	}

	calculator := newSensorLimitsCalculator(state)
	calculator.calculateSensorLimits()

	if state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor1"] != 0 {
		t.Errorf("Expected sensor1 limit for ec1 to be 0, got %f", state.energyCommunitySensorLimits["ec1"].SensorLimits["sensor1"])
	}

	if len(state.energyCommunitySensorLimits) != 1 {
		t.Errorf("Expected 1 energy community sensor limits, got %d", len(state.energyCommunitySensorLimits))
	}
}
