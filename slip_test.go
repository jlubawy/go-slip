// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package slip

import (
	"bytes"
	"encoding/hex"
	"testing"
)

func mustDecodeHex(s string) []byte {
	data, err := hex.DecodeString(s)
	if err != nil {
		panic(err)
	}
	return data
}

func TestEncode(t *testing.T) {
	var cases = []struct {
		expectErr bool
		input     []byte
		output    []byte
	}{
		{
			expectErr: false,
			input:     mustDecodeHex("0601EFCD060A01265C0000BC010000D6BE898E402100FD0327D3DAD80201061106BA5689A6FABFA2BD01467D6E00FBABAD05160A1805061E77F2"),
			output:    mustDecodeHex("AB0601EFCDCE060A01265C0000CDBD010000D6BE898E402100FD0327D3DAD80201061106BA5689A6FABFA2BD01467D6E00FBCDACAD05160A1805061E77F2BC"),
		},
		{
			expectErr: false,
			input:     mustDecodeHex(""),
			output:    mustDecodeHex("ABBC"),
		},
	}

	for i, tc := range cases {
		t.Logf("Test case %d", i)

		enc := Encode(tc.input)
		if !bytes.Equal(tc.output, enc) {
			t.Error("data mismatch")
			t.Errorf("expected: %X", tc.output)
			t.Errorf("actual  : %X", enc)
		}
	}
}

func TestDecode(t *testing.T) {
	var cases = []struct {
		expectErr bool
		input     []byte
		output    []byte
	}{
		{
			expectErr: false,
			input:     mustDecodeHex("AB0601EFCDCE060A01265C0000CDBD010000D6BE898E402100FD0327D3DAD80201061106BA5689A6FABFA2BD01467D6E00FBCDACAD05160A1805061E77F2BC"),
			output:    mustDecodeHex("0601EFCD060A01265C0000BC010000D6BE898E402100FD0327D3DAD80201061106BA5689A6FABFA2BD01467D6E00FBABAD05160A1805061E77F2"),
		},

		// Empty packet
		{
			expectErr: true,
			input:     mustDecodeHex(""),
			output:    mustDecodeHex(""),
		},
		// Missing start
		{
			expectErr: true,
			input:     mustDecodeHex("00BC"),
			output:    mustDecodeHex(""),
		},
		// Missing end
		{
			expectErr: true,
			input:     mustDecodeHex("AB00"),
			output:    mustDecodeHex(""),
		},
		// Short escape sequence
		{
			expectErr: true,
			input:     mustDecodeHex("ABCDBC"),
			output:    mustDecodeHex(""),
		},
		// Non-escape character
		{
			expectErr: true,
			input:     mustDecodeHex("ABCD01BC"),
			output:    mustDecodeHex(""),
		},
	}

	for i, tc := range cases {
		t.Logf("Test case %d", i)

		data, err := Decode(tc.input)
		if tc.expectErr {
			if err == nil {
				t.Error("expected error")
			}
		} else {
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else {
				if !bytes.Equal(tc.output, data) {
					t.Error("data mismatch")
					t.Errorf("expected: %X", tc.output)
					t.Errorf("actual  : %X", data)
				}
			}
		}
	}
}
