package display

import "github.com/tinyzimmer/gsvnc/pkg/log"

func (d *Display) handleKeyEvents() {
	for {
		select {
		case ev, ok := <-d.keyEvQueue:
			if !ok {
				// Client disconnected.
				return
			}
			log.Debug("Got key event: ", ev)
			if ev.IsDown() {
				d.appendDownKeyIfMissing(ev.Key)
				d.dispatchDownKeys()
			} else {
				d.removeDownKey(ev.Key)
			}
		}
	}
}

func (d *Display) handlePointerEvents() {
	for {
		select {
		case ev, ok := <-d.ptrEvQueue:
			if !ok {
				// Client disconnected.
				return
			}
			log.Debug("Got pointer event: ", ev)

		}
	}
}

func (d *Display) handleFrameBufferEvents() {
	for {
		select {
		// Framebuffer update requests
		case ur, ok := <-d.fbReqQueue:
			if !ok {
				// Client disconnected.
				return
			}
			log.Debug("Handling framebuffer update request")
			d.pushFrame(ur)

		// Send a frame update anyway if there are no updates on the queue
		default:
			log.Debug("Pushing latest frame to client")
			last := d.GetLastImage()
			d.pushImage(last)
		}
	}
}

func (d *Display) watchChannels() {
	go d.handleKeyEvents()
	go d.handlePointerEvents()
	go d.handleFrameBufferEvents()
}
