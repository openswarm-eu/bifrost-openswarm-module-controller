package common

import (
	"time"
)

const CHARGER_NODE_TYPE = "charger"
const PV_NODE_TPYE = "pv"

type RegisterNodeMessage struct {
	NodeId    string
	SensorId  string
	NodeType  string
	Timestamp time.Time
}

type RegisterSensorMessage struct {
	SensorId       string
	ParentSensorId string
	Limit          float64
	Timestamp      time.Time
}

type RegisterEnergyCommunityMessage struct {
	EnergyCommunityId string
	Timestamp         time.Time
}

type Message struct {
	Id        string
	Timestamp time.Time
}

type Value struct {
	Message
	Value float64
}

type FlowProposalsMessage struct {
	EnergyCommunityId string
	Proposals         map[string]FlowProposal // sensorId -> FlowProposal
	Timestamp         time.Time
}

type FlowProposal struct {
	Flow          float64
	NumberOfNodes int
}

type EnergyCommunitySensorLimitMessage struct {
	SensorLimits map[string]float64 // sensorId -> limit
}

type TopologyMessage struct {
	Topology  map[string][]string // parentSensorId -> []childSensorId
	Timestamp int64
}

const REGISTER_ACTION = "com.siemens.openswarm.register"
const DEREGISTER_ACTION = "com.siemens.openswarm.deregister"
const REGISTER_ENERGY_COMMUNITY_ACTION = "com.siemens.openswarm.registerenergycommunity"
const DEREGISTER_ENERGY_COMMUNITY_ACTION = "com.siemens.openswarm.deregisterenergycommunity"
const TOPOLOGY_UPDATE_ACTION = "com.siemens.openswarm.topologyupdate"

const GET_CHARGER_DEMAND_ACTION = "com.siemens.openswarm.chargerdemand"
const GET_PV_DEMAND_ACTION = "com.siemens.openswarm.pvdemand"
const GET_SENSOR_MEASUREMENT_ACTION = "com.siemens.openswarm.measurement"
const SET_POINT = "com.siemens.openswarm.setpoint"

const GET_FLOW_PROPOSAL_ACTION = "com.siemens.openswarm.floproposal"
const SET_SENSOR_LIMITS_EVENT = "com.siemens.openswarm.setsensorlimits"

func AppendId(ddaType string, id string) string {
	return ddaType + "_" + id
}
