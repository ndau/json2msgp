package json2msgp

// ----- ---- --- -- -
// Copyright 2020 The Axiom Foundation. All Rights Reserved.
//
// Licensed under the Apache License 2.0 (the "License").  You may not use
// this file except in compliance with the License.  You can obtain a copy
// in the file LICENSE in the source distribution or at
// https://www.apache.org/licenses/LICENSE-2.0.txt
// - -- --- ---- -----

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"unicode/utf8"

	"github.com/oneiro-ndev/ndaumath/pkg/address"
	"github.com/pkg/errors"
	"github.com/tinylib/msgp/msgp"
)

// Converter manages state during the converstion process.
type Converter struct {
	// The key we're currently processing.
	currentKey string              

	// Use this map with the current key to find its expected type.
	typeHints map[string][]string

	// When there are multiple types per hint name, it is used with arrays of values in json.
	// This is an index into the []string of typeHints[currentKey].
	currentHint int
}

// - if the string is not valid utf-8, it is passed through as a byte array without modification.
// - if the string is a valid ndau address, it's represented as a string.
// - if the string is valid padded base64 in the standard encoding, it is decoded and represented in the MSGP as a byte array.
// - otherwise, it is assumed to be a string, and represented as a string.
func (c *Converter) stringHeuristic(s string, buffer []byte) []byte {
	if !utf8.ValidString(s) {
		return msgp.AppendBytes(buffer, []byte(s))
	}
	_, err := address.Validate(s)
	if err == nil {
		return msgp.AppendString(buffer, s)
	}
	b64bytes, err := base64.StdEncoding.DecodeString(s)
	if err == nil {
		return msgp.AppendBytes(buffer, b64bytes)
	}
	return msgp.AppendString(buffer, s)
}

func (c *Converter) convertMapStrStr(m map[string]string, b []byte) []byte {
	sz := uint32(len(m))
	b = msgp.AppendMapHeader(b, sz)

	// sort keys for deterministic output
	// not critical for actual behavior, but we can't really test properly
	// without this
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	for _, key := range keys {
		val := m[key]
		b = msgp.AppendString(b, key)
		c.currentKey = key
		b = c.stringHeuristic(val, b)
	}
	return b
}

func (c *Converter) convertMapStrIntf(m map[string]interface{}, b []byte) ([]byte, error) {
	sz := uint32(len(m))
	b = msgp.AppendMapHeader(b, sz)

	// sort keys for deterministic output
	// not critical for actual behavior, but we can't really test properly
	// without this
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

	var err error
	for _, key := range keys {
		val := m[key]
		b = msgp.AppendString(b, key)
		c.currentKey = key
		b, err = c.convert(val, b)
		if err != nil {
			return b, err
		}
	}
	return b, nil
}

func (c *Converter) convert(in interface{}, buffer []byte) ([]byte, error) {
	switch x := in.(type) {
	case string:
		return c.stringHeuristic(x, buffer), nil
	case map[string]interface{}:
		return c.convertMapStrIntf(x, buffer)
	case map[string]string:
		return c.convertMapStrStr(x, buffer), nil
	case []interface{}:
		buffer = msgp.AppendArrayHeader(buffer, uint32(len(x)))
		var err error
		// Because we reset this every time, this only works on the innermost of nested arrays.
		// TODO: Generalize the type hint spec.  The way this is done now, with the % operator
		// to grab currentHint below, is just a minimal solution to account for arrays with
		// unnamed values.  We might consider generalized nested arrays, and also allowing a
		// single-type hint without having to be inside an array within the hint json.
		c.currentHint = 0
		for _, v := range x {
			buffer, err = c.convert(v, buffer)
			if err != nil {
				return buffer, err
			}
			c.currentHint++
		}
		return buffer, nil
	case float64:
		// The json input treats all numeric values as float64.  Without knowing the original
		// data type, we don't know how to encode numeric values.  First, see if there's a hint.
		if c.typeHints != nil {
			if typeHint, ok := c.typeHints[c.currentKey]; ok {
				currentHint := typeHint[c.currentHint % len(typeHint)]
				// Support type hints for all msgp numeric formats.  We don't ensure that the
				// value fits into the hinted type.  If there is a casting problem, the tool's
				// user will have to supply a different type hint, or alter the input json.
				switch currentHint {
				case "byte":
					return msgp.AppendByte(buffer, byte(x)), nil
				case "float32":
					return msgp.AppendFloat32(buffer, float32(x)), nil
				case "float64":
					return msgp.AppendFloat64(buffer, x), nil
				case "int":
					return msgp.AppendInt(buffer, int(x)), nil
				case "int8":
					return msgp.AppendInt8(buffer, int8(x)), nil
				case "int16":
					return msgp.AppendInt16(buffer, int16(x)), nil
				case "int32":
					return msgp.AppendInt32(buffer, int32(x)), nil
				case "int64":
					return msgp.AppendInt64(buffer, int64(x)), nil
				case "uint":
					return msgp.AppendUint(buffer, uint(x)), nil
				case "uint8":
					return msgp.AppendUint8(buffer, uint8(x)), nil
				case "uint16":
					return msgp.AppendUint16(buffer, uint16(x)), nil
				case "uint32":
					return msgp.AppendUint32(buffer, uint32(x)), nil
				case "uint64":
					return msgp.AppendUint64(buffer, uint64(x)), nil
				default:
					return buffer, fmt.Errorf(
						"Unsupported numeric type hint %s=%s", c.currentKey, currentHint)
				}
			}
		}

		// Most of what we encode are of type int64, so we make that assumption here as part of
		// this heuristic if we didn't find a type hint for it.
		i := int64(x)
		// Make sure the value is indeed an integer (has no fractional part).  This is meant as
		// a convenience check for the tool's user.  We'll error below if this check fails.
		if float64(i) == x {
			return msgp.AppendInt64(buffer, i), nil
		}

		// We error here, rather than encoding to general float64.  Otherwise we could wind up
		// encoding a blob of json containing multiple occurrences of a given variable (e.g. an
		// array of objects), some of which get encoded one way, the rest another way.  In that
		// case, when msgp unmarshals it later, it won't be able to handle the two different ways
		// we encode the numeric values.  So, it's better to make this clear at encode-time.
		return buffer, fmt.Errorf("Unsupported numeric value %v", x)
	}

	var err error
	v := reflect.ValueOf(in)
	switch v.Kind() {
	case reflect.Ptr:
		return c.convert(v.Elem().Interface(), buffer)
	case reflect.Array, reflect.Slice:
		l := v.Len()
		buffer = msgp.AppendArrayHeader(buffer, uint32(l))
		for i := 0; i < l; i++ {
			buffer, err = c.convert(v.Index(i).Interface(), buffer)
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
// - if the string is a valid ndau address, it's represented as a string.
// - if the string is valid padded base64 in the standard encoding, it is decoded and represented in the MSGP as a byte array.
// - otherwise, it is assumed to be a string, and represented as a string.
//
// This is primarily intended to assist conversion from JSON to MSGP, so certain
// conversions such as structs are intentionally excluded. If you have a struct,
// use `msgp.Marshal` directly.
func Convert(in interface{}, typeHints map[string][]string) ([]byte, error) {
	buffer := make([]byte, 0)
	c := Converter{typeHints: typeHints}
	return c.convert(in, buffer)
}

// ConvertStream reads JSON from `in` and copies it as MSGP to `out` until EOF.
//
// Strings are converted using the following heuristic:
//
// - if the string is not valid utf-8, it is passed through as a byte array without modification.
// - if the string is a valid ndau address, it's represented as a string.
// - if the string is valid padded base64 in the standard encoding, it is decoded and represented in the MSGP as a byte array.
// - otherwise, it is assumed to be a string, and represented as a string.
//
// Numeric valuse need extra help to know their types.  Use the typeHints map for that.
//
// - if all "Fee" variables are to be encoded as int64, and "ChangeOn" as uint64, then use:
//   typeHints = {"Fee": []string{"int64"}, "ChangeOn": []string{"uint64"}}
// - if there are blobs of json without names, yet there are arrays of differing numeric types,
//   such as: [[0,1],[-2,3],[4,5]], then use:
//   typeHints = {"": []string{"int64", "uint64"}}
func ConvertStream(in io.Reader, out io.Writer, typeHints map[string][]string) error {
	// JSON isn't length-prefixed, so we kind of have to parse the whole thing.
	// It's a nice convenience function, at least, and we all have Effectively
	// Infinite Memory, right?
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(in)
	if err != nil {
		return errors.Wrap(err, "ConvertStream reading input")
	}

	var jsobj interface{}
	err = json.Unmarshal(buffer.Bytes(), &jsobj)
	if err != nil {
		return errors.Wrap(err, "ConvertStream unmarshalling JSON")
	}

	msgp, err := Convert(jsobj, typeHints)
	if err != nil {
		return err
	}

	_, err = out.Write(msgp)
	if err != nil {
		return errors.Wrap(err, "ConvertStream writing to out stream")
	}

	return nil
}
