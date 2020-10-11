package auth

import (
	"github.com/tinyzimmer/gsvnc/pkg/buffer"
)

// Type represents an authentication type.
type Type interface {
	Code() uint8
	Negotiate(wr *buffer.ReadWriter) error
}

// DefaultAuthTypes is the default enabled list of auth types.
var DefaultAuthTypes = []Type{
	&None{},
	&VNCAuth{},
	&TightSecurity{},
}

// GetDefaults returns a slice of the default auth handlers.
func GetDefaults() []Type {
	out := make([]Type, len(DefaultAuthTypes))
	for i, t := range DefaultAuthTypes {
		out[i] = t
	}
	return out
}
