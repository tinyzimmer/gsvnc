package buffer

import (
	"bufio"
	"encoding/binary"
	"errors"
	"net"
	"reflect"
)

// ReadWriter is a buffer read/writer for RFB conncetions. It is held in a separate
// package to be passed easily between handlers.
//
// It doesn't implement an actual io.ReadWriter, rather is intended soley for use
// by the rfb package.
type ReadWriter struct {
	br *bufio.Reader
	bw *bufio.Writer

	wq chan []byte
}

// NewReadWriteBuffer returns a new ReadWriter for the given connection.
func NewReadWriteBuffer(c net.Conn) *ReadWriter {
	rw := &ReadWriter{
		br: bufio.NewReader(c),
		bw: bufio.NewWriter(c),
		wq: make(chan []byte, 100),
	}
	go func() {
		for msg := range rw.wq {
			rw.write(msg)
			rw.flush()
		}
	}()
	return rw
}

// Close will stop this buffer from processing messages.
func (rw *ReadWriter) Close() {
	close(rw.wq)
}

// Reader returns a direct reference to the underlying reader.
func (rw *ReadWriter) Reader() *bufio.Reader { return rw.br }

// ReadByte reads a single byte from the buffer.
func (rw *ReadWriter) ReadByte() (byte, error) {
	b, err := rw.br.ReadByte()
	if err != nil {
		return 0, err
	}
	return b, nil
}

// ReadPadding pops padding off the read buffer of the given size.
func (rw *ReadWriter) ReadPadding(size int) error {
	for i := 0; i < size; i++ {
		if _, err := rw.ReadByte(); err != nil {
			return err
		}
	}
	return nil
}

// Read will read from the buffer into the given interface. This method
// is not intended to be used with structs. Use ReadInto for that.
func (rw *ReadWriter) Read(v interface{}) error {
	return binary.Read(rw.br, binary.BigEndian, v)
}

// ReadInto reflects on the given struct and populates its fields from the
// read buffer. The struct fields must be in the order they appear on the
// buffer.
func (rw *ReadWriter) ReadInto(data interface{}) error {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("Data is invalid (nil or non-pointer)")
	}
	val := reflect.ValueOf(data).Elem()
	nVal := rv.Elem()
	for i := 0; i < val.NumField(); i++ {
		nvField := nVal.Field(i)
		if err := rw.Read(nvField.Addr().Interface()); err != nil {
			return err
		}
	}
	return nil
}

// Write writes the given interface to the buffer.
func (rw *ReadWriter) write(v interface{}) error {
	return binary.Write(rw.bw, binary.BigEndian, v)
}

// Flush will flush the contents of the write buffer.
func (rw *ReadWriter) flush() error {
	return rw.bw.Flush()
}

// Dispatch will push packed message(s) onto the buffer queue.
func (rw *ReadWriter) Dispatch(msg []byte) { rw.wq <- msg }
