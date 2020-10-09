package display

import (
	"bytes"
	"log"
	"time"

	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
	"github.com/tinyzimmer/gsvnc/pkg/util"
)

func (d *Display) pushFramesLoop() {
	ticker := time.NewTicker(time.Millisecond * 100)
	for {
		select {
		case ur, ok := <-d.fbReqQueue:
			if !ok {
				// Client disconnected.
				return
			}
			d.pushFrame(ur)
		case <-ticker.C:
			last := d.GetLastImage()
			d.pushImage(last)
		}
	}
}

// Server -> Client
const (
	encodingCopyRect     = 1
	cmdFramebufferUpdate = 0
)

func (d *Display) pushFrame(ur *types.FrameBufferUpdateRequest) {

	li := d.GetLastImage()
	if li == nil {
		return
	}

	if ur.Incremental() {
		width, height := d.GetDimensions()
		buf := new(bytes.Buffer)

		// log.Printf("Client wants incremental update, sending none. %#v", ur)

		util.Write(buf, uint8(cmdFramebufferUpdate))
		// padding byte
		util.Write(buf, uint8(0))
		// num rectangles
		util.Write(buf, uint16(1))

		util.PackStruct(buf, &types.FrameBufferRectangle{
			X: 0, Y: 0, Width: uint16(width), Height: uint16(height), EncType: encodingCopyRect, // TODO make sure supported
		})

		util.Write(buf, uint16(0)) // src-x
		util.Write(buf, uint16(0)) // src-y

		d.buf.Dispatch(buf.Bytes())
		return
	}

	d.pushImage(li)
}

func (d *Display) pushImage(imgData []byte) {

	width, height := d.GetDimensions()

	buf := new(bytes.Buffer)

	util.Write(buf, uint8(cmdFramebufferUpdate))
	util.Write(buf, uint8(0))  // padding byte
	util.Write(buf, uint16(1)) // 1 rectangle

	//log.Printf("sending %d x %d pixels", width, height)
	format := d.GetPixelFormat()
	if format.TrueColour == 0 {
		log.Println("only true-colour supported")
		return
	}

	enc := d.GetCurrentEncoding()

	// Send that rectangle:
	util.PackStruct(buf, &types.FrameBufferRectangle{
		X: 0, Y: 0, Width: uint16(width), Height: uint16(height), EncType: enc.Code(), // TODO make sure supported
	})

	enc.HandleBuffer(buf, d.GetPixelFormat(), imgData)

	d.buf.Dispatch(buf.Bytes())
}
