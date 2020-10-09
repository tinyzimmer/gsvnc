package util

import (
	"encoding/binary"
	"errors"
	"io"
	"reflect"
)

// PackStruct will reflect over the given pointer and write the fields
// to the buffer in the order of declaration. This function uses BigEndian
// encoding.
func PackStruct(buf io.Writer, data interface{}) error {
	rv := reflect.ValueOf(data)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("Data is invalid (nil or non-pointer)")
	}
	val := reflect.ValueOf(data).Elem()
	nVal := rv.Elem()
	for i := 0; i < val.NumField(); i++ {
		nvField := nVal.Field(i)
		var wdata interface{}
		if nvField.Kind() == reflect.String {
			str := nvField.Interface().(string)
			wdata = []byte(str)
		} else {
			wdata = nvField.Interface()
		}
		if err := binary.Write(buf, binary.BigEndian, wdata); err != nil {
			return err
		}
	}
	return nil
}

// Write is a convenience wrapper for using the binary package to write to a buffer.
func Write(buf io.Writer, v interface{}) error {
	return binary.Write(buf, binary.BigEndian, v)
}
