package events

import (
	"log"

	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/display"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

// SetPixelFormat handles the client set-pixel-format event.
type SetPixelFormat struct{}

// Code returns the code.
func (s *SetPixelFormat) Code() uint8 { return 0 }

// Handle handles the event.
func (s *SetPixelFormat) Handle(buf *buffer.ReadWriter, d *display.Display) error {
	log.Printf("handling setpixel format")

	if err := buf.ReadPadding(3); err != nil {
		return err
	}

	var pf types.PixelFormat
	if err := buf.ReadInto(&pf); err != nil {
		return err
	}

	log.Printf("Client wants pixel format: %#v", pf)
	d.SetPixelFormat(&pf)

	if err := buf.ReadPadding(3); err != nil {
		return err
	}
	return nil
}
