// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package slip implements a codec for the Serial Line Internet Protocol (SLIP).

See https://en.wikipedia.org/wiki/Serial_Line_Internet_Protocol.
*/
package slip

import (
	"fmt"
)

const (
	Start    byte = 0xAB
	StartEsc byte = 0xAC

	End    byte = 0xBC
	EndEsc byte = 0xBD

	Esc    byte = 0xCD
	EscEsc byte = 0xCE
)

// Encode encodes the given data as a SLIP message.
func Encode(data []byte) (enc []byte) {
	// Count the number of characters that will need to be escaped so we can
	// pre-calculate the length of the buffer
	resCount := ReservedCharCount(data)

	enc = make([]byte, len(data)+2+resCount)

	enc[0] = Start

	j := 1
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case Start:
			enc[j] = Esc
			enc[j+1] = StartEsc
			j += 2
		case End:
			enc[j] = Esc
			enc[j+1] = EndEsc
			j += 2
		case Esc:
			enc[j] = Esc
			enc[j+1] = EscEsc
			j += 2

		default:
			enc[j] = data[i]
			j += 1
		}
	}

	enc[len(enc)-1] = End

	return
}

// Decode decodes a SLIP message and returns the data or an error if there was
// one.
func Decode(enc []byte) (data []byte, err error) {
	if len(enc) < 2 {
		return nil, fmt.Errorf("slip.Decode: packet is too short %d", len(enc))
	}
	if enc[0] != Start {
		return nil, fmt.Errorf("slip.Decode: expected SLIP start (0x%02X) but got 0x%02X", Start, enc[0])
	}
	if enc[len(enc)-1] != End {
		return nil, fmt.Errorf("slip.Decode: expected SLIP end (0x%02X) but got 0x%02X", End, enc[0])
	}

	enc = enc[1 : len(enc)-1] // discard the start/end characters
	data = make([]byte, 0, len(enc))

	var c byte
	for i := 0; i < len(enc); i++ {
		if enc[i] == Esc {
			// Make sure there is at least one more character
			if i+1 >= len(enc) {
				return nil, fmt.Errorf("slip.Decode: expected an escaped character but reached the end of the buffer")
			}

			i += 1 // skip to the next character

			switch enc[i] {
			case StartEsc:
				c = Start
			case EndEsc:
				c = End
			case EscEsc:
				c = Esc
			default:
				return nil, fmt.Errorf("slip.Decode: expected an escaped character but got 0x%02X", enc[i])
			}
		} else {
			c = enc[i]
		}

		data = append(data, c)
	}

	return
}

// IsReservedChar returns true if the character is reserved.
func IsReservedChar(b byte) bool {
	return b == Start || b == End || b == Esc
}

// ReservedCharCount returns the number of characters that need to be escaped.
func ReservedCharCount(data []byte) int {
	count := 0
	for _, b := range data {
		if IsReservedChar(b) {
			count += 1
		}
	}
	return count
}
