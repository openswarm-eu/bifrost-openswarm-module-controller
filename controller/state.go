package controller

import "code.siemens.com/energy-community-controller/common"

type state struct {
	pvProductionValues []common.Value
	chargerIds         []common.Message
	setPoints          []common.Value
	topology           map[string][]string
}
