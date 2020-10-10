package events

import (
	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/display"
)

// Event is an inteface implemented by client message handlers.
type Event interface {
	Code() uint8
	Handle(buf *buffer.ReadWriter, d *display.Display) error
}

// EnabledEvents is a list of the currently enabled event handlers.
var EnabledEvents = []Event{
	&SetEncodings{},
	&SetPixelFormat{},
	&FrameBufferUpdate{},
	&KeyEvent{},
	&PointerEvent{},
}

// GetDefaults returns a slice of the default event handlers.
func GetDefaults() []Event {
	out := make([]Event, len(EnabledEvents))
	for i, t := range EnabledEvents {
		out[i] = t
	}
	return out
}

// DisableEvent removes the given event from the list of EnabledEvents.
func DisableEvent(ev Event) {
	EnabledEvents = remove(EnabledEvents, ev)
}

func remove(ee []Event, e Event) []Event {
	newEvs := make([]Event, 0)
	for _, enabled := range ee {
		if enabled.Code() != e.Code() {
			newEvs = append(newEvs, enabled)
		}
	}
	return newEvs
}

// CloseEventHandlers will iterate each event handler in the map and if it provides a Close
// function it will execute it.
//
// This does not exist anymore, but left for convenience should it be useful in the future.
// The Display object handles most of the state regarding an RFB session.
func CloseEventHandlers(hdlrs map[uint8]Event) {
	for _, ev := range hdlrs {
		closer, ok := ev.(interface{ Close() })
		if ok {
			closer.Close()
		}
	}
}
