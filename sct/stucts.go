package sct

type state struct {
	transitions map[event]state
}

type event struct {
	name         string
	controllable bool
}
