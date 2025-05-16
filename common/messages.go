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

const CHARGER_NODE_TYPE = "charger"
const PV_NODE_TPYE = "pv"

const REGISTER_EVENT = "com.siemens.openswarm.register"
const DEREGISTER_EVENT = "com.siemens.openswarm.deregister"
const REGISTER_RESPONSE_EVENT = "com.siemens.openswarm.registerresponse"
const GET_CHARGER_DEMAND_ACTION = "com.siemens.openswarm.chargerDemand"
const GET_PV_DEMAND_ACTION = "com.siemens.openswarm.pvDemand"
const SET_POINT = "com.siemens.openswarm.setpoint"

type DdaRegisterMessage struct {
	NodeId    string
	SensorId  string
	NodeType  string
	Timestamp int64
}
