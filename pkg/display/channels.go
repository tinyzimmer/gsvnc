package display

import (
	"time"
)

func (d *Display) watchChannels() {
	ticker := time.NewTicker(time.Millisecond * 100)

	for {
		select {

		// Framebuffer update requests
		case ur, ok := <-d.fbReqQueue:
			if !ok {
				// Client disconnected.
				return
			}
			d.pushFrame(ur)

		// Key events
		case ev, ok := <-d.keyEvQueue:
			if !ok {
				// Client disconnected.
				return
			}
			if ev.IsDown() {
				d.appendDownKeyIfMissing(ev.Key)
				d.dispatchDownKeys()
			} else {
				d.removeDownKey(ev.Key)
			}

		// Pointer events
		case _, ok := <-d.ptrEvQueue:
			if !ok {
				// Client disconnected.
				return
			}

		// Send a frame update every 100 msec if there are no other events
		// to serve
		case <-ticker.C:
			last := d.GetLastImage()
			d.pushImage(last)
		}
	}
}
