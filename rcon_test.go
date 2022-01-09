package rcon

import (
	"os"
	"strings"
	"testing"
	"time"

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
