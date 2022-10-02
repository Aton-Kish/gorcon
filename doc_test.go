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

package rcon_test

import (
	"fmt"
	"log"
	"regexp"
	"time"

	nbt "github.com/Aton-Kish/gonbt"
	rcon "github.com/Aton-Kish/gorcon"
)

func ExampleDial() {
	conn, err := rcon.Dial("localhost:25575", "minecraft")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	res, err := conn.Command("/seed")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
}

func ExampleDialTimeout() {
	conn, err := rcon.DialTimeout("localhost:25575", "minecraft", 500*time.Millisecond)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	res, err := conn.Command("/seed")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
}

func Example_command() {
	conn, err := rcon.Dial("localhost:25575", "minecraft")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// NOTE: `/player` is a carpet mod command
	res, err := conn.Command("/player jeb_ spawn")
	if err != nil {
		log.Fatal(err)
	}

	res, err = conn.Command("/give jeb_ minecraft:dirt 1")
	if err != nil {
		log.Fatal(err)
	}

	// NOTE: `/player` is a carpet mod command
	res, err = conn.Command("/player jeb_ kill")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(res)
}

func Example_nbtData() {
	conn, err := rcon.Dial("localhost:25575", "minecraft")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	// NOTE: uuid `853c80ef-3c37-49fd-aa49-938b674adae6` is jeb_
	res, err := conn.Command("/data get entity 853c80ef-3c37-49fd-aa49-938b674adae6")
	if err != nil {
		log.Fatal(err)
	}

	re := regexp.MustCompile(`^[a-zA-Z0-9_]{3,16} has the following entity data: (.*)$`)
	g := re.FindStringSubmatch(res)
	if len(g) < 2 {
		log.Fatal("invalid data")
	}

	// NOTE: parse NBT data
	snbt := g[1]
	dat, err := nbt.Parse(snbt)
	if err != nil {
		log.Fatal(err)
	}

	json := nbt.Json(dat)
	fmt.Println(json)
}
