package rcon

import (
	"net"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Aton-Kish/gorcon/packet"
	"github.com/Aton-Kish/gorcon/types"
	"github.com/caarlos0/env/v6"
	"github.com/stretchr/testify/assert"
)

var cfg = new(Config)

type Config struct {
	Addr     string `env:"RCON_ADDRESS" envDefault:"minecraft:25575"`
	Password string `env:"RCON_PASSWORD" envDefault:"minecraft"`
}

func TestMain(m *testing.M) {
	if err := env.Parse(cfg); err != nil {
		panic(err)
	}

	exit := m.Run()
	os.Exit(exit)
}

func TestDial(t *testing.T) {
	cases := []struct {
		name     string
		addr     string
		password string
		hasError bool
	}{
		{
			name:     "Valid Case",
			addr:     cfg.Addr,
			password: cfg.Password,
			hasError: false,
		},
		{
			name:     "Invalid Case: missing address and password",
			addr:     "",
			password: "",
			hasError: true,
		},
		{
			name:     "Invalid Case: missing address",
			addr:     "",
			password: cfg.Password,
			hasError: true,
		},
		{
			name:     "Invalid Case: missing port",
			addr:     strings.Split(cfg.Addr, ":")[0],
			password: cfg.Password,
			hasError: true,
		},
		{
			name:     "Invalid Case: invalid address",
			addr:     "dummy:25575",
			password: cfg.Password,
			hasError: true,
		},
		{
			name:     "Invalid Case: invalid port",
			addr:     strings.Split(cfg.Addr, ":")[0] + ":80",
			password: cfg.Password,
			hasError: true,
		},
		{
			name:     "Invalid Case: missing password",
			addr:     cfg.Addr,
			password: "",
			hasError: true,
		},
		{
			name:     "Invalid Case: invalid password",
			addr:     cfg.Addr,
			password: "dummy",
			hasError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conn, err := Dial(c.addr, c.password)
			if c.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				defer conn.Close()
			}
		})
	}
}

func TestDialTimeout(t *testing.T) {
	cases := []struct {
		name     string
		addr     string
		password string
		timeout  time.Duration
		hasError bool
	}{
		{
			name:     "Valid Case",
			addr:     cfg.Addr,
			password: cfg.Password,
			timeout:  time.Duration(1) * time.Second,
			hasError: false,
		},
		{
			name:     "Invalid Case: too short timeout",
			addr:     cfg.Addr,
			password: cfg.Password,
			timeout:  time.Duration(1) * time.Nanosecond,
			hasError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			conn, err := DialTimeout(c.addr, c.password, c.timeout)
			if c.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				defer conn.Close()
			}
		})
	}
}

func TestCommand(t *testing.T) {
	conn, err := Dial(cfg.Addr, cfg.Password)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	cases := []struct {
		name     string
		command  string
		contains string
	}{
		{
			name:     "Valid Case: /seed",
			command:  "/seed",
			contains: "Seed: ",
		},
		{
			name:     "Valid Case: /time query day",
			command:  "/time query day",
			contains: "The time is ",
		},
		{
			name:     "Invalid Case: invalid command",
			command:  "/",
			contains: "Unknown or incomplete command, see below for error",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			res, err := conn.Command(c.command)
			assert.NoError(t, err)
			assert.Contains(t, res, c.contains)
		})
	}
}

func TestReadPackets(t *testing.T) {
	cases := []struct {
		name     string
		raw      []byte
		want     []*packet.Packet
		hasError bool
	}{
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
			want: []*packet.Packet{
				{Header: packet.Header{Length: 10, RequestID: 789012, Type: types.AuthResponse}, Payload: []byte("")},
			},
			hasError: false,
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
			want: []*packet.Packet{
				{Header: packet.Header{Length: 18, RequestID: 901234, Type: types.CommandResponse}, Payload: []byte("response")},
			},
			hasError: false,
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
			want: []*packet.Packet{
				{Header: packet.Header{Length: 28, RequestID: 123456, Type: types.CommandResponse}, Payload: []byte("Unknown request 64")},
			},
			hasError: false,
		},
		{
			name: "Valid Case: CommandResponse + UnknownResponse",
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
			want: []*packet.Packet{
				{Header: packet.Header{Length: 18, RequestID: 901234, Type: types.CommandResponse}, Payload: []byte("response")},
				{Header: packet.Header{Length: 28, RequestID: 123456, Type: types.CommandResponse}, Payload: []byte("Unknown request 64")},
			},
			hasError: false,
		},
		{
			name: "Invalid Case: invalid format",
			raw: []byte{
				0x00,
			},
			want:     nil,
			hasError: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, clt := net.Pipe()
			conn := Rcon{clt}
			defer conn.Close()

			errc := make(chan error)
			go func() {
				// server mock
				defer srv.Close()

				if _, err := srv.Write(c.raw); err != nil {
					errc <- err
				}

				close(errc)
			}()

			pacs, err := conn.readPackets()
			if c.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.EqualValues(t, c.want, pacs)

			if err := <-errc; err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestWritePackets(t *testing.T) {
	cases := []struct {
		name    string
		packets []*packet.Packet
		want    []byte
	}{
		{
			name: "Valid Case: AuthRequest",
			packets: []*packet.Packet{
				{Header: packet.Header{Length: 14, RequestID: 123456, Type: types.AuthRequest}, Payload: []byte("auth")},
			},
			want: []byte{
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
		},
		{
			name: "Valid Case: CommandRequest",
			packets: []*packet.Packet{
				{Header: packet.Header{Length: 17, RequestID: 345678, Type: types.CommandRequest}, Payload: []byte("command")},
			},
			want: []byte{
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
		},
		{
			name: "Valid Case: DummyRequest",
			packets: []*packet.Packet{
				{Header: packet.Header{Length: 23, RequestID: 567890, Type: types.DummyRequest}, Payload: []byte("dummy request")},
			},
			want: []byte{
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
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			srv, clt := net.Pipe()
			conn := Rcon{clt}
			defer conn.Close()

			errc := make(chan error)
			rawc := make(chan []byte)
			go func() {
				// server mock
				defer srv.Close()

				raw := make([]byte, 4+4+4+ResponsePayloadMaxLength+1+1)

				n, err := srv.Read(raw)
				if err != nil {
					errc <- err
				}
				close(errc)

				rawc <- raw[:n]
				close(rawc)
			}()

			err := conn.writePackets(c.packets...)
			assert.NoError(t, err)

			if err := <-errc; err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, c.want, <-rawc)
		})
	}
}
