package json2msgp_test

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
	"encoding/hex"
	"strings"
	"testing"

	"github.com/ndau/json2msgp"
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

			got, err := json2msgp.Convert(tt.in, nil)
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
	// The bytes we use for these sample system variables were gotten using the actual encoded
	// bytes used on the blockchain.  This way, the tests assert that when we convert from json
	// to msgp, it'll match the proper msgp format for these system variables.  For example:
	// To get the input string:
	//   ./chaos get sysvar EAIFeeTable -m | jq . -S -c
	// To get the output bytes:
	//   ./chaos get sysvar EAIFeeTable | base64 --decode | hexdump -ve '1/1 "%.2x "'
	tests := []struct {
		name    string
		in      string
		hints   map[string][]string
		wantOut string
		wantErr bool
	}{
		{
			"AccountAttributes",
			`{"ndaegwggj8qv7tqccvz6ffrthkbnmencp9t2y4mn89gdq3yk":{"x":{}}}`,
			nil,
			"81 d9 30 6e 64 61 65 67 77 67 67 6a 38 71 76 37 74 71 63 63 76 7a 36 66 66 72 74 68 6b 62 6e 6d 65 6e 63 70 39 74 32 79 34 6d 6e 38 39 67 64 71 33 79 6b 81 a1 78 80",
			false,
		},
		{
			"CommandValidatorChangeAddress",
			`["ndnf9ffbzhyf8mk7z5vvqc4quzz5i2exp5zgsmhyhc9cuwr4"]`,
			nil,
			"91 d9 30 6e 64 6e 66 39 66 66 62 7a 68 79 66 38 6d 6b 37 7a 35 76 76 71 63 34 71 75 7a 7a 35 69 32 65 78 70 35 7a 67 73 6d 68 79 68 63 39 63 75 77 72 34",
			false,
		},
		{
			"DefaultRecourseDuration",
			`172800000000`,
			nil,
			"d3 00 00 00 28 3b ae c0 00",
			false,
		},
		{
			"EAIFeeTable",
			`[{"Fee":4000000,"To":["ndaea8w9gz84ncxrytepzxgkg9ymi4k7c9p427i6b57xw3r4"]},{"Fee":1000000,"To":["ndmmw2cwhhgcgk9edp5tiieqab3pq7uxdic2wabzx49twwxh"]},{"Fee":100000,"To":["ndakj49v6nnbdq3yhnf8f2j6ivfzicedvfwtunckivfsw9qt"]},{"Fee":100000,"To":["ndnf9ffbzhyf8mk7z5vvqc4quzz5i2exp5zgsmhyhc9cuwr4"]},{"Fee":9800000,"To":null}]`,
			nil,
			"95 82 a3 46 65 65 d2 00 3d 09 00 a2 54 6f 91 d9 30 6e 64 61 65 61 38 77 39 67 7a 38 34 6e 63 78 72 79 74 65 70 7a 78 67 6b 67 39 79 6d 69 34 6b 37 63 39 70 34 32 37 69 36 62 35 37 78 77 33 72 34 82 a3 46 65 65 d2 00 0f 42 40 a2 54 6f 91 d9 30 6e 64 6d 6d 77 32 63 77 68 68 67 63 67 6b 39 65 64 70 35 74 69 69 65 71 61 62 33 70 71 37 75 78 64 69 63 32 77 61 62 7a 78 34 39 74 77 77 78 68 82 a3 46 65 65 d2 00 01 86 a0 a2 54 6f 91 d9 30 6e 64 61 6b 6a 34 39 76 36 6e 6e 62 64 71 33 79 68 6e 66 38 66 32 6a 36 69 76 66 7a 69 63 65 64 76 66 77 74 75 6e 63 6b 69 76 66 73 77 39 71 74 82 a3 46 65 65 d2 00 01 86 a0 a2 54 6f 91 d9 30 6e 64 6e 66 39 66 66 62 7a 68 79 66 38 6d 6b 37 7a 35 76 76 71 63 34 71 75 7a 7a 35 69 32 65 78 70 35 7a 67 73 6d 68 79 68 63 39 63 75 77 72 34 82 a3 46 65 65 d2 00 95 89 40 a2 54 6f c0",
			false,
		},
		{
			"LockedRateTable",
			`[[7776000000000,10000000000],[15552000000000,20000000000],[31536000000000,30000000000],[63072000000000,40000000000],[94608000000000,50000000000]]`,
			map[string][]string{"": []string{"int64", "uint64"}},
			"95 92 d3 00 00 07 12 7d b7 c0 00 cf 00 00 00 02 54 0b e4 00 92 d3 00 00 0e 24 fb 6f 80 00 cf 00 00 00 04 a8 17 c8 00 92 d3 00 00 1c ae 8c 13 e0 00 cf 00 00 00 06 fc 23 ac 00 92 d3 00 00 39 5d 18 27 c0 00 cf 00 00 00 09 50 2f 90 00 92 d3 00 00 56 0b a4 3b a0 00 cf 00 00 00 0b a4 3b 74 00",
			false,
		},
		{
			"MinDurationBetweenNodeRewardNominations",
			`86400000000`,
			nil,
			"d3 00 00 00 14 1d d7 60 00",
			false,
		},
		{
			"MinNodeRegistrationStakeAmount",
			`100000000000`,
			nil,
			"d3 00 00 00 17 48 76 e8 00",
			false,
		},
		{
			"NodeGoodnessFunction",
			`"oACI"`,
			nil,
			"c4 03 a0 00 88",
			false,
		},
		{
			"NodeRewardNominationTimeout",
			`30000000`,
			nil,
			"d2 01 c9 c3 80",
			false,
		},
		{
			"NominateNodeRewardAddress",
			`["ndnf9ffbzhyf8mk7z5vvqc4quzz5i2exp5zgsmhyhc9cuwr4"]`,
			nil,
			"91 d9 30 6e 64 6e 66 39 66 66 62 7a 68 79 66 38 6d 6b 37 7a 35 76 76 71 63 34 71 75 7a 7a 35 69 32 65 78 70 35 7a 67 73 6d 68 79 68 63 39 63 75 77 72 34",
			false,
		},
		{
			"ReleaseFromEndowmentAddress",
			`["ndmfgnz9qby6nyi35aadjt9nasjqxqyd4vrswucwfmceqs3y"]`,
			nil,
			"91 d9 30 6e 64 6d 66 67 6e 7a 39 71 62 79 36 6e 79 69 33 35 61 61 64 6a 74 39 6e 61 73 6a 71 78 71 79 64 34 76 72 73 77 75 63 77 66 6d 63 65 71 73 33 79",
			false,
		},
		{
			"TransactionFeeScript",
			`"oAAgiA=="`,
			nil,
			"c4 04 a0 00 20 88",
			false,
		},
		{
			"UnlockedRateTable",
			`[[2592000000000,20000000000],[5184000000000,30000000000],[7776000000000,40000000000],[10368000000000,50000000000],[12960000000000,60000000000],[15552000000000,70000000000],[18144000000000,80000000000],[20736000000000,90000000000],[23328000000000,100000000000]]`,
			map[string][]string{"": []string{"int64", "uint64"}},
			"99 92 d3 00 00 02 5b 7f 3d 40 00 cf 00 00 00 04 a8 17 c8 00 92 d3 00 00 04 b6 fe 7a 80 00 cf 00 00 00 06 fc 23 ac 00 92 d3 00 00 07 12 7d b7 c0 00 cf 00 00 00 09 50 2f 90 00 92 d3 00 00 09 6d fc f5 00 00 cf 00 00 00 0b a4 3b 74 00 92 d3 00 00 0b c9 7c 32 40 00 cf 00 00 00 0d f8 47 58 00 92 d3 00 00 0e 24 fb 6f 80 00 cf 00 00 00 10 4c 53 3c 00 92 d3 00 00 10 80 7a ac c0 00 cf 00 00 00 12 a0 5f 20 00 92 d3 00 00 12 db f9 ea 00 00 cf 00 00 00 14 f4 6b 04 00 92 d3 00 00 15 37 79 27 40 00 cf 00 00 00 17 48 76 e8 00",
			false,
		},
		{
			"svi",
			`{"CommandValidatorChangeAddress":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Q29tbWFuZFZhbGlkYXRvckNoYW5nZUFkZHJlc3M="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Q29tbWFuZFZhbGlkYXRvckNoYW5nZUFkZHJlc3M="]},"DefaultRecourseDuration":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","RGVmYXVsdFNldHRsZW1lbnREdXJhdGlvbg=="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","RGVmYXVsdFNldHRsZW1lbnREdXJhdGlvbg=="]},"EAIFeeTable":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","RUFJRmVlVGFibGU="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","RUFJRmVlVGFibGU="]},"LockedRateTable":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TG9ja2VkUmF0ZVRhYmxl"],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TG9ja2VkUmF0ZVRhYmxl"]},"MinDurationBetweenNodeRewardNominations":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TWluRHVyYXRpb25CZXR3ZWVuTm9kZVJld2FyZE5vbWluYXRpb25z"],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TWluRHVyYXRpb25CZXR3ZWVuTm9kZVJld2FyZE5vbWluYXRpb25z"]},"MinNodeRegistrationStakeAmount":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TWluTm9kZVJlZ2lzdHJhdGlvblN0YWtlQW1vdW50"],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","TWluTm9kZVJlZ2lzdHJhdGlvblN0YWtlQW1vdW50"]},"NodeGoodnessFunction":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9kZUdvb2RuZXNzRnVuY3Rpb24="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9kZUdvb2RuZXNzRnVuY3Rpb24="]},"NodeRewardNominationTimeout":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9kZVJld2FyZE5vbWluYXRpb25UaW1lb3V0"],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9kZVJld2FyZE5vbWluYXRpb25UaW1lb3V0"]},"NominateNodeRewardAddress":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9taW5hdGVOb2RlUmV3YXJkQWRkcmVzcw=="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","Tm9taW5hdGVOb2RlUmV3YXJkQWRkcmVzcw=="]},"ReleaseFromEndowmentAddress":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","UmVsZWFzZUZyb21FbmRvd21lbnRBZGRyZXNz"],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","UmVsZWFzZUZyb21FbmRvd21lbnRBZGRyZXNz"]},"TransactionFeeScript":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","VHJhbnNhY3Rpb25GZWVTY3JpcHQ="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","VHJhbnNhY3Rpb25GZWVTY3JpcHQ="]},"UnlockedRateTable":{"ChangeOn":0,"Current":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","VW5sb2NrZWRSYXRlVGFibGU="],"Future":["A2etqqaA3qQExilg+ywQ4ElRsyoDJh9lR5A+Thg5PcTR","VW5sb2NrZWRSYXRlVGFibGU="]}}`,
			map[string][]string{"ChangeOn": []string{"uint64"}},
			"8c bd 43 6f 6d 6d 61 6e 64 56 61 6c 69 64 61 74 6f 72 43 68 61 6e 67 65 41 64 64 72 65 73 73 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 1d 43 6f 6d 6d 61 6e 64 56 61 6c 69 64 61 74 6f 72 43 68 61 6e 67 65 41 64 64 72 65 73 73 a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 1d 43 6f 6d 6d 61 6e 64 56 61 6c 69 64 61 74 6f 72 43 68 61 6e 67 65 41 64 64 72 65 73 73 b7 44 65 66 61 75 6c 74 52 65 63 6f 75 72 73 65 44 75 72 61 74 69 6f 6e 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 19 44 65 66 61 75 6c 74 53 65 74 74 6c 65 6d 65 6e 74 44 75 72 61 74 69 6f 6e a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 19 44 65 66 61 75 6c 74 53 65 74 74 6c 65 6d 65 6e 74 44 75 72 61 74 69 6f 6e ab 45 41 49 46 65 65 54 61 62 6c 65 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 0b 45 41 49 46 65 65 54 61 62 6c 65 a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 0b 45 41 49 46 65 65 54 61 62 6c 65 af 4c 6f 63 6b 65 64 52 61 74 65 54 61 62 6c 65 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 0f 4c 6f 63 6b 65 64 52 61 74 65 54 61 62 6c 65 a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 0f 4c 6f 63 6b 65 64 52 61 74 65 54 61 62 6c 65 d9 27 4d 69 6e 44 75 72 61 74 69 6f 6e 42 65 74 77 65 65 6e 4e 6f 64 65 52 65 77 61 72 64 4e 6f 6d 69 6e 61 74 69 6f 6e 73 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 27 4d 69 6e 44 75 72 61 74 69 6f 6e 42 65 74 77 65 65 6e 4e 6f 64 65 52 65 77 61 72 64 4e 6f 6d 69 6e 61 74 69 6f 6e 73 a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 27 4d 69 6e 44 75 72 61 74 69 6f 6e 42 65 74 77 65 65 6e 4e 6f 64 65 52 65 77 61 72 64 4e 6f 6d 69 6e 61 74 69 6f 6e 73 be 4d 69 6e 4e 6f 64 65 52 65 67 69 73 74 72 61 74 69 6f 6e 53 74 61 6b 65 41 6d 6f 75 6e 74 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 1e 4d 69 6e 4e 6f 64 65 52 65 67 69 73 74 72 61 74 69 6f 6e 53 74 61 6b 65 41 6d 6f 75 6e 74 a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 1e 4d 69 6e 4e 6f 64 65 52 65 67 69 73 74 72 61 74 69 6f 6e 53 74 61 6b 65 41 6d 6f 75 6e 74 b4 4e 6f 64 65 47 6f 6f 64 6e 65 73 73 46 75 6e 63 74 69 6f 6e 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 14 4e 6f 64 65 47 6f 6f 64 6e 65 73 73 46 75 6e 63 74 69 6f 6e a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 14 4e 6f 64 65 47 6f 6f 64 6e 65 73 73 46 75 6e 63 74 69 6f 6e bb 4e 6f 64 65 52 65 77 61 72 64 4e 6f 6d 69 6e 61 74 69 6f 6e 54 69 6d 65 6f 75 74 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 1b 4e 6f 64 65 52 65 77 61 72 64 4e 6f 6d 69 6e 61 74 69 6f 6e 54 69 6d 65 6f 75 74 a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 1b 4e 6f 64 65 52 65 77 61 72 64 4e 6f 6d 69 6e 61 74 69 6f 6e 54 69 6d 65 6f 75 74 b9 4e 6f 6d 69 6e 61 74 65 4e 6f 64 65 52 65 77 61 72 64 41 64 64 72 65 73 73 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 19 4e 6f 6d 69 6e 61 74 65 4e 6f 64 65 52 65 77 61 72 64 41 64 64 72 65 73 73 a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 19 4e 6f 6d 69 6e 61 74 65 4e 6f 64 65 52 65 77 61 72 64 41 64 64 72 65 73 73 bb 52 65 6c 65 61 73 65 46 72 6f 6d 45 6e 64 6f 77 6d 65 6e 74 41 64 64 72 65 73 73 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 1b 52 65 6c 65 61 73 65 46 72 6f 6d 45 6e 64 6f 77 6d 65 6e 74 41 64 64 72 65 73 73 a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 1b 52 65 6c 65 61 73 65 46 72 6f 6d 45 6e 64 6f 77 6d 65 6e 74 41 64 64 72 65 73 73 b4 54 72 61 6e 73 61 63 74 69 6f 6e 46 65 65 53 63 72 69 70 74 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 14 54 72 61 6e 73 61 63 74 69 6f 6e 46 65 65 53 63 72 69 70 74 a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 14 54 72 61 6e 73 61 63 74 69 6f 6e 46 65 65 53 63 72 69 70 74 b1 55 6e 6c 6f 63 6b 65 64 52 61 74 65 54 61 62 6c 65 83 a8 43 68 61 6e 67 65 4f 6e 00 a7 43 75 72 72 65 6e 74 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 11 55 6e 6c 6f 63 6b 65 64 52 61 74 65 54 61 62 6c 65 a6 46 75 74 75 72 65 92 c4 21 03 67 ad aa a6 80 de a4 04 c6 29 60 fb 2c 10 e0 49 51 b3 2a 03 26 1f 65 47 90 3e 4e 18 39 3d c4 d1 c4 11 55 6e 6c 6f 63 6b 65 64 52 61 74 65 54 61 62 6c 65",
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want, err := hex.DecodeString(strings.Replace(tt.wantOut, " ", "", -1))
			require.NoError(t, err)

			in := bytes.NewBufferString(tt.in)
			out := &bytes.Buffer{}
			if err := json2msgp.ConvertStream(in, out, tt.hints); (err != nil) != tt.wantErr {
				t.Errorf("ConvertStream() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			require.Equal(t, want, out.Bytes())
		})
	}
}
