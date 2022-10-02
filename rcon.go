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
	"math/rand"
	"net"
	"time"
)

const (
	unauthorizedRequestID = -1
)

type Rcon interface {
	net.Conn

	auth(password string) error
	Command(command string) (string, error)
	request(id int32, typ packetType, payload []byte) (*packet, error)
}

type rcon struct {
	net.Conn
}

func Dial(addr string, password string) (Rcon, error) {
	c, err := DialTimeout(addr, password, 0)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func DialTimeout(addr string, password string, timeout time.Duration) (Rcon, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return nil, err
	}

	c := &rcon{conn}
	if err := c.auth(password); err != nil {
		defer c.Close()
		return nil, err
	}

	return c, nil
}

func (c *rcon) auth(password string) error {
	id := rand.Int31()
	res, err := c.request(id, authRequestType, []byte(password))
	if err != nil {
		return err
	}

	if res.requestId != id || res.requestId == unauthorizedRequestID {
		return errors.New("unauthorized")
	}

	return nil
}

func (c *rcon) Command(command string) (string, error) {
	id := rand.Int31()
	res, err := c.request(id, commandRequestType, []byte(command))
	if err != nil {
		return "", err
	}

	return string(res.payload), nil
}

func (c *rcon) request(id int32, typ packetType, payload []byte) (*packet, error) {
	req := newPacket(id, typ, payload)
	if err := req.encode(c); err != nil {
		return nil, err
	}

	res := new(packet)
	if err := res.decode(c); err != nil {
		return nil, err
	}

	if typ == authRequestType {
		return res, nil
	}

	// NOTE: dummy request
	dummy := newPacket(id, dummyRequestType, []byte{})
	if err := dummy.encode(c); err != nil {
		return nil, err
	}

	for {
		more := new(packet)
		if err := more.decode(c); err != nil {
			return nil, err
		}

		if string(more.payload) == "Unknown request 64" {
			// NOTE: termination
			break
		}

		res.payload = append(res.payload, more.payload...)
	}

	return res, nil
}
