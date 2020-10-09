package display

import (
	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/encodings"
)

// Display represents a session with the local display.
//
// It is meant to be used as an object to be passed to event handlers.
type Display struct {
	width, height int
	pixelFormat   *PixelFormat
	encodings     []int32
	currentEnc    encodings.Encoding
	pipeline      *gst.Pipeline

	buf      *buffer.ReadWriter
	reqQueue chan *FrameBufferUpdateRequest
	queue    chan []byte
}

// PixelFormat represents the current pixel format for the display.
// Fields are declared in the order they appear in a SetPixelFormat
// request.
type PixelFormat struct {
	BPP, Depth                      uint8
	BigEndian, TrueColour           uint8 // flags; 0 or non-zero
	RedMax, GreenMax, BlueMax       uint16
	RedShift, GreenShift, BlueShift uint8
}

// IsScreensThousands returns if the format requested by the OS X "Screens" app's "Thousands" mode.
func (f *PixelFormat) IsScreensThousands() bool {
	// Note: Screens asks for Depth 16; RealVNC asks for Depth 15 (which is more accurate)
	// Accept either. Same format.
	return f.BPP == 16 && (f.Depth == 16 || f.Depth == 15) && f.TrueColour != 0 &&
		f.RedMax == 0x1f && f.GreenMax == 0x1f && f.BlueMax == 0x1f &&
		f.RedShift == 10 && f.GreenShift == 5 && f.BlueShift == 0
}

// DefaultPixelFormat is the default pixel format used in ServerInit messages.
var DefaultPixelFormat = &PixelFormat{
	BPP:        16,
	Depth:      16,
	BigEndian:  0,
	TrueColour: 1,
	RedMax:     0x1f,
	GreenMax:   0x1f,
	BlueMax:    0x1f,
	RedShift:   0xa,
	GreenShift: 0x5,
	BlueShift:  0,
}

// NewDisplay returns a new display with the given dimensions. These
// dimensions can be mutated later on depending on client support.
func NewDisplay(width, height int, buf *buffer.ReadWriter) *Display {
	display := &Display{
		width:       width,
		height:      height,
		buf:         buf,
		pixelFormat: DefaultPixelFormat,
		reqQueue:    make(chan *FrameBufferUpdateRequest, 128),
		queue:       make(chan []byte, 128),
	}
	go display.pushFramesLoop()
	return display
}

// GetDimensions returns the current dimensions of the display.
func (d *Display) GetDimensions() (width, height int) { return d.width, d.height }

// SetDimensions sets the dimensions of the display.
func (d *Display) SetDimensions(width, height int) {
	d.width = width
	d.height = height
}

// GetPixelFormat returns the current pixel format for the display.
func (d *Display) GetPixelFormat() *PixelFormat { return d.pixelFormat }

// SetPixelFormat sets the pixel format for the display.
func (d *Display) SetPixelFormat(pf *PixelFormat) { d.pixelFormat = pf }

// GetEncodings returns the encodings currently supported by the client
// connected to this display.
func (d *Display) GetEncodings() []int32 { return d.encodings }

// SetEncodings sets the encodings that the connected client supports.
func (d *Display) SetEncodings(encs []int32) {
	d.encodings = encs
	d.currentEnc = encodings.GetEncoding(encs)
}

// GetCurrentEncoding returns the encoder that is currently being used.
func (d *Display) GetCurrentEncoding() encodings.Encoding { return d.currentEnc }

// GetLastImage returns the most recent frame for the display.
func (d *Display) GetLastImage() []byte { return <-d.queue }

// Dispatch dispatches a FrameBufferUpdateRequest on the request queue.
func (d *Display) Dispatch(req *FrameBufferUpdateRequest) { d.reqQueue <- req }

// Close will stop the gstreamer pipeline.
func (d *Display) Close() error {
	close(d.reqQueue)
	return d.pipeline.Destroy()
}
