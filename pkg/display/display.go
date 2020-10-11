package display

import (
	"image"

	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/display/providers"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/encodings"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

// Display represents a session with the local display. It manages the gstreamer pipelines
// and listens for events from the RFB event handlers.
type Display struct {
	displayProvider providers.Display

	width, height    int
	pixelFormat      *types.PixelFormat
	getEncodingsFunc GetEncodingsFunc
	encodings        []int32
	pseudoEncodings  []int32
	currentEnc       encodings.Encoding

	// Read/writer for the connected client
	buf *buffer.ReadWriter

	// Incoming event queues
	fbReqQueue chan *types.FrameBufferUpdateRequest
	ptrEvQueue chan *types.PointerEvent
	keyEvQueue chan *types.KeyEvent
	cutTxtEvsQ chan *types.ClientCutText

	// Memory of keys that are currently down. Reiterated in order
	// on every down subsequent down event.
	downKeys []uint32
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

// GetEncodingsFunc is a function that can be used to retrieve an encoder
// from a list of client supplied options.
type GetEncodingsFunc func(encs []int32) encodings.Encoding

// Opts represents options for building a new display.
type Opts struct {
	DisplayProvider providers.Provider
	Width, Height   int
	Buffer          *buffer.ReadWriter
	GetEncodingFunc GetEncodingsFunc
}

// NewDisplay returns a new display with the given dimensions. These
// dimensions can be mutated later on depending on client support.
func NewDisplay(opts *Opts) *Display {
	return &Display{
		displayProvider:  providers.GetDisplayProvider(opts.DisplayProvider),
		width:            opts.Width,
		height:           opts.Height,
		buf:              opts.Buffer,
		getEncodingsFunc: opts.GetEncodingFunc,
		pixelFormat:      DefaultPixelFormat,
		// Buffered channels
		fbReqQueue: make(chan *types.FrameBufferUpdateRequest, 128),
		ptrEvQueue: make(chan *types.PointerEvent, 128),
		keyEvQueue: make(chan *types.KeyEvent, 128),
		cutTxtEvsQ: make(chan *types.ClientCutText, 128),
		// down key memory
		downKeys: make([]uint32, 0),
	}
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
func (d *Display) SetEncodings(encs []int32, pseudoEns []int32) {
	d.encodings = encs
	d.pseudoEncodings = pseudoEns
	d.currentEnc = d.getEncodingsFunc(encs)
}

// GetCurrentEncoding returns the encoder that is currently being used.
func (d *Display) GetCurrentEncoding() encodings.Encoding { return d.currentEnc }

// GetLastImage returns the most recent frame for the display.
func (d *Display) GetLastImage() *image.RGBA { return d.displayProvider.PullFrame() }

// DispatchFrameBufferUpdate dispatches a FrameBufferUpdateRequest on the request queue.
func (d *Display) DispatchFrameBufferUpdate(req *types.FrameBufferUpdateRequest) { d.fbReqQueue <- req }

// DispatchKeyEvent dispatches a key event to the queue.
func (d *Display) DispatchKeyEvent(ev *types.KeyEvent) { d.keyEvQueue <- ev }

// DispatchPointerEvent dispatches a pointer event to the queue.
func (d *Display) DispatchPointerEvent(ev *types.PointerEvent) { d.ptrEvQueue <- ev }

// Start will start the underlying display provider.
func (d *Display) Start() error {
	if err := d.displayProvider.Start(d.GetDimensions()); err != nil {
		return err
	}
	go d.watchChannels()
	return nil
}

// Close will stop the gstreamer pipeline.
func (d *Display) Close() error {
	close(d.fbReqQueue)
	close(d.ptrEvQueue)
	close(d.keyEvQueue)
	close(d.cutTxtEvsQ)
	return d.displayProvider.Close()
}
