/*
Copyright (c) 2022 Aton-Kish

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

/*
Package packet provides RCON packet structure and handlers.
*/
package packet

import (
	"bytes"
	"encoding/binary"

	"github.com/Aton-Kish/gorcon/types"
)

var (
	order = binary.LittleEndian
)

type Header struct {
	Length    int32
	RequestID int32
	Type      types.Packet
}

type Packet struct {
	Header
	Payload []byte
}

func NewPacket(id int32, typ types.Packet, payload []byte) *Packet {
	l := int32(4 + 4 + len(payload) + 1 + 1)
	// Request ID                :                4 bit
	// Packet Type               :                4 bit
	// Payload (NULL-terminated) : len(payload) + 1 bit
	// 1-byte Pad                :                1 bit

	h := Header{Length: l, RequestID: id, Type: typ}
	pac := Packet{Header: h, Payload: payload}

	return &pac
}

func Pack(raw []byte) (*Packet, error) {
	r := bytes.NewReader(raw)

	h := new(Header)
	if err := binary.Read(r, order, h); err != nil {
		return nil, err
	}

	p := make([]byte, h.Length-(4+4+2))
	if err := binary.Read(r, order, p); err != nil {
		return nil, err
	}

	pac := Packet{Header: *h, Payload: p}
	return &pac, nil
}

func Unpack(pac *Packet) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, order, pac.Length); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, order, pac.RequestID); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, order, pac.Type); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, order, pac.Payload); err != nil {
		return nil, err
	}

	if err := binary.Write(buf, order, [2]byte{}); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
