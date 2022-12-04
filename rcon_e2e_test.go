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
	"fmt"
	"os"
	"testing"

	"github.com/Aton-Kish/gonbt/slices"
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
			expectedErr: &rcon.RCONError{},
		},
		{
			name:        "negative case: missing port",
			address:     "localhost",
			password:    cfg.Password,
			expectedErr: &rcon.RCONError{},
		},
		{
			name:        "negative case: invalid port",
			address:     "localhost:50000",
			password:    cfg.Password,
			expectedErr: &rcon.RCONError{},
		},
		{
			name:        "negative case: missing password",
			address:     fmt.Sprintf("localhost:%s", cfg.Port),
			password:    "",
			expectedErr: &rcon.RCONError{},
		},
		{
			name:        "negative case: invalid password",
			address:     fmt.Sprintf("localhost:%s", cfg.Port),
			password:    "tfarcenim",
			expectedErr: &rcon.RCONError{},
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			conn, err := rcon.Dial(tt.address, tt.password)
			if tt.expectedErr == nil {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.IsType(t, tt.expectedErr, err)
			}

			if conn != nil {
				conn.Close()
			}
		})
	}
}

func TestRCON_Command(t *testing.T) {
	type Case struct {
		name        string
		command     string
		expected    string
		expectedErr error
	}

	items := []string{
		"minecraft:dirt",
		"minecraft:oak_log",
		"minecraft:birch_log",
		"minecraft:spruce_log",
		"minecraft:jungle_log",
		"minecraft:acacia_log",
		"minecraft:dark_oak_log",
		"minecraft:white_wool",
		"minecraft:cobblestone",
		"minecraft:stone",
		"minecraft:granite",
		"minecraft:diorite",
		"minecraft:andesite",
		"minecraft:terracotta",
		"minecraft:sand",
		"minecraft:glass",
		"minecraft:emerald",
		"minecraft:redstone",
		"minecraft:lapis_lazuli",
		"minecraft:copper_ingot",
		"minecraft:iron_ingot",
		"minecraft:gold_ingot",
		"minecraft:diamond",
		"minecraft:netherrack",
		"minecraft:quartz",
		"minecraft:netherite_ingot",
		"minecraft:end_stone",
		"minecraft:purpur_block",
		"minecraft:shulker_box",
	}

	giveCases := make([]Case, 0, len(items))

	for _, item := range items {
		giveCases = append(giveCases, Case{
			name:        fmt.Sprintf("positive case: /give jeb_ %s 1", item),
			command:     fmt.Sprintf("/give jeb_ %s 1", item),
			expected:    `^Gave 1 \[[a-zA-Z ]+\] to jeb_$`,
			expectedErr: nil,
		})
	}

	cases := slices.Concat(
		[]Case{
			{
				name:        "positive case: /",
				command:     "/",
				expected:    `^Unknown or incomplete command, see below for error<--\[HERE\]$`,
				expectedErr: nil,
			},
			{
				name:        "positive case: /seed",
				command:     "/seed",
				expected:    `^Seed: \[-?\d+\]$`,
				expectedErr: nil,
			},
			{
				name:        "positive case: /time query day",
				command:     "/time query day",
				expected:    `^The time is \d+$`,
				expectedErr: nil,
			},
			{
				name:        "positive case: /list uuids - no player exists",
				command:     "/list uuids",
				expected:    `^There are 0 of a max of \d+ players online: $`,
				expectedErr: nil,
			},
			{
				name:        "positive case: /player jeb_ spawn",
				command:     "/player jeb_ spawn",
				expected:    `^$`,
				expectedErr: nil,
			},
		},
		giveCases,
		[]Case{
			{
				name:        "positive case: /list uuids - jeb_ exists",
				command:     "/list uuids",
				expected:    `^There are 1 of a max of \d+ players online: jeb_ \(853c80ef-3c37-49fd-aa49-938b674adae6\)$`,
				expectedErr: nil,
			},
			{
				name:        "positive case: /data get entity 853c80ef-3c37-49fd-aa49-938b674adae6",
				command:     "/data get entity 853c80ef-3c37-49fd-aa49-938b674adae6",
				expected:    `^jeb_ has the following entity data: .*$`,
				expectedErr: nil,
			},
			{
				name:        "positive case: /player jeb_ kill",
				command:     "/player jeb_ kill",
				expected:    `^$`,
				expectedErr: nil,
			},
			{
				name:        "positive case: /kill @e[type=item]",
				command:     "/kill @e[type=item]",
				expected:    `^Killed .*$`,
				expectedErr: nil,
			},
		},
	)

	addr := fmt.Sprintf("localhost:%s", cfg.Port)
	conn, err := rcon.Dial(addr, cfg.Password)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			actual, err := conn.Command(tt.command)
			if tt.expectedErr == nil {
				assert.NoError(t, err)
				assert.Regexp(t, tt.expected, actual)
			} else {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			}
		})
	}
}
