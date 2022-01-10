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

package rcon

import (
	"errors"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/Aton-Kish/gorcon/packet"
	"github.com/Aton-Kish/gorcon/types"
)

const (
	badAuthRequestID         = -1
	requestPayloadMaxLength  = 1446
	responsePayloadMaxLength = 4096
)

type Rcon struct {
	net.Conn
}

func Dial(addr string, password string) (Rcon, error) {
	c, err := DialTimeout(addr, password, 0)
	if err != nil {
		return Rcon{}, err
	}

	return c, nil
}

func DialTimeout(addr string, password string, timeout time.Duration) (Rcon, error) {
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return Rcon{}, err
	}

	c := Rcon{conn}
	if err := c.auth(password); err != nil {
		return Rcon{}, err
	}

	return c, nil
}

func (c *Rcon) auth(password string) error {
	id := rand.Int31()
	p := []byte(password)
	res, err := c.request(id, types.AuthRequest, p)
	if err != nil {
		return err
	}

	if res.RequestID != id || res.RequestID == badAuthRequestID {
		return errors.New("bad auth")
	}

	return nil
}

func (c *Rcon) Command(command string) (string, error) {
	id := rand.Int31()

	p := []byte(command)
	if len(p) > requestPayloadMaxLength {
		return "", fmt.Errorf("request payload is over %d", requestPayloadMaxLength)
	}

	res, err := c.requestWithEndConfirmation(id, types.CommandRequest, p)
	if err != nil {
		return "", err
	}

	return string(res.Payload), nil
}

func (c *Rcon) request(id int32, typ types.Packet, payload []byte) (*packet.Packet, error) {
	var res *packet.Packet

	req := packet.NewPacket(id, typ, payload)
	if err := c.writePackets(req); err != nil {
		return nil, err
	}

	pacs, err := c.readPackets()
	if err != nil {
		return nil, err
	}

	for _, pac := range pacs {
		if res == nil {
			res = pac
		} else {
			res.Length += int32(len(pac.Payload))
			res.Payload = append(res.Payload, pac.Payload...)
		}
	}

	return res, nil
}

func (c *Rcon) requestWithEndConfirmation(id int32, typ types.Packet, payload []byte) (*packet.Packet, error) {
	res, err := c.request(id, typ, payload)
	if err != nil {
		return nil, err
	}

	// Dummy Request
	req := packet.NewPacket(id, types.DummyRequest, []byte{})
	if err := c.writePackets(req); err != nil {
		return nil, err
	}

	for {
		pacs, err := c.readPackets()
		if err != nil {
			return nil, err
		}

		for _, pac := range pacs {
			if pac.RequestID != id {
				continue
			}

			body := string(pac.Payload)
			if body == "Unknown request 64" {
				// Termination
				return res, nil
			}

			res.Length += int32(len(pac.Payload))
			res.Payload = append(res.Payload, pac.Payload...)
		}
	}
}

func (c *Rcon) readPackets() ([]*packet.Packet, error) {
	raw := make([]byte, 4+4+4+responsePayloadMaxLength+1+1)

	n, err := c.Read(raw)
	if err != nil {
		return nil, err
	}

	l := 0
	pacs := make([]*packet.Packet, 0, 1)
	for l < n {
		pac, err := packet.Pack(raw[l:n])
		if err != nil {
			return nil, err
		}

		pacs = append(pacs, pac)

		l += 4 + int(pac.Length)
	}

	return pacs, nil
}

func (c *Rcon) writePackets(pacs ...*packet.Packet) error {
	for _, pac := range pacs {
		raw, err := packet.Unpack(pac)
		if err != nil {
			return err
		}

		if _, err := c.Write(raw); err != nil {
			return err
		}
	}

	return nil
}
