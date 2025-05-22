package common

import (
	"time"
)

type Message struct {
	Id        string
	Timestamp time.Time
}

type Value struct {
	Message
	Value float64
}

type FlowProposalsMessage struct {
	LeaderId  string
	Proposals []FlowProposal
	Timestamp time.Time
}

type FlowProposal struct {
	SensorId      string
	Flow          float64
	NumberOfNodes int
}

type SensorLimitsMessage struct {
	Limits    []SensorLimit
	Timestamp time.Time
}

type SensorLimit struct {
	SensorId string
	Limit    float64
}

type DdaRegisterNodeMessage struct {
	NodeId    string
	SensorId  string
	NodeType  string
	Timestamp int64
}

type DdaRegisterSensorMessage struct {
	SensorId       string
	ParentSensorId string
	Timestamp      int64
}

type DdaRegisterEnergyCommunityMessage struct {
	EnergyCommunityId string
	Timestamp         int64
}

const CHARGER_NODE_TYPE = "charger"
const PV_NODE_TPYE = "pv"

const REGISTER_ACTION = "com.siemens.openswarm.register"
const DEREGISTER_ACTION = "com.siemens.openswarm.deregister"
const REGISTER_ENERGY_COMMUNITY_ACTION = "com.siemens.openswarm.registerenergycommunity"
const DEREGISTER_ENERGY_COMMUNITY_ACTION = "com.siemens.openswarm.deregisterenergycommunity"
const NEW_ROUND_EVENT = "com.siemens.openswarm.newround"
const FLOW_PROPOSAL_EVENT = "com.siemens.openswarm.flowproposal"
const SENSOR_LIMITS_EVENT = "com.siemens.openswarm.sensorlimits"
const GET_CHARGER_DEMAND_ACTION = "com.siemens.openswarm.chargerdemand"
const GET_PV_DEMAND_ACTION = "com.siemens.openswarm.pvdemand"
const GET_MEASUREMENT_ACTION = "com.siemens.openswarm.measurement"
const SET_POINT = "com.siemens.openswarm.setpoint"

func AppendId(ddaType string, id string) string {
	return ddaType + "_" + id
}
