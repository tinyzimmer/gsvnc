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

// DefaultEvents is a list of the default enabled event handlers.
var DefaultEvents = []Event{
	&SetEncodings{},
	&SetPixelFormat{},
	&FrameBufferUpdate{},
	&KeyEvent{},
	&PointerEvent{},
	&ClientCutText{},
}

// GetDefaults returns a slice of the default event handlers.
func GetDefaults() []Event {
	out := make([]Event, len(DefaultEvents))
	for i, t := range DefaultEvents {
		out[i] = t
	}
	return out
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
