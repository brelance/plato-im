package source

import (
	"fmt"

	"github.com/brelance/plato/common/discovery"
)

type EventType string

const (
	AddNodeEvent = "addNode"
	DelNodeEvent = "delNode"
)

// Note: Chan
var eventChan chan *Event

type Event struct {
	Type EventType
	IP   string
	Port string
	// the following two fields provide dispatcher with data for sorting
	ConnNum      float64
	MessageBytes float64
}

func NewEvent(ed *discovery.EndpointInfo) *Event {
	if ed == nil || ed.MetaData == nil {
		return nil
	}

	var connNum, msgBytes float64
	if data, ok := ed.MetaData["connect_num"]; ok {
		connNum = data.(float64)
	}

	if data, ok := ed.MetaData["message_bytes"]; ok {
		msgBytes = data.(float64)
	}
	return &Event{
		Type:         AddNodeEvent,
		IP:           ed.IP,
		Port:         ed.Port,
		ConnNum:      connNum,
		MessageBytes: msgBytes,
	}
}

func (e *Event) Key() string {
	return fmt.Sprintf("%s:%s", e.IP, e.Port)
}
