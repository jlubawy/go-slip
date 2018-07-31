// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slip

import (
	"bytes"
	"encoding/hex"
	"strings"
	"testing"
)

func mustDecodeHex(s string) []byte {
	data, err := hex.DecodeString(strings.Replace(s, " ", "", -1))
	if err != nil {
		panic(err)
	}
	return data
}

func TestStandardScanner(t *testing.T) {
	var cases = []struct {
		expectErr bool
		input     []byte
		outputs   [][]byte
	}{
		{
			expectErr: false,
			input:     mustDecodeHex("010203C0 04DBDCC0 DBDD05C0"),
			outputs: [][]byte{
				mustDecodeHex("010203"),
				mustDecodeHex("04C0"),
				mustDecodeHex("DB05"),
			},
		},
	}
	for i, tc := range cases {
		t.Logf("Test case %d", i)

		scanner := NewScanner(StdEncoding, bytes.NewReader(tc.input))
		actuals := make([][]byte, 0)
		for scanner.Scan() {
			actuals = append(actuals, scanner.Bytes())
		}
		if err := scanner.Err(); err != nil {
			if !tc.expectErr {
				t.Errorf("unexpected error: %v", err)
			}
		} else {
			if tc.expectErr {
				t.Error("expected error")
			} else {
				if len(actuals) != len(tc.outputs) {
					t.Errorf("expected %d outputs but only got %d", len(tc.outputs), len(actuals))
				} else {
					for j := 0; j < len(tc.outputs); j++ {
						if !bytes.Equal(tc.outputs[j], actuals[j]) {
							t.Errorf("mismatch at index %d", j)
							t.Errorf("   expected=% X", tc.outputs[j])
							t.Errorf("   actual  =% X", actuals[j])
						}
					}
				}
			}
		}
	}
}

func TestBluefruitScanner(t *testing.T) {
	var cases = []struct {
		expectErr bool
		input     []byte
		outputs   [][]byte
	}{
		{
			expectErr: false,
			input:     mustDecodeHex("AB010203BC AB04CDACCDBDBC ABCDCE05BC"),
			outputs: [][]byte{
				mustDecodeHex("010203"),
				mustDecodeHex("04ABBC"),
				mustDecodeHex("CD05"),
			},
		},

		// Invalid control character
		{
			expectErr: true,
			input:     mustDecodeHex("AB04CDABCDBDBC"),
			outputs:   nil,
		},
	}
	for i, tc := range cases {
		t.Logf("Test case %d", i)

		scanner := NewScanner(BluefruitEncoding, bytes.NewReader(tc.input))
		actuals := make([][]byte, 0)
		for scanner.Scan() {
			actuals = append(actuals, scanner.Bytes())
		}
		if err := scanner.Err(); err != nil {
			if !tc.expectErr {
				t.Errorf("unexpected error: %v", err)
			}
		} else {
			if tc.expectErr {
				t.Error("expected error")
			} else {
				if len(actuals) != len(tc.outputs) {
					t.Errorf("expected %d outputs but only got %d", len(tc.outputs), len(actuals))
				} else {
					for j := 0; j < len(tc.outputs); j++ {
						if !bytes.Equal(tc.outputs[j], actuals[j]) {
							t.Errorf("mismatch at index %d", j)
							t.Errorf("   expected=% X", tc.outputs[j])
							t.Errorf("   actual  =% X", actuals[j])
						}
					}
				}
			}
		}
	}
}

func TestStandardEncode(t *testing.T) {
	var cases = []struct {
		input  []byte
		output []byte
	}{
		{
			input:  mustDecodeHex("010203 C0   DC DB   DD"),
			output: mustDecodeHex("010203 DBDC DC DBDD DD C0"),
		},
	}

	for i, tc := range cases {
		t.Logf("Test case %d", i)

		enc := StdEncoding.Encode(tc.input)
		if !bytes.Equal(tc.output, enc) {
			t.Error("data mismatch")
			t.Errorf("expected: %X", tc.output)
			t.Errorf("actual  : %X", enc)
		}
	}
}
