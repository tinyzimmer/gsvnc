package types

// PixelFormat represents the current pixel format for a display.
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

// FrameBufferUpdateRequest represents a request to update the frame buffer.
type FrameBufferUpdateRequest struct {
	IncrementalFlag     uint8
	X, Y, Width, Height uint16
}

// Incremental returns true if the incremental flag is set on this request.
func (r *FrameBufferUpdateRequest) Incremental() bool { return r.IncrementalFlag != 0 }

// KeyEvent represents an RFB key event.
type KeyEvent struct {
	DownFlag uint8
	Key      uint32
}

// IsDown returns true if the event is a down event.
func (k *KeyEvent) IsDown() bool { return k.DownFlag != 0 }

// PointerEvent represents an RFB pointer event.
type PointerEvent struct {
	ButtonMask uint8
	X, Y       uint16
}

// FrameBufferRectangle represents a frame buffer rectangle.
type FrameBufferRectangle struct {
	X, Y          uint16
	Width, Height uint16
	EncType       int32
}
