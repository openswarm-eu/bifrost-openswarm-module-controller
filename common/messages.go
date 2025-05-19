package common

import (
	"time"
)

type RegisterMessage struct {
	NodeId    string
	SensorId  string
	NodeType  string
	Timestamp time.Time
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
	LeaderId  string
	Proposals []FlowProposal
	Timestamp time.Time
}

type FlowProposal struct {
	SensorId string
	Flow     float64
}

type SensorLimitsMessage struct {
	Limits    []SensorLimit
	Timestamp time.Time
}

type SensorLimit struct {
	SensorId string
	Limit    float64
}

type DdaRegisterMessage struct {
	NodeId    string
	SensorId  string
	NodeType  string
	Timestamp int64
}

const CHARGER_NODE_TYPE = "charger"
const PV_NODE_TPYE = "pv"

const REGISTER_EVENT = "com.siemens.openswarm.register"
const DEREGISTER_EVENT = "com.siemens.openswarm.deregister"
const REGISTER_RESPONSE_EVENT = "com.siemens.openswarm.registerresponse"
const FLOW_PROPOSAL_EVENT = "com.siemens.openswarm.flowProposal"
const FLOW_PROPOSAL_RESPONSE_EVENT = "com.siemens.openswarm.flowProposalResponse"
const SENSOR_LIMITS_EVENT = "com.siemens.openswarm.sensorLimits"
const GET_CHARGER_DEMAND_ACTION = "com.siemens.openswarm.chargerDemand"
const GET_PV_DEMAND_ACTION = "com.siemens.openswarm.pvDemand"
const SET_POINT = "com.siemens.openswarm.setpoint"

func AppendId(ddaType string, id string) string {
	return ddaType + "_" + id
}
