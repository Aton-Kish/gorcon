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
Package rcon extends net.Conn for RCON.


Example


A simple RCON client for Minecraft:

	c, err := rcon.Dial("localhost:25575", "password")
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer c.Close()

	// Exec any commands
	res, err := c.Command("/seed")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	fmt.Println(res) // Seed: [...]
*/
package rcon

import (
	"math/rand"
	"net"
	"time"

	"github.com/pkg/errors"

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
		return Rcon{}, errors.WithStack(err)
	}

	c := Rcon{conn}
	if err := c.auth(password); err != nil {
		defer c.Close()
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

// For details about commands, see the wiki https://minecraft.fandom.com/wiki/Commands.
func (c *Rcon) Command(command string) (string, error) {
	id := rand.Int31()

	p := []byte(command)
	if len(p) > requestPayloadMaxLength {
		return "", errors.Errorf("request payload is over %d", requestPayloadMaxLength)
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
	if err := c.writePacket(req); err != nil {
		return nil, err
	}

	res, err := c.readPacket()
	if err != nil {
		return nil, err
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
	if err := c.writePacket(req); err != nil {
		return nil, err
	}

	for {
		pac, err := c.readPacket()
		if err != nil {
			return nil, err
		}

		if pac.RequestID != id {
			continue
		}

		body := string(pac.Payload)
		if body == "Unknown request 64" {
			// Termination
			break
		}

		res.Length += int32(len(pac.Payload))
		res.Payload = append(res.Payload, pac.Payload...)
	}

	return res, nil
}

func (c *Rcon) readPacket() (*packet.Packet, error) {
	hraw := make([]byte, 0, 4+4+4)
	for len(hraw) < 4+4+4 {
		buf := make([]byte, 4+4+4-len(hraw))
		n, err := c.Read(buf)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		hraw = append(hraw, buf[:n]...)
	}

	h, err := packet.ParseHeader(hraw)
	if err != nil {
		return nil, err
	}

	praw := make([]byte, 0, h.Length-(4+4))
	for len(praw) < int(h.Length)-(4+4) {
		buf := make([]byte, int(h.Length)-(4+4+len(praw)))
		n, err := c.Read(buf)
		if err != nil {
			return nil, errors.WithStack(err)
		}

		praw = append(praw, buf[:n]...)
	}

	pac, err := packet.PackWithHeader(praw, h)
	if err != nil {
		return nil, err
	}

	return pac, nil
}

func (c *Rcon) writePacket(pac *packet.Packet) error {
	raw, err := packet.Unpack(pac)
	if err != nil {
		return err
	}

	if _, err := c.Write(raw); err != nil {
		return errors.WithStack(err)
	}

	return nil
}
