package rfb

import (
	"bytes"
	"fmt"
	"io"
	"log"

	"github.com/tinyzimmer/gsvnc/pkg/buffer"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/auth"
	"github.com/tinyzimmer/gsvnc/pkg/rfb/versions"
	"github.com/tinyzimmer/gsvnc/pkg/util"
)

func (c *Conn) doHandshake() error {

	ver, err := versions.NegotiateProtocolVersion(c.buf)
	if err != nil {
		return err
	}

	var authType auth.Type
	if authType, err = negotiateAuth(ver, c.buf); err != nil {
		return err
	}

	log.Println("reading client init")

	// ClientInit
	if _, err := c.buf.ReadByte(); err != nil {
		return err
	}

	log.Println("sending server init")
	format := c.display.GetPixelFormat()

	// 6.3.2. ServerInit
	width, height := c.display.GetDimensions()
	buf := new(bytes.Buffer)
	util.Write(buf, uint16(width))
	util.Write(buf, uint16(height))
	util.PackStruct(buf, format)
	util.Write(buf, uint8(0)) // pad1
	util.Write(buf, uint8(0)) // pad2
	util.Write(buf, uint8(0)) // pad3
	serverName := "gsvnc"
	util.Write(buf, int32(len(serverName)))
	util.Write(buf, []byte(serverName))

	// Chcek if we are extending server init. This is only applicable to TightSecurity.
	if extender, ok := authType.(interface{ ExtendServerInit(io.Writer) }); ok {
		extender.ExtendServerInit(buf)
	}

	c.buf.Dispatch(buf.Bytes())
	return nil
}

const (
	statusOK     = 0
	statusFailed = 1
)

// NegotiateAuth wil negotiate authentication on the given connection, for the
// given version.
func negotiateAuth(ver string, rw *buffer.ReadWriter) (auth.Type, error) {
	var authType auth.Type
	buf := new(bytes.Buffer)

	if ver >= versions.V7 {
		util.Write(buf, uint8(len(auth.EnabledAuthTypes)))
		for _, t := range auth.EnabledAuthTypes {
			util.Write(buf, t.Code())
		}
		rw.Dispatch(buf.Bytes())
		wanted, err := rw.ReadByte()
		if err != nil {
			return nil, err
		}
		if !auth.IsSupported(wanted) {
			return nil, fmt.Errorf("client wanted unsupported auth type %d", int(wanted))
		}
		authType = auth.GetAuth(wanted)
	} else {
		// Old way. Just tell client we're doing no auth.
		authType = auth.GetNone()
		util.Write(buf, uint32(authType.Code()))
		rw.Dispatch(buf.Bytes())
	}

	if err := authType.Negotiate(rw); err != nil {
		buf = new(bytes.Buffer)
		util.Write(buf, uint32(statusFailed))
		rw.Dispatch(buf.Bytes())
		return nil, err
	}

	if ver >= versions.V8 {
		// 6.1.3. SecurityResult
		buf = new(bytes.Buffer)
		util.Write(buf, uint32(statusOK))
		rw.Dispatch(buf.Bytes())
	}

	return authType, nil
}
