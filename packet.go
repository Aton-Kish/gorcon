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
	"encoding/binary"
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

func (p *packet) length() int32 {
	// Request ID                :                4 bit
	// Packet Type               :                4 bit
	// Payload (NULL-terminated) : len(payload) + 1 bit
	// 1-byte Pad                :                1 bit
	return int32(4 + 4 + (len(p.payload) + 1) + 1)
}

func (p *packet) encode(w io.Writer) error {
	l := p.length()
	if err := binary.Write(w, binary.LittleEndian, &l); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, &p.requestId); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, &p.packetType); err != nil {
		return err
	}

	if err := binary.Write(w, binary.LittleEndian, p.payload); err != nil {
		return err
	}

	// NOTE: payload is NULL-terminated
	if err := binary.Write(w, binary.LittleEndian, []byte{0x00}); err != nil {
		return err
	}

	// NOTE: packet has 1-byte pad
	if err := binary.Write(w, binary.LittleEndian, []byte{0x00}); err != nil {
		return err
	}

	return nil
}

func (p *packet) decode(r io.Reader) error {
	var l int32
	if err := binary.Read(r, binary.LittleEndian, &l); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &p.requestId); err != nil {
		return err
	}

	if err := binary.Read(r, binary.LittleEndian, &p.packetType); err != nil {
		return err
	}

	p.payload = make([]byte, l-(4+4+1+1))
	if err := binary.Read(r, binary.LittleEndian, p.payload); err != nil {
		return err
	}

	// NOTE: payload is NULL-terminated
	if err := binary.Read(r, binary.LittleEndian, make([]byte, 1)); err != nil {
		return err
	}

	// NOTE: packet has 1-byte pad
	if err := binary.Read(r, binary.LittleEndian, make([]byte, 1)); err != nil {
		return err
	}

	return nil
}
