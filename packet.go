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
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// Type
type packetType int32

const (
	authRequestType     = packetType(3)
	authResponseType    = packetType(2)
	commandRequestType  = packetType(2)
	commandResponseType = packetType(0)
	dummyRequestType    = packetType(100)
)

// Packet
type packet struct {
	requestId  int32
	packetType packetType
	payload    []byte
}

func newPacket(id int32, typ packetType, payload []byte) *packet {
	return &packet{
		requestId:  id,
		packetType: typ,
		payload:    payload,
	}
}

func (p *packet) String() string {
	if p == nil {
		return "<nil>"
	}

	var payload string
	if p.payload == nil {
		payload = "<nil>"
	} else {
		payload = string(p.payload)
	}

	return fmt.Sprintf("Length: %d, RequestId: %d, Type: %d, Payload: %s", p.length(), p.requestId, p.packetType, payload)
}

func (p *packet) length() int {
	// Request ID                :                4 bit
	// Packet Type               :                4 bit
	// Payload (NULL-terminated) : len(payload) + 1 bit
	// 1-byte Pad                :                1 bit
	return 4 + 4 + (len(p.payload) + 1) + 1
}

func (p *packet) encode(w io.Writer) error {
	// NOTE: prevent split packets using bufio
	buf := bufio.NewWriter(w)

	l := int32(p.length())
	if err := binary.Write(buf, binary.LittleEndian, &l); err != nil {
		return &PacketError{Op: "encode", Err: err}
	}

	if err := binary.Write(buf, binary.LittleEndian, &p.requestId); err != nil {
		return &PacketError{Op: "encode", Err: err}
	}

	if err := binary.Write(buf, binary.LittleEndian, &p.packetType); err != nil {
		return &PacketError{Op: "encode", Err: err}
	}

	if err := binary.Write(buf, binary.LittleEndian, p.payload); err != nil {
		return &PacketError{Op: "encode", Err: err}
	}

	// NOTE: payload is NULL-terminated
	if err := binary.Write(buf, binary.LittleEndian, []byte{0x00}); err != nil {
		return &PacketError{Op: "encode", Err: err}
	}

	// NOTE: packet has 1-byte pad
	if err := binary.Write(buf, binary.LittleEndian, []byte{0x00}); err != nil {
		return &PacketError{Op: "encode", Err: err}
	}

	if err := buf.Flush(); err != nil {
		return &PacketError{Op: "encode", Err: err}
	}

	return nil
}

func (p *packet) decode(r io.Reader) error {
	var l int32
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return &PacketError{Op: "decode", Err: err}
	}

	if err := binary.Read(r, binary.LittleEndian, &p.requestId); err != nil {
		return &PacketError{Op: "decode", Err: err}
	}

	if err := binary.Read(r, binary.LittleEndian, &p.packetType); err != nil {
		return &PacketError{Op: "decode", Err: err}
	}

	p.payload = make([]byte, l-(4+4+1+1))
	if err := binary.Read(r, binary.LittleEndian, p.payload); err != nil {
		return &PacketError{Op: "decode", Err: err}
	}

	// NOTE: payload is NULL-terminated
	if err := binary.Read(r, binary.LittleEndian, make([]byte, 1)); err != nil {
		return &PacketError{Op: "decode", Err: err}
	}

	// NOTE: packet has 1-byte pad
	if err := binary.Read(r, binary.LittleEndian, make([]byte, 1)); err != nil {
		return &PacketError{Op: "decode", Err: err}
	}

	return nil
}
