package events

import (
	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/display"
)

// KeyEvent handles key events.
type KeyEvent struct{}

// Code returns the code.
func (s *KeyEvent) Code() uint8 { return 4 }

type rfbKeyEvent struct {
	DownFlag uint8
	Key      uint32
}

// Handle handles the event.
func (s *KeyEvent) Handle(buf *buffer.ReadWriter, d *display.Display) error {
	var req rfbKeyEvent
	if err := buf.ReadInto(&req); err != nil {
		return err
	}
	// TODO
	return nil
}
