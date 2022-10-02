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

//go:build !e2e

package rcon

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

var packetCases = []struct {
	name   string
	packet *packet
	raw    []byte
}{
	{
		name: "positive case: Auth Request",
		packet: &packet{
			requestId:  123456,
			packetType: authRequestType,
			payload:    []byte("auth"),
		},
		raw: []byte{
			// Length: 14
			0x0E, 0x00, 0x00, 0x00,
			// Request ID: 123456
			0x40, 0xE2, 0x01, 0x00,
			// Packet Type: Auth Request (=3)
			0x03, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): "auth"
			0x61, 0x75, 0x74, 0x68, 0x00,
			// 1-byte Pad
			0x00,
		},
	},
	{
		name: "positive case: Auth Response",
		packet: &packet{
			requestId:  123456,
			packetType: authResponseType,
			payload:    []byte(""),
		},
		raw: []byte{
			// Length: 14
			0x0A, 0x00, 0x00, 0x00,
			// Request ID: 123456
			0x40, 0xE2, 0x01, 0x00,
			// Packet Type: Auth Response (=2)
			0x02, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): ""
			0x00,
			// 1-byte Pad
			0x00,
		},
	},
	{
		name: "positive case: Command Request",
		packet: &packet{
			requestId:  123456,
			packetType: commandRequestType,
			payload:    []byte("command"),
		},
		raw: []byte{
			// Length: 14
			0x11, 0x00, 0x00, 0x00,
			// Request ID: 123456
			0x40, 0xE2, 0x01, 0x00,
			// Packet Type: Command Request (=2)
			0x02, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): "command"
			0x63, 0x6F, 0x6D, 0x6D, 0x61, 0x6E, 0x64, 0x00,
			// 1-byte Pad
			0x00,
		},
	},
	{
		name: "positive case: Command Response",
		packet: &packet{
			requestId:  123456,
			packetType: commandResponseType,
			payload:    []byte("response"),
		},
		raw: []byte{
			// Length: 18
			0x12, 0x00, 0x00, 0x00,
			// Request ID: 123456
			0x40, 0xE2, 0x01, 0x00,
			// Packet Type: Command Response (=0)
			0x00, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): "response"
			0x72, 0x65, 0x73, 0x70, 0x6F, 0x6E, 0x73, 0x65, 0x00,
			// 1-byte Pad
			0x00,
		},
	},
	{
		name: "positive case: Dummy Request",
		packet: &packet{
			requestId:  123456,
			packetType: dummyRequestType,
			payload:    []byte("dummy request"),
		},
		raw: []byte{
			// Length: 23
			0x17, 0x00, 0x00, 0x00,
			// Request ID: 123456
			0x40, 0xE2, 0x01, 0x00,
			// Packet Type: Dummy Request (=100)
			0x64, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): "dummy request"
			0x64, 0x75, 0x6D, 0x6D, 0x79, 0x20, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x00,
			// 1-byte Pad
			0x00,
		},
	},
	{
		name: "positive case: Unknown Response",
		packet: &packet{
			requestId:  123456,
			packetType: commandResponseType,
			payload:    []byte("Unknown request 64"),
		},
		raw: []byte{
			// Length: 28
			0x1C, 0x00, 0x00, 0x00,
			// Request ID: 123456
			0x40, 0xE2, 0x01, 0x00,
			// Packet Type: Command Response (=0)
			0x00, 0x00, 0x00, 0x00,
			// Payload (NULL-terminated): "Unknown request 64"
			0x55, 0x6E, 0x6B, 0x6E, 0x6F, 0x77, 0x6E, 0x20, 0x72, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x20, 0x36, 0x34, 0x00,
			// 1-byte Pad
			0x00,
		},
	},
}

func Test_newPacket(t *testing.T) {
	type Case struct {
		name       string
		requestId  int32
		packetType packetType
		payload    []byte
		expected   *packet
	}

	cases := []Case{}

	for _, c := range packetCases {
		cases = append(cases, Case{
			name:       c.name,
			requestId:  c.packet.requestId,
			packetType: c.packet.packetType,
			payload:    c.packet.payload,
			expected: &packet{
				requestId:  c.packet.requestId,
				packetType: c.packet.packetType,
				payload:    c.packet.payload,
			},
		})
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual := newPacket(tt.requestId, tt.packetType, tt.payload)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func Test_packet_encode(t *testing.T) {
	type Case struct {
		name        string
		packet      *packet
		expected    []byte
		expectedErr error
	}

	cases := []Case{}

	for _, c := range packetCases {
		cases = append(cases, Case{
			name:        c.name,
			packet:      c.packet,
			expected:    c.raw,
			expectedErr: nil,
		})
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			buf := new(bytes.Buffer)
			err := tt.packet.encode(buf)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, buf.Bytes())
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func Test_packet_decode(t *testing.T) {
	type Case struct {
		name        string
		raw         []byte
		expected    *packet
		expectedErr error
	}

	cases := []Case{}

	for _, c := range packetCases {
		cases = append(cases, Case{
			name:        c.name,
			raw:         c.raw,
			expected:    c.packet,
			expectedErr: nil,
		})
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.NewBuffer(tt.raw)

			packet := new(packet)
			err := packet.decode(buf)

			if tt.expectedErr == nil {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, packet)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}

func Test_packet_net(t *testing.T) {
	type Case struct {
		name      string
		packet    *packet
		clientErr error
		serverErr error
	}

	cases := []Case{}

	for _, c := range packetCases {
		cases = append(cases, Case{
			name:      c.name,
			packet:    c.packet,
			clientErr: nil,
			serverErr: nil,
		})
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			srv, clt := pipe()
			defer clt.Close()

			errCh := make(chan error, 1)
			defer close(errCh)

			packetCh := make(chan *packet, 1)
			defer close(packetCh)

			go func() {
				defer srv.Close()

				packet := new(packet)
				// NOTE: receive packet
				if err := packet.decode(srv); err != nil {
					errCh <- err
					packetCh <- nil
					return
				}

				errCh <- nil
				packetCh <- packet
			}()

			// NOTE: send packet
			cltErr := tt.packet.encode(clt)

			if tt.clientErr == nil {
				assert.NoError(t, cltErr)
			} else {
				assert.Error(t, cltErr)
				assert.Equal(t, tt.clientErr, cltErr)
			}

			srvErr := <-errCh
			srvPacket := <-packetCh
			if tt.serverErr == nil {
				assert.NoError(t, srvErr)
				assert.Equal(t, tt.packet, srvPacket)
			} else {
				assert.Error(t, srvErr)
			}
		})
	}
}
