// Copyright 2018 Josh Lubawy. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

/*
Package slip implements a codec for the Serial Line Internet Protocol (SLIP).

See https://en.wikipedia.org/wiki/Serial_Line_Internet_Protocol.
*/
package slip

import (
	"bufio"
	"fmt"
	"io"
)

// StartDisabled is used to disable Start character encoding.
const StartDisabled = -1

// An Encoding describes a SLIP encoding. If a Start character is not used by
// a particular encoding scheme then set Start to StartDisabled.
type Encoding struct {
	Start, EscStart rune
	End, EscEnd     rune
	Esc, EscEsc     rune
}

var StdEncoding = &Encoding{
	Start:    StartDisabled,
	EscStart: StartDisabled,
	End:      0xC0,
	EscEnd:   0xDC,
	Esc:      0xDB,
	EscEsc:   0xDD,
}

var BluefruitEncoding = &Encoding{
	Start:    0xAB,
	EscStart: 0xAC,
	End:      0xBC,
	EscEnd:   0xBD,
	Esc:      0xCD,
	EscEsc:   0xCE,
}

// An InvalidControlCharError is returned when a non-control character follows
// an Escape character. It gives the index in the byte slice where the character
// was found, and which character it was.
type InvalidControlCharError struct {
	Index       int
	ControlChar byte
}

func (e InvalidControlCharError) Error() string {
	return fmt.Sprintf("invalid control character 0x%02X Escaped at index %d", e.ControlChar, e.Index)
}

var _ bufio.SplitFunc = (*Encoding)(nil).SplitPackets

// SplitPackets is a split function for a bufio.Scanner that returns a packet
// for each token.
func (enc *Encoding) SplitPackets(data []byte, atEOF bool) (advance int, token []byte, err error) {
	EndIndex := -1
	tokenByteCount := 0
	for i := 0; i < len(data); i++ {
		r := rune(data[i])
		if r == enc.End {
			EndIndex = i
			break
		} else if r != enc.Esc {
			tokenByteCount += 1
		}
	}
	if EndIndex == -1 {
		if atEOF {
			advance = len(data)
			token = data
			err = io.EOF
		}
		return
	}

	StartIndex := 0
	if enc.Start != StartDisabled {
		if rune(data[0]) == enc.Start {
			StartIndex = 1
			tokenByteCount -= 1
		}
	}

	// Advance past the End character
	advance = EndIndex + 1
	token = make([]byte, tokenByteCount)

	// Decode the input
	inEscSeq := false
	j := 0
	for i := StartIndex; i < EndIndex; i++ {
		r := rune(data[i])
		if inEscSeq {
			if !enc.isValidControlEscChar(data[i]) {
				err = InvalidControlCharError{i, data[i]}
				return
			}

			inEscSeq = false

			switch r {
			case enc.EscStart:
				token[j] = byte(enc.Start)
				j += 1
			case enc.EscEnd:
				token[j] = byte(enc.End)
				j += 1
			case enc.EscEsc:
				token[j] = byte(enc.Esc)
				j += 1
			default:
				return
			}
		} else {
			switch r {
			case enc.Esc:
				inEscSeq = true
			default:
				token[j] = data[i]
				j += 1
			}
		}
	}

	return
}

// NewScanner returns a new bufio.Scanner with the split function set to SplitPackets.
func NewScanner(enc *Encoding, r io.Reader) *bufio.Scanner {
	scanner := bufio.NewScanner(r)
	scanner.Split(enc.SplitPackets)
	return scanner
}

func (enc *Encoding) minLength() int {
	if enc.Start == StartDisabled {
		return 1 // End
	}
	return 2 // Start+End
}

// isValidControlEscChar returns true if the character is a valid control character
// for a given encoding.
func (enc *Encoding) isValidControlEscChar(b byte) bool {
	rb := rune(b)
	if enc.EscStart == StartDisabled {
		return rb == enc.EscEnd || rb == enc.EscEsc
	}
	return rb == enc.EscStart || rb == enc.EscEnd || rb == enc.EscEsc
}

// isValidControlChar returns true if the character is a valid control character
// for a given encoding.
func (enc *Encoding) isValidControlChar(b byte) bool {
	rb := rune(b)
	if enc.Start == StartDisabled {
		return rb == enc.End || rb == enc.Esc
	}
	return rb == enc.Start || rb == enc.End || rb == enc.Esc
}

// controlCharCount returns the number of control characters that need to be
// Escaped for a given encoding.
func (enc *Encoding) controlCharCount(src []byte) (count int) {
	for i := 0; i < len(src); i++ {
		if enc.isValidControlChar(src[i]) {
			count += 1
		}
	}
	return
}

// EncodedLen returns the size of the destination buffer needed to encode a
// buffer for a given encoding.
func (enc *Encoding) EncodedLen(src []byte) int {
	return len(src) + enc.minLength() + enc.controlCharCount(src)
}

// Encode encodes the given data as a SLIP message.
func (enc *Encoding) Encode(src []byte) (dst []byte) {
	dst = make([]byte, enc.EncodedLen(src))

	j := 0
	if enc.Start != StartDisabled {
		dst[j] = byte(enc.Start)
		j += 1
	}

	for i := 0; i < len(src) && (j < len(dst)-1); i++ {
		if enc.isValidControlChar(src[i]) {
			dst[j] = byte(enc.Esc)
			switch rune(src[i]) {
			case enc.Start:
				dst[j+1] = byte(enc.EscStart)
			case enc.End:
				dst[j+1] = byte(enc.EscEnd)
			case enc.Esc:
				dst[j+1] = byte(enc.EscEsc)
			}
			j += 2
		} else {
			dst[j] = src[i]
			j += 1
		}
	}
	dst[j] = byte(enc.End)

	return
}
