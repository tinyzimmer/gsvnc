package display

import (
	"github.com/go-vgo/robotgo"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/types"
)

func (d *Display) syncToClipboard(ev *types.ClientCutText) { robotgo.WriteAll(toUTF8(ev.Text)) }

func toUTF8(in []byte) string {
	buf := make([]rune, len(in))
	for i, b := range in {
		buf[i] = rune(b)
	}
	return string(buf)
}
