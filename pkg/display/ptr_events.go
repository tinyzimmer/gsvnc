package display

import (
	"github.com/go-vgo/robotgo"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

func (d *Display) servePointerEvent(ev *types.PointerEvent) {
	btns := make(map[string]bool)
	for mask, maskType := range btnMasks {
		btns[maskType] = nthBitOf(ev.ButtonMask, mask) == 1
	}
	// This is just a mouse move event
	robotgo.MoveMouse(int(ev.X), int(ev.Y))
}

var btnMasks = map[int]string{
	0: "left",
	1: "middle",
	2: "right",
	3: "scroll-up",
	4: "scroll-down",
	5: "scroll-left",
	6: "scroll-right",
	7: "unhandled",
}

func nthBitOf(bit uint8, n int) uint8 {
	return (bit & (1 << n)) >> n
}
