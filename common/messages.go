package common

import (
	"time"
)

type RegisterMessage struct {
	NodeId    string
	SensorId  string
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

const REGISTER_EVENT = "com.siemens.openswarm.register"
const DEREGISTER_EVENT = "com.siemens.openswarm.deregister"
const REGISTER_RESPONSE_EVENT = "com.siemens.openswarm.registerresponse"
const CHARGER_ACTION = "com.siemens.openswarm.charger"
const PRODUCTION_ACTION = "com.siemens.openswarm.production"
const CHARGING_SET_POINT = "com.siemens.openswarm.chargersetpoint"

type DdaRegisterMessage struct {
	NodeId    string
	SensorId  string
	Timestamp int64
}
