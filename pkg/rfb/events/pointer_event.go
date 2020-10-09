package events

import (
	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/display"
)

// PointerEvent handles pointer events.
type PointerEvent struct{}

// Code returns the code.
func (s *PointerEvent) Code() uint8 { return 5 }

type rfbPointerEvent struct {
	ButtonMask uint8
	X, Y       uint16
}

// Handle handles the event.
func (s *PointerEvent) Handle(buf *buffer.ReadWriter, d *display.Display) error {
	var req rfbPointerEvent
	if err := buf.ReadInto(&req); err != nil {
		return err
	}
	// TODO
	return nil
}
