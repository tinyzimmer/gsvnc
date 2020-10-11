package events

import (
	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/display"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

// ClientCutText handles new text in the client's cut buffer.
type ClientCutText struct{}

// Code returns the code.
func (c *ClientCutText) Code() uint8 { return 6 }

// Handle handles the event.
func (c *ClientCutText) Handle(buf *buffer.ReadWriter, d *display.Display) error {
	var req types.ClientCutText

	buf.ReadPadding(3)

	if err := buf.Read(&req.Length); err != nil {
		return err
	}

	req.Text = make([]byte, req.Length)

	if err := buf.Read(&req.Text); err != nil {
		return err
	}

	return nil
}
