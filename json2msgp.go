package json2msgp

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

// Convert recursively converts the input object into a MSGP representation.
//
// Strings are converted using the following heuristic:
//
// - if the string is not valid utf-8, it is passed through as a byte array without modification.
// - if the string is valid padded base64 in the standard encoding, it is decoded and represented in the MSGP as a byte array.
// - otherwise, it is assumed to be a string, and represented as a string.
func Convert(in interface{}) ([]byte, error) {
	return nil, errors.New("unimplemented")
}

// ConvertStream reads JSON from `in` and copies it as MSGP to `out` until EOF.
//
// Strings are converted using the following heuristic:
//
// - if the string is not valid utf-8, it is passed through as a byte array without modification.
// - if the string is valid padded base64 in the standard encoding, it is decoded and represented in the MSGP as a byte array.
// - otherwise, it is assumed to be a string, and represented as a string.
func ConvertStream(in io.Reader, out io.Writer) error {
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
