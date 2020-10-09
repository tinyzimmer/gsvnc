package events

import (
	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/display"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

// PointerEvent handles pointer events.
type PointerEvent struct{}

// Code returns the code.
func (s *PointerEvent) Code() uint8 { return 5 }

// Handle handles the event.
func (s *PointerEvent) Handle(buf *buffer.ReadWriter, d *display.Display) error {
	var req types.PointerEvent
	if err := buf.ReadInto(&req); err != nil {
		return err
	}
	d.DispatchPointerEvent(&req)
	return nil
}
