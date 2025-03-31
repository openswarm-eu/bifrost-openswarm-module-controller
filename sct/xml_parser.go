package sct

import (
	"encoding/xml"
	"fmt"
	"io"
)

// Structs to represent the XML structure
type Model struct {
	XMLName xml.Name `xml:"model"`
	Data    Data     `xml:"data"`
}

type Data struct {
	States      []State      `xml:"state"`
	Events      []Event      `xml:"event"`
	Transitions []Transition `xml:"transition"`
}

type State struct {
	ID      string `xml:"id,attr"`
	Name    string `xml:"name,attr"`
	Initial string `xml:"initial,attr"`
	Marked  string `xml:"marked,attr"`
	X       string `xml:"x,attr"`
	Y       string `xml:"y,attr"`
}

type Event struct {
	ID           string `xml:"id,attr"`
	Name         string `xml:"name,attr"`
	Controllable string `xml:"controllable,attr"`
	Observable   string `xml:"observable,attr"`
}

type Transition struct {
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
	Event  string `xml:"event,attr"`
}

// Function to parse the XML file
func parseXML(content io.Reader) (*Model, error) {
	var model Model
	decoder := xml.NewDecoder(content)
	if err := decoder.Decode(&model); err != nil {
		return nil, fmt.Errorf("failed to decode XML: %v", err)
	}

	return &model, nil
}
