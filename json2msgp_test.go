package json2msgp_test

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"

	"github.com/oneiro-ndev/json2msgp"
	"github.com/stretchr/testify/require"
)

func TestConvert(t *testing.T) {
	anint := 0xff

	tests := []struct {
		name    string
		in      interface{}
		want    string
		wantErr bool
	}{
		{"non-utf8 string", string([]byte{0xff, 0x00}), "c4 02 ff 00", false},
		{"b64 string", "DwA=", "c4 02 0f 00", false},
		{"string", "foo", "A3 66 6F 6F", false},
		{"int", 0xff, "d1 00 ff", false},
		{"int64", int64(0xff), "d1 00 ff", false},
		{"nil", nil, "c0", false},
		{"pointer to int", &anint, "d1 00 ff", false},
		{"json max int", int64(9007199254740991), "d3 00 1F FF FF FF FF FF FF", false},
		{"json max unsigned int", uint64(9007199254740991), "cf 00 1F FF FF FF FF FF FF", false},
		{"map string -> string", map[string]string{"foo": "beefeater"}, "81 a3 66 6f 6f a9 62 65 65 66 65 61 74 65 72", false},
		{"map string -> b64 string", map[string]string{"foo": "vu/q3qo="}, "81 a3 66 6f 6f c4 05 be ef ea de aa", false}, // the value is as spoken with a cold
		{"map string -> intf (int)", map[string]interface{}{"foo": 0xff}, "81 a3 66 6f 6f d1 00 ff", false},
		{"map string -> intf (b64 str)", map[string]interface{}{"foo": "vu/q3qo="}, "81 a3 66 6f 6f c4 05 be ef ea de aa", false},
		{"array of int", []int{1, 2, 3, 4}, "94 01 02 03 04", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, err := hex.DecodeString(strings.Replace(tt.want, " ", "", -1))
			require.NoError(t, err)

			got, err := json2msgp.Convert(tt.in)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.Equal(t, want, got)
		})
	}
}

// more complicated tests go here because it's easier to do complicated
// nesting structures in json than raw go
func TestConvertStream(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		wantOut string
		wantErr bool
	}{
		{
			"TransactionFeeScript",
			`"oAAgiA=="`,
			"c4 04 a0 00 20 88",
			false,
		},
		{
			"EAIFeeTable",
			`[{"Fee":4000000,"To":["ndaea8w9gz84ncxrytepzxgkg9ymi4k7c9p427i6b57xw3r4"]},{"Fee":1000000,"To":["ndmmw2cwhhgcgk9edp5tiieqab3pq7uxdic2wabzx49twwxh"]},{"Fee":100000,"To":["ndbmgby86qw9bds9f8wrzut5zrbxuehum5kvgz9sns9hgknh"]},{"Fee":100000,"To":["ndnf9ffbzhyf8mk7z5vvqc4quzz5i2exp5zgsmhyhc9cuwr4"]},{"Fee":9800000,"To":null}]`,
			"95 82 a3 46 65 65 cb 41 4e 84 80 00 00 00 00 a2 54 6f 91 d9 30 6e 64 61 65 61 38 77 39 67 7a 38 34 6e 63 78 72 79 74 65 70 7a 78 67 6b 67 39 79 6d 69 34 6b 37 63 39 70 34 32 37 69 36 62 35 37 78 77 33 72 34 82 a3 46 65 65 cb 41 2e 84 80 00 00 00 00 a2 54 6f 91 d9 30 6e 64 6d 6d 77 32 63 77 68 68 67 63 67 6b 39 65 64 70 35 74 69 69 65 71 61 62 33 70 71 37 75 78 64 69 63 32 77 61 62 7a 78 34 39 74 77 77 78 68 82 a3 46 65 65 cb 40 f8 6a 00 00 00 00 00 a2 54 6f 91 d9 30 6e 64 62 6d 67 62 79 38 36 71 77 39 62 64 73 39 66 38 77 72 7a 75 74 35 7a 72 62 78 75 65 68 75 6d 35 6b 76 67 7a 39 73 6e 73 39 68 67 6b 6e 68 82 a3 46 65 65 cb 40 f8 6a 00 00 00 00 00 a2 54 6f 91 d9 30 6e 64 6e 66 39 66 66 62 7a 68 79 66 38 6d 6b 37 7a 35 76 76 71 63 34 71 75 7a 7a 35 69 32 65 78 70 35 7a 67 73 6d 68 79 68 63 39 63 75 77 72 34 82 a3 46 65 65 cb 41 62 b1 28 00 00 00 00 a2 54 6f c0",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, err := hex.DecodeString(strings.Replace(tt.wantOut, " ", "", -1))
			require.NoError(t, err)

			in := bytes.NewBufferString(tt.in)
			out := &bytes.Buffer{}
			if err := json2msgp.ConvertStream(in, out); (err != nil) != tt.wantErr {
				t.Errorf("ConvertStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Equal(t, want, out.Bytes())
		})
	}
}
