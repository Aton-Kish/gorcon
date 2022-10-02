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
