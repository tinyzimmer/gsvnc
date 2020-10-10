package events

import (
	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/display"
	"github.com/tinyzimmer/gsvnc/pkg/internal/log"
)

// SetEncodings handles the client set-encodings event.
type SetEncodings struct{}

// Code returns the code.
func (s *SetEncodings) Code() uint8 { return 2 }

// Handle handles the event.
func (s *SetEncodings) Handle(buf *buffer.ReadWriter, d *display.Display) error {
	if err := buf.ReadPadding(1); err != nil {
		return err
	}

	var numEncodings uint16
	if err := buf.Read(&numEncodings); err != nil {
		return err
	}
	encTypes := make([]int32, int(numEncodings))
	for i := 0; i < int(numEncodings); i++ {
		if err := buf.Read(&encTypes[i]); err != nil {
			return err
		}
	}

	log.Infof("Client encodings: %#v", encTypes)
	d.SetEncodings(encTypes)

	return nil
}
