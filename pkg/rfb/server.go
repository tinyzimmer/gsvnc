package rfb

import (
	"log"
	"net"
)

// NewServer creates a new RFB server with an initial width and height.
func NewServer(width, height int) *Server {
	return &Server{
		width:  width,
		height: height,
	}
}

// Server represents an RFB server. A channel is exposed for handling incoming client
// connections.
type Server struct {
	width, height int
	conns         chan *Conn // read/write version of Conns

	// Conns is a channel of incoming connections.
	Conns <-chan *Conn
}

// Serve binds the RFB server to the given listener and starts serving connections.
func (s *Server) Serve(ln net.Listener) error {
	for {

		// Accept a new connection
		c, err := ln.Accept()
		if err != nil {
			return err
		}

		// Create a new client connection
		conn := s.newConn(c)

		// Do the rfb handshake
		if err := conn.doHandshake(); err != nil {
			log.Println("Error during server-client handshake:", err.Error())
			conn.c.Close()
			continue
		}

		// handle events
		go conn.serve()
	}
}
