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

//go:build e2e

package rcon_test

import (
	"errors"
	"fmt"
	"os"
	"testing"

	rcon "github.com/Aton-Kish/gorcon"
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

const (
	dotenvPath = "server/.env"
)

var cfg = new(config)

type config struct {
	Port     string `env:"RCON_PORT"`
	Password string `env:"RCON_PASSWORD"`
}

func TestMain(m *testing.M) {
	if err := godotenv.Load(dotenvPath); err != nil {
		os.Exit(1)
	}

	if err := env.Parse(cfg); err != nil {
		os.Exit(1)
	}

	code := m.Run()
	os.Exit(code)
}

func TestDial(t *testing.T) {
	cases := []struct {
		name        string
		address     string
		password    string
		expectedErr error
	}{
		{
			name:        "positive case",
			address:     fmt.Sprintf("localhost:%s", cfg.Port),
			password:    cfg.Password,
			expectedErr: nil,
		},
		{
			name:        "negative case: missing address",
			address:     "",
			password:    cfg.Password,
			expectedErr: errors.New(""), // TODO
		},
		{
			name:        "negative case: missing port",
			address:     "localhost",
			password:    cfg.Password,
			expectedErr: errors.New(""), // TODO
		},
		{
			name:        "negative case: invalid port",
			address:     "localhost:50000",
			password:    cfg.Password,
			expectedErr: errors.New(""), // TODO
		},
		{
			name:        "negative case: missing password",
			address:     fmt.Sprintf("localhost:%s", cfg.Port),
			password:    "",
			expectedErr: errors.New("unauthorized"),
		},
		{
			name:        "negative case: invalid password",
			address:     fmt.Sprintf("localhost:%s", cfg.Port),
			password:    "tfarcenim",
			expectedErr: errors.New("unauthorized"),
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := rcon.Dial(tt.address, tt.password)
			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				// assert.Equal(t, tt.expectedErr, err)
			}

			if conn != nil {
				conn.Close()
			}
		})
	}
}

func TestRcon_Command(t *testing.T) {
	addr := fmt.Sprintf("localhost:%s", cfg.Port)
	conn, err := rcon.Dial(addr, cfg.Password)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	cases := []struct {
		name        string
		command     string
		contains    string
		expectedErr error
	}{
		{
			name:        "positive case: /seed",
			command:     "/seed",
			contains:    "Seed: ",
			expectedErr: nil,
		},
		{
			name:        "positive case: /time query day",
			command:     "/time query day",
			contains:    "The time is ",
			expectedErr: nil,
		},
		{
			name:        "positive case: /",
			command:     "/",
			contains:    "Unknown or incomplete command, see below for error<--[HERE]",
			expectedErr: nil,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := conn.Command(tt.command)
			if tt.expectedErr == nil {
				assert.NoError(t, err)
				assert.Contains(t, actual, tt.contains)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
