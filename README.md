# `json2msgp`: heuristic recursive self-describing message transformation

JSON is a well-known human-readable self-describing message format. [MsgPack](https://msgpack.org/index.html) is a well-known binary self-describing message format.

It is reasonably straightforward to transform arbitrary MsgPack data into JSON; there are [library methods](https://godoc.org/github.com/tinylib/msgp/msgp#CopyToJSON) available to perform this conversion.

However, the reverse operation is not well-supported. This is because the MSGP type system is a superset of JSON's. In particular, MSGP supports byte arrays, and JSON does not. Typical conversions from MSGP to JSON, such as the one linked above, will transform byte arrays into base64-encoded strings. This works well, but when moving from JSON to MSGP, there is no way to know with certainty whether a given string is intended to be a string or a byte array.

This library uses a very simple heuristic to make that decision:

- if the JSON string is not valid utf-8, it is passed through as a byte array without modification.
- if the JSON string is valid padded base64 in the standard encoding, it is decoded and represented in the MSGP as a byte array.
- otherwise, it is assumed to be a string, and represented as a string.
