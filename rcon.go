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
	BadAuthRequestID         = -1
	RequestPayloadMaxLength  = 1446
	ResponsePayloadMaxLength = 4096
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
	res, err := c.getSingleResponse(id, types.AuthRequest, p)
	if err != nil {
		return err
	}

	if res.RequestID == BadAuthRequestID {
		return errors.New("bad auth")
	}

	return nil
}

func (c *Rcon) Command(command string) (string, error) {
	id := rand.Int31()

	p := []byte(command)
	if len(p) > RequestPayloadMaxLength {
		return "", fmt.Errorf("request payload is over %d", RequestPayloadMaxLength)
	}

	res, err := c.getMultipleResponse(id, types.CommandRequest, p)
	if err != nil {
		return "", err
	}

	return string(res.Payload), nil
}

func (c *Rcon) getSingleResponse(id int32, typ types.Packet, payload []byte) (*packet.Packet, error) {
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
			res.Payload = append(res.Payload, pac.Payload...)
		}
	}

	return res, nil
}

func (c *Rcon) getMultipleResponse(id int32, typ types.Packet, payload []byte) (*packet.Packet, error) {
	res, err := c.getSingleResponse(id, typ, payload)
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

			res.Payload = append(res.Payload, pac.Payload...)
		}
	}
}

func (c *Rcon) readPackets() ([]*packet.Packet, error) {
	raw := make([]byte, 4+4+4+ResponsePayloadMaxLength+2)

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
