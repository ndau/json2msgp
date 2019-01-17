package json2msgp

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
	"reflect"
	"unicode/utf8"

	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// - if the string is not valid utf-8, it is passed through as a byte array without modification.
// - if the string is valid padded base64 in the standard encoding, it is decoded and represented in the MSGP as a byte array.
// - otherwise, it is assumed to be a string, and represented as a string.
func stringHeuristic(s string, buffer []byte) []byte {
	if !utf8.ValidString(s) {
		return msgp.AppendBytes(buffer, []byte(s))
	}
	b64bytes, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return msgp.AppendBytes(buffer, b64bytes)
	}
	return msgp.AppendString(buffer, s)
}

func convertMapStrStr(m map[string]string, b []byte) []byte {
	sz := uint32(len(m))
	b = msgp.AppendMapHeader(b, sz)
	for key, val := range m {
		b = msgp.AppendString(b, key)
		b = stringHeuristic(val, b)
	}
	return b
}

func convertMapStrIntf(m map[string]interface{}, b []byte) ([]byte, error) {
	sz := uint32(len(m))
	b = msgp.AppendMapHeader(b, sz)
	var err error
	for key, val := range m {
		b = msgp.AppendString(b, key)
		b, err = convert(val, b)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}

func convert(in interface{}, buffer []byte) ([]byte, error) {
	switch x := in.(type) {
	case map[string]interface{}:
		return convertMapStrIntf(x, buffer)
	case map[string]string:
		return convertMapStrStr(x, buffer), nil
	case []interface{}:
		buffer = msgp.AppendArrayHeader(buffer, uint32(len(x)))
		var err error
		for _, v := range x {
			buffer, err = convert(v, buffer)
			if err != nil {
				return buffer, err
			}
		}
		return buffer, nil
	}

	var err error
	v := reflect.ValueOf(in)
	switch v.Kind() {
	case reflect.String:
		return stringHeuristic(in.(string), buffer), nil
	case reflect.Ptr:
		return convert(v.Elem(), buffer)
	case reflect.Array, reflect.Slice:
		l := v.Len()
		buffer = msgp.AppendArrayHeader(buffer, uint32(l))
		for i := 0; i < l; i++ {
			buffer, err = convert(v.Index(i).Interface(), buffer)
			if err != nil {
				return buffer, err
			}
		}
		return buffer, nil
	default:
		return msgp.AppendIntf(buffer, in)
	}
}

// Convert recursively converts the input object into a MSGP representation.
//
// Strings are converted using the following heuristic:
//
// - if the string is not valid utf-8, it is passed through as a byte array without modification.
// - if the string is valid padded base64 in the standard encoding, it is decoded and represented in the MSGP as a byte array.
// - otherwise, it is assumed to be a string, and represented as a string.
//
// This is primarily intended to assist conversion from JSON to MSGP, so certain
// conversions such as structs are intentionally excluded. If you have a struct,
// use `msgp.Marshal` directly.
func Convert(in interface{}) ([]byte, error) {
	buffer := make([]byte, 0)
	return convert(in, buffer)
}

// ConvertStream reads JSON from `in` and copies it as MSGP to `out` until EOF.
//
// Strings are converted using the following heuristic:
//
// - if the string is not valid utf-8, it is passed through as a byte array without modification.
// - if the string is valid padded base64 in the standard encoding, it is decoded and represented in the MSGP as a byte array.
// - otherwise, it is assumed to be a string, and represented as a string.
func ConvertStream(in io.Reader, out io.Writer) error {
	// JSON isn't length-prefixed, so we kind of have to parse the whole thing.
	// It's a nice convenience function, at least, and we all have Effectively
	// Infinite Memory, right?
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(in)
	if err != nil {
		return errors.Wrap(err, "ConvertStream reading input")
	}

	var jsobj interface{}
	err = json.Unmarshal(buffer.Bytes(), jsobj)
	if err != nil {
		return errors.Wrap(err, "ConvertStream unmarshalling JSON")
	}

	msgp, err := Convert(jsobj)
	if err != nil {
		return err
	}

	_, err = out.Write(msgp)
	if err != nil {
		return errors.Wrap(err, "ConvertStream writing to out stream")
	}

	return nil
}
