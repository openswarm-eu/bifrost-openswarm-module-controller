package controller

type state struct {
	leader          bool
	registeredAtDso bool
	toplogy         *toplogy

	clusterMembers int
}
