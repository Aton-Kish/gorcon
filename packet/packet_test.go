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

package packet

import (
	"testing"

	"github.com/Aton-Kish/gorcon/types"
	"github.com/stretchr/testify/assert"
)

var packetTestCases = []struct {
	name   string
	raw    []byte
	packet Packet
}{
	{
		name: "Valid Case: AuthRequest",
		raw: []byte{
			// Length: 14
			0x0E, 0x00, 0x00, 0x00,
			// RequestId: 123456
			0x40, 0xE2, 0x01, 0x00,
			// Type: AuthRequest (=3)
			0x03, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): "auth"
			0x61, 0x75, 0x74, 0x68, 0x00,
			// 1-byte Pad
			0x00,
		},
		packet: Packet{Header: Header{Length: 14, RequestID: 123456, Type: types.AuthRequest}, Payload: []byte("auth")},
	},
	{
		name: "Valid Case: AuthResponse",
		raw: []byte{
			// Length: 10
			0x0A, 0x00, 0x00, 0x00,
			// RequestId: 789012
			0x14, 0x0A, 0x0C, 0x00,
			// Type: AuthResponse (=2)
			0x02, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): ""
			0x00,
			// 1-byte Pad
			0x00,
		},
		packet: Packet{Header: Header{Length: 10, RequestID: 789012, Type: types.AuthResponse}, Payload: []byte("")},
	},
	{
		name: "Valid Case: CommandRequest",
		raw: []byte{
			// Length: 17
			0x11, 0x00, 0x00, 0x00,
			// RequestId: 345678
			0x4E, 0x46, 0x05, 0x00,
			// Type: CommandRequest (=2)
			0x02, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): "command"
			0x63, 0x6F, 0x6D, 0x6D, 0x61, 0x6E, 0x64, 0x00,
			// 1-byte Pad
			0x00,
		},
		packet: Packet{Header: Header{Length: 17, RequestID: 345678, Type: types.CommandRequest}, Payload: []byte("command")},
	},
	{
		name: "Valid Case: CommandResponse",
		raw: []byte{
			// Length: 18
			0x12, 0x00, 0x00, 0x00,
			// RequestId: 901234
			0x72, 0xC0, 0x0D, 0x00,
			// Type: CommandResponse (=0)
			0x00, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): "response"
			0x72, 0x65, 0x73, 0x70, 0x6F, 0x6E, 0x73, 0x65, 0x00,
			// 1-byte Pad
			0x00,
		},
		packet: Packet{Header: Header{Length: 18, RequestID: 901234, Type: types.CommandResponse}, Payload: []byte("response")},
	},
	{
		name: "Valid Case: DummyRequest",
		raw: []byte{
			// Length: 23
			0x17, 0x00, 0x00, 0x00,
			// RequestId: 567890
			0x52, 0xAA, 0x08, 0x00,
			// Type: DummyRequest (=100)
			0x64, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): "dummy request"
			0x64, 0x75, 0x6D, 0x6D, 0x79, 0x20, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x00,
			// 1-byte Pad
			0x00,
		},
		packet: Packet{Header: Header{Length: 23, RequestID: 567890, Type: types.DummyRequest}, Payload: []byte("dummy request")},
	},
	{
		name: "Valid Case: UnknownResponse",
		raw: []byte{
			// Length: 28
			0x1C, 0x00, 0x00, 0x00,
			// RequestId: 123456
			0x40, 0xE2, 0x01, 0x00,
			// Type: CommandResponse (=0)
			0x00, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): "Unknown request 64"
			0x55, 0x6E, 0x6B, 0x6E, 0x6F, 0x77, 0x6E, 0x20, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x20, 0x36, 0x34, 0x00,
			// 1-byte Pad
			0x00,
		},
		packet: Packet{Header: Header{Length: 28, RequestID: 123456, Type: types.CommandResponse}, Payload: []byte("Unknown request 64")},
	},
}

func TestNewPacket(t *testing.T) {
	type testCase struct {
		name    string
		id      int32
		typ     types.Packet
		payload []byte
		want    Packet
	}

	cases := make([]testCase, 0, len(packetTestCases))

	for _, v := range packetTestCases {
		c := testCase{
			name:    v.name,
			id:      v.packet.RequestID,
			typ:     v.packet.Type,
			payload: v.packet.Payload,
			want:    v.packet,
		}
		cases = append(cases, c)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pac := NewPacket(c.id, c.typ, c.payload)
			assert.Equal(t, c.want, *pac)
		})
	}
}

func TestPack(t *testing.T) {
	type testCase struct {
		name string
		raw  []byte
		want Packet
	}

	cases := make([]testCase, 0, len(packetTestCases))

	for _, v := range packetTestCases {
		c := testCase{
			name: v.name,
			raw:  v.raw,
			want: v.packet,
		}
		cases = append(cases, c)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			pac, err := Pack(c.raw)
			assert.NoError(t, err)
			assert.Equal(t, c.want, *pac)
		})
	}
}

func TestUnpack(t *testing.T) {
	type testCase struct {
		name   string
		packet Packet
		want   []byte
	}

	cases := make([]testCase, 0, len(packetTestCases))

	for _, v := range packetTestCases {
		c := testCase{
			name:   v.name,
			packet: v.packet,
			want:   v.raw,
		}
		cases = append(cases, c)
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			raw, err := Unpack(&c.packet)
			assert.NoError(t, err)
			assert.Equal(t, c.want, raw)
		})
	}
}
