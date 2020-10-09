package events

import (
	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/display"
)

// FrameBufferUpdate handles framebuffer update events.
type FrameBufferUpdate struct {
	gotFirstFrame bool
	buf8          []uint8 // temporary buffer to avoid generating garbage
}

// Code returns the code.
func (f *FrameBufferUpdate) Code() uint8 { return 3 }

// Handle handles the event.
func (f *FrameBufferUpdate) Handle(buf *buffer.ReadWriter, d *display.Display) error {

	var req display.FrameBufferUpdateRequest
	if err := buf.ReadInto(&req); err != nil {
		return err
	}

	d.Dispatch(&req)
	return nil
}
