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
	"errors"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	mockAdderss  = ":25575"
	mockPassword = "minecraft"
	mockTimeout  = 100 * time.Millisecond
)

func mockServer(address string, timeout time.Duration) (net.Listener, error) {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return nil, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, err
	}

	if err := l.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, err
	}

	return l, nil
}

func pipe() (Rcon, Rcon) {
	srv, clt := net.Pipe()
	return &rcon{srv}, &rcon{clt}
}

func TestDialTimeout(t *testing.T) {
	cases := []struct {
		name      string
		addr      string
		password  string
		clientErr error
		serverErr error
	}{
		{
			name:      "positive case",
			addr:      "localhost:25575",
			password:  "minecraft",
			clientErr: nil,
			serverErr: nil,
		},
		{
			name:      "negative case: invalid addr",
			addr:      "192.0.2.100:25575",
			password:  "minecraft",
			clientErr: errors.New(""), // TODO
			serverErr: errors.New("timeout"),
		},
		{
			name:      "negative case: invalid port",
			addr:      "localhost:25565",
			password:  "minecraft",
			clientErr: errors.New(""), // TODO
			serverErr: errors.New("timeout"),
		},
		{
			name:      "negative case: invalid password",
			addr:      "localhost:25575",
			password:  "tfarcenim",
			clientErr: errors.New("unauthorized"),
			serverErr: nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			errCh := make(chan error, 1)
			defer close(errCh)

			go func() {
				l, err := mockServer(mockAdderss, mockTimeout)
				if err != nil {
					errCh <- err
					return
				}
				defer l.Close()

				conn, err := l.Accept()
				if err != nil {
					if err.Error() == "accept tcp [::]:25575: i/o timeout" {
						errCh <- errors.New("timeout")
					} else {
						errCh <- err
					}

					return
				}
				defer conn.Close()

				req := new(packet)
				if err := req.decode(conn); err != nil {
					errCh <- err
					return
				}

				var res *packet
				if string(req.payload) == mockPassword {
					res = newPacket(req.requestId, authResponseType, []byte{})
				} else {
					res = newPacket(unauthorizedRequestID, authResponseType, []byte{})
				}
				if err := res.encode(conn); err != nil {
					errCh <- err
					return
				}

				errCh <- nil
			}()

			conn, cltErr := DialTimeout(tt.addr, tt.password, mockTimeout)

			if tt.clientErr == nil {
				assert.NoError(t, cltErr)
				assert.NotNil(t, conn)
			} else {
				assert.Error(t, cltErr)
				assert.Nil(t, conn)
			}

			srvErr := <-errCh
			if tt.serverErr == nil {
				assert.NoError(t, srvErr)
			} else {
				assert.Error(t, srvErr)
				assert.Equal(t, tt.serverErr, srvErr)
			}
		})
	}
}

func Test_rcon_auth(t *testing.T) {
	cases := []struct {
		name      string
		password  string
		clientErr error
		serverErr error
	}{
		{
			name:      "positive case",
			password:  "minecraft",
			clientErr: nil,
			serverErr: nil,
		},
		{
			name:      "negative case",
			password:  "tfarcenim",
			clientErr: errors.New("unauthorized"),
			serverErr: nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			srv, clt := pipe()
			defer clt.Close()

			errCh := make(chan error, 1)
			defer close(errCh)

			go func() {
				defer srv.Close()

				req := new(packet)
				if err := req.decode(srv); err != nil {
					errCh <- err
					return
				}

				var res *packet
				if string(req.payload) == mockPassword {
					res = newPacket(req.requestId, authResponseType, []byte{})
				} else {
					res = newPacket(unauthorizedRequestID, authResponseType, []byte{})
				}
				if err := res.encode(srv); err != nil {
					errCh <- err
				}

				errCh <- nil
			}()

			cltErr := clt.auth(tt.password)

			if tt.clientErr == nil {
				assert.NoError(t, cltErr)
			} else {
				assert.Error(t, cltErr)
				assert.Equal(t, tt.clientErr, cltErr)
			}

			srvErr := <-errCh
			if tt.serverErr == nil {
				assert.NoError(t, srvErr)
			} else {
				assert.Error(t, srvErr)
			}
		})
	}
}

func Test_rcon_Command(t *testing.T) {
	cases := []struct {
		name      string
		command   string
		responses []packet
		expected  string
		clientErr error
		serverErr error
	}{
		{
			name:    "positive case: non-fragment response",
			command: "request",
			responses: []packet{
				{requestId: 123456, packetType: commandResponseType, payload: []byte("response")},
				{requestId: 123456, packetType: commandResponseType, payload: []byte("Unknown request 64")},
			},
			expected:  "response",
			clientErr: nil,
			serverErr: nil,
		},
		{
			name:    "positive case: fragment response",
			command: "request",
			responses: []packet{
				{requestId: 123456, packetType: commandResponseType, payload: []byte(strings.Repeat("response", 4096/len("response")))},
				{requestId: 123456, packetType: commandResponseType, payload: []byte(strings.Repeat("response", 4096/len("response")))},
				{requestId: 123456, packetType: commandResponseType, payload: []byte(strings.Repeat("response", 1808/len("response")))},
				{requestId: 123456, packetType: commandResponseType, payload: []byte("Unknown request 64")},
			},
			expected:  strings.Repeat("response", 10000/len("response")),
			clientErr: nil,
			serverErr: nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			srv, clt := pipe()
			defer clt.Close()

			errCh := make(chan error, 1)
			defer close(errCh)

			go func() {
				defer srv.Close()

				req := new(packet)
				if err := req.decode(srv); err != nil {
					errCh <- err
					return
				}

				if err := tt.responses[0].encode(srv); err != nil {
					errCh <- err
				}

				dummy := new(packet)
				if err := dummy.decode(srv); err != nil {
					errCh <- err
					return
				}

				for _, res := range tt.responses[1:] {
					if err := res.encode(srv); err != nil {
						errCh <- err
					}
				}

				errCh <- nil
			}()

			actual, cltErr := clt.Command(tt.command)

			if tt.clientErr == nil {
				assert.NoError(t, cltErr)
				assert.Equal(t, tt.expected, actual)
			} else {
				assert.Error(t, cltErr)
				assert.Equal(t, tt.clientErr, cltErr)
			}

			srvErr := <-errCh
			if tt.serverErr == nil {
				assert.NoError(t, srvErr)
			} else {
				assert.Error(t, srvErr)
			}
		})
	}
}

func Test_rcon_request(t *testing.T) {
	cases := []struct {
		name      string
		id        int32
		typ       packetType
		payload   []byte
		responses []packet
		expected  *packet
		clientErr error
		serverErr error
	}{
		{
			name:    "positive case: Auth Request",
			id:      123456,
			typ:     authRequestType,
			payload: []byte("minecraft"),
			responses: []packet{
				{requestId: 123456, packetType: authResponseType, payload: []byte{}},
			},
			expected:  &packet{requestId: 123456, packetType: authResponseType, payload: []byte{}},
			clientErr: nil,
			serverErr: nil,
		},
		{
			name:    "positive case: Command Request - non-fragment response",
			id:      123456,
			typ:     commandRequestType,
			payload: []byte("request"),
			responses: []packet{
				{requestId: 123456, packetType: commandResponseType, payload: []byte("response")},
				{requestId: 123456, packetType: commandResponseType, payload: []byte("Unknown request 64")},
			},
			expected:  &packet{requestId: 123456, packetType: commandResponseType, payload: []byte("response")},
			clientErr: nil,
			serverErr: nil,
		},
		{
			name:    "positive case: Command Request - fragment response",
			id:      123456,
			typ:     commandRequestType,
			payload: []byte("request"),
			responses: []packet{
				{requestId: 123456, packetType: commandResponseType, payload: []byte(strings.Repeat("response", 4096/len("response")))},
				{requestId: 123456, packetType: commandResponseType, payload: []byte(strings.Repeat("response", 4096/len("response")))},
				{requestId: 123456, packetType: commandResponseType, payload: []byte(strings.Repeat("response", 1808/len("response")))},
				{requestId: 123456, packetType: commandResponseType, payload: []byte("Unknown request 64")},
			},
			expected:  &packet{requestId: 123456, packetType: commandResponseType, payload: []byte(strings.Repeat("response", 10000/len("response")))},
			clientErr: nil,
			serverErr: nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			srv, clt := pipe()
			defer clt.Close()

			errCh := make(chan error, 1)
			defer close(errCh)

			go func() {
				defer srv.Close()

				req := new(packet)
				if err := req.decode(srv); err != nil {
					errCh <- err
					return
				}

				if err := tt.responses[0].encode(srv); err != nil {
					errCh <- err
				}

				if tt.typ == authRequestType {
					errCh <- nil
					return
				}

				dummy := new(packet)
				if err := dummy.decode(srv); err != nil {
					errCh <- err
					return
				}

				for _, res := range tt.responses[1:] {
					if err := res.encode(srv); err != nil {
						errCh <- err
					}
				}

				errCh <- nil
			}()

			actual, cltErr := clt.request(tt.id, tt.typ, tt.payload)

			if tt.clientErr == nil {
				assert.NoError(t, cltErr)
				assert.Equal(t, tt.expected, actual)
			} else {
				assert.Error(t, cltErr)
				assert.Equal(t, tt.clientErr, cltErr)
			}

			srvErr := <-errCh
			if tt.serverErr == nil {
				assert.NoError(t, srvErr)
			} else {
				assert.Error(t, srvErr)
			}
		})
	}
}
