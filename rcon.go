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
	"math/rand"
	"net"
	"time"

	"github.com/rs/zerolog/log"
)

const (
	unauthorizedRequestID  = -1
	maxResponsePayloadSize = 4096
	maxResponseLength      = 4 + 4 + (maxResponsePayloadSize + 1) + 1
)

type Rcon interface {
	Read(b []byte) (int, error)
	Write(b []byte) (int, error)
	Close() error

	Command(command string) (string, error)
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
		err = &RconError{Op: "dial", Err: err}
		log.Error().Caller().Str("func", "DialTimeout").Err(err).Msg("")
		return nil, err
	}

	c := &rcon{conn}
	if err := c.auth(password); err != nil {
		defer c.Close()
		err = &RconError{Op: "dial", Err: err}
		log.Error().Caller().Str("func", "DialTimeout").Err(err).Msg("")
		return nil, err
	}

	log.Debug().Caller().Str("func", "DialTimeout").Msg("connection established")
	return c, nil
}

func pipe() (*rcon, *rcon) {
	srv, clt := net.Pipe()
	return &rcon{srv}, &rcon{clt}
}

func (c *rcon) auth(password string) error {
	id := rand.Int31()
	res, err := c.request(id, authRequestType, []byte(password))
	if err != nil {
		err = &RconError{Op: "auth", Err: err}
		log.Error().Caller().Str("func", "(*rcon).auth").Err(err).Msg("")
		return err
	}

	if res.requestId != id || res.requestId == unauthorizedRequestID {
		err = &RconError{Op: "auth"}
		log.Error().Caller().Str("func", "(*rcon).auth").Int32("requestId", res.requestId).Err(err).Msg("")
		return err
	}

	log.Debug().Caller().Str("func", "(*rcon).auth").Int32("requestId", res.requestId).Msg("")
	return nil
}

func (c *rcon) Command(command string) (string, error) {
	id := rand.Int31()
	res, err := c.request(id, commandRequestType, []byte(command))
	if err != nil {
		err = &RconError{Op: "command", Err: err}
		log.Error().Caller().Str("func", "(*rcon).Command").Err(err).Msg("")
		return "", err
	}

	payload := string(res.payload)

	log.Debug().Caller().Str("func", "(*rcon).Command").Int32("requestId", res.requestId).Str("response", payload).Msg("")
	return payload, nil
}

func (c *rcon) request(id int32, typ packetType, payload []byte) (*packet, error) {
	req := newPacket(id, typ, payload)
	if err := req.encode(c); err != nil {
		log.Error().Caller().Str("func", "(*rcon).request").Err(err).Msg("")
		return nil, err
	}

	res := new(packet)
	if err := res.decode(c); err != nil {
		log.Error().Caller().Str("func", "(*rcon).request").Err(err).Msg("")
		return nil, err
	}

	if res.length() < maxResponseLength {
		log.Debug().Caller().Str("func", "(*rcon).request").Str("packet", res.String()).Msg("")
		return res, nil
	}

	// NOTE: dummy request
	dummy := newPacket(id, dummyRequestType, []byte{})
	if err := dummy.encode(c); err != nil {
		log.Error().Caller().Str("func", "(*rcon).request").Err(err).Msg("")
		return nil, err
	}

	for {
		more := new(packet)
		if err := more.decode(c); err != nil {
			log.Error().Caller().Str("func", "(*rcon).request").Err(err).Msg("")
			return nil, err
		}

		if string(more.payload) == "Unknown request 64" {
			// NOTE: termination
			break
		}

		res.payload = append(res.payload, more.payload...)
	}

	log.Debug().Caller().Str("func", "(*rcon).request").Str("packet", res.String()).Msg("")
	return res, nil
}
