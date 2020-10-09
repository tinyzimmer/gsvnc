package versions

import (
	"fmt"
	"log"

	"github.com/tinyzimmer/gsvnc/pkg/buffer"
)

// Protocol version strings
const (
	V3 = "RFB 003.003\n"
	V7 = "RFB 003.007\n"
	V8 = "RFB 003.008\n"
)

// NegotiateProtocolVersion will negotiate the protocol version with the given connection.
func NegotiateProtocolVersion(buf *buffer.ReadWriter) (string, error) {
	buf.Dispatch([]byte(V8))

	sl, err := buf.Reader().ReadSlice('\n')
	if err != nil {
		return "", fmt.Errorf("reading client protocol version: %v", err)
	}
	ver := string(sl)
	log.Printf("client wants: %q", ver)
	switch ver {
	case V3, V7, V8: // cool.
	default:
		return "", fmt.Errorf("bogus client-requested security type %q", ver)
	}
	return ver, nil
}
