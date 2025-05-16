package controller

type state struct {
	sensors    map[string]*sensor
	chargers   map[string]*component
	pvs        map[string]*component
	rootSensor *sensor
}

type component struct {
	id string

	demand   float64
	setPoint float64
}
