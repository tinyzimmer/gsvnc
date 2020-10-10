package auth

import "github.com/tinyzimmer/gsvnc/pkg/internal/buffer"

// None represents no authentication.
type None struct{}

// Code returns the code for no-auth.
func (a *None) Code() uint8 { return 1 }

// Negotiate immediately returns nil.
func (a *None) Negotiate(rw *buffer.ReadWriter) error { return nil }
