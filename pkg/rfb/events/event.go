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

// GetEventHandlerMap returns a map that can be used for handling events on an
// rfb connection. The states of the handlers returned remain persistent so they
// can be used consistently throughout a client session.
func GetEventHandlerMap() map[uint8]Event {
	setEncodings := &SetEncodings{}
	setPixelFormat := &SetPixelFormat{}
	fbUpdate := &FrameBufferUpdate{}
	keyEvent := &KeyEvent{}
	ptrEvent := &PointerEvent{}
	return map[uint8]Event{
		setEncodings.Code():   setEncodings,
		setPixelFormat.Code(): setPixelFormat,
		fbUpdate.Code():       fbUpdate,
		keyEvent.Code():       keyEvent,
		ptrEvent.Code():       ptrEvent,
	}
}

// CloseEventHandlers will iterate each event handler in the map and if it provides a Close
// function it will execute it.
func CloseEventHandlers(hdlrs map[uint8]Event) {
	for _, ev := range hdlrs {
		closer, ok := ev.(interface{ Close() })
		if ok {
			closer.Close()
		}
	}
}
