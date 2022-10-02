// Copyright (c) 2022 Aton-Kish
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package rcon

import (
	"fmt"
)

type RconError struct {
	op  string
	err error
}

func NewRconError(op string, err error) error {
	return &RconError{op, err}
}

func (e *RconError) Error() string {
	if e == nil {
		return "<nil>"
	}

	var err string
	if e.err == nil {
		err = "<nil>"
	} else {
		err = e.err.Error()
	}

	return fmt.Sprintf("failed to %s; error: %s", e.op, err)
}

func (e *RconError) Unwrap() error {
	return e.err
}

type PacketError struct {
	op     string
	packet *packet
	err    error
}

func NewPacketError(op string, packet *packet, err error) error {
	return &PacketError{op, packet, err}
}

func (e *PacketError) Error() string {
	if e == nil {
		return "<nil>"
	}

	var err string
	if e.err == nil {
		err = "<nil>"
	} else {
		err = e.err.Error()
	}

	return fmt.Sprintf("failed to %s packet{%s}; error: %s", e.op, e.packet.String(), err)
}

func (e *PacketError) Unwrap() error {
	return e.err
}
