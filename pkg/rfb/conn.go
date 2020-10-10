package rfb

import (
	"log"
	"net"

	"github.com/tinyzimmer/gsvnc/pkg/internal/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/internal/display"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/events"
)

// Conn represents a client connection.
type Conn struct {
	c       net.Conn
	s       *Server
	buf     *buffer.ReadWriter
	display *display.Display
}

func (s *Server) newConn(c net.Conn) *Conn {
	buf := buffer.NewReadWriteBuffer(c)
	conn := &Conn{
		c:       c,
		s:       s,
		buf:     buf,
		display: display.NewDisplay(s.width, s.height, buf, s.GetEncoding),
	}
	return conn
}

func (c *Conn) serve() {
	defer c.c.Close()

	if err := c.display.Start(); err != nil {
		log.Println("Error starting display:", err)
		return
	}
	defer c.display.Close()

	// Get a map of event handlers for this connection
	eventHandlers := c.s.GetEventHandlerMap()
	defer events.CloseEventHandlers(eventHandlers)

	// handle events
	for {
		cmd, err := c.buf.ReadByte()
		if err != nil {
			log.Println("Client disconnect:", err.Error())
			return
		}
		if hdlr, ok := eventHandlers[cmd]; ok {
			if err := hdlr.Handle(c.buf, c.display); err != nil {
				log.Printf("Error handling cmd %d: %s", cmd, err.Error())
				return
			}
		} else {
			log.Printf("unsupported command type %d from client\n", int(cmd))
		}
	}
}
