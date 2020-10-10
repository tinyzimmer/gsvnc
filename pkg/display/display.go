package display

import (
	"image"

	"github.com/tinyzimmer/go-gst/gst"
	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/encodings"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

// Display represents a session with the local display. It manages the gstreamer pipelines
// and listens for events from the RFB event handlers.
type Display struct {
	width, height int
	pixelFormat   *types.PixelFormat
	encodings     []int32
	currentEnc    encodings.Encoding
	pipeline      *gst.Pipeline

	// Read/writer for the connected client
	buf *buffer.ReadWriter

	// Incoming event queues
	fbReqQueue chan *types.FrameBufferUpdateRequest
	ptrEvQueue chan *types.PointerEvent
	keyEvQueue chan *types.KeyEvent

	// Memory of keys that are currently down. Reiterated in order
	// on every down subsequent down event.
	downKeys []uint32

	// contains the incoming samples from the screen
	frameQueue chan *image.RGBA

	stopCh chan struct{}
}

// DefaultPixelFormat is the default pixel format used in ServerInit messages.
var DefaultPixelFormat = &types.PixelFormat{
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
		// Buffered channels
		fbReqQueue: make(chan *types.FrameBufferUpdateRequest, 128),
		ptrEvQueue: make(chan *types.PointerEvent, 128),
		keyEvQueue: make(chan *types.KeyEvent, 128),
		// Image channel
		frameQueue: make(chan *image.RGBA, 2), // A channel that will essentially only ever have the latest frame available.
		// down key memory
		downKeys: make([]uint32, 0),
		// stop channel for image capturing
		stopCh: make(chan struct{}),
	}
	go display.watchChannels()
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
func (d *Display) GetPixelFormat() *types.PixelFormat { return d.pixelFormat }

// SetPixelFormat sets the pixel format for the display.
func (d *Display) SetPixelFormat(pf *types.PixelFormat) { d.pixelFormat = pf }

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
func (d *Display) GetLastImage() *image.RGBA { return <-d.frameQueue }

// DispatchFrameBufferUpdate dispatches a FrameBufferUpdateRequest on the request queue.
func (d *Display) DispatchFrameBufferUpdate(req *types.FrameBufferUpdateRequest) { d.fbReqQueue <- req }

// DispatchKeyEvent dispatches a key event to the queue.
func (d *Display) DispatchKeyEvent(ev *types.KeyEvent) { d.keyEvQueue <- ev }

// DispatchPointerEvent dispatches a pointer event to the queue.
func (d *Display) DispatchPointerEvent(ev *types.PointerEvent) { d.ptrEvQueue <- ev }

// Close will stop the gstreamer pipeline.
func (d *Display) Close() error {
	close(d.fbReqQueue)
	close(d.ptrEvQueue)
	close(d.keyEvQueue)
	d.stopCh <- struct{}{}
	if d.pipeline != nil {
		return d.pipeline.Destroy()
	}
	return nil
}
