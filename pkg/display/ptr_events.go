package display

func (d *Display) handlePtrEventsLoop() {
	for {
		select {
		case _, ok := <-d.ptrEvQueue:
			if !ok {
				// Client disconnected.
				return
			}
		}
	}
}
