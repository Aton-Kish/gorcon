package main

import (
	"errors"
	"fmt"
	"log"
	"regexp"

	"github.com/Aton-Kish/gonbt"
	rcon "github.com/Aton-Kish/gorcon"
	"github.com/caarlos0/env/v6"
)

var (
	listPattern = regexp.MustCompile(`([a-zA-Z0-9_]{3,16}) \(([0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12})\)`)
	dataPattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,16} has the following entity data: (.*)$`)
)

type Config struct {
	Addr     string `env:"RCON_ADDRESS" envDefault:"minecraft:25575"`
	Password string `env:"RCON_PASSWORD" envDefault:"minecraft"`
}

func main() {
	cfg := new(Config)
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("%+v", err)
	}

	c, err := rcon.Dial(cfg.Addr, cfg.Password)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer c.Close()

	players, err := listCommand(&c)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	for uuid, name := range players {
		raw, err := dataCommand(&c, uuid)
		if err != nil {
			log.Fatalf("%+v", err)
		}

		fmt.Printf("Player: %s (%s)\n", name, uuid)
		fmt.Printf("NBT   : %s\n", raw)
	}
}

func listCommand(rcon *rcon.Rcon) (map[string]string, error) {
	res, err := rcon.Command("/list uuids")
	if err != nil {
		return nil, err
	}

	gs := listPattern.FindAllStringSubmatch(res, -1)
	players := make(map[string]string)
	for _, g := range gs {
		players[g[2]] = g[1]
	}

	return players, nil
}

func dataCommand(rcon *rcon.Rcon, uuid string) ([]byte, error) {
	res, err := rcon.Command("/data get entity " + uuid)
	if err != nil {
		return nil, err
	}

	g := dataPattern.FindStringSubmatch(res)
	if len(g) < 2 {
		return nil, errors.New("invalid data format")
	}

	// parse SNBT to NBT
	bm := gonbt.NewSnbtTokenBitmaps(g[1])
	bm.SetTokenBitmaps()
	bm.SetMaskBitmaps()
	bm = bm.Compact()

	tag := new(gonbt.Tag)
	if err := gonbt.Parse(&bm, tag); err != nil {
		return nil, err
	}

	// parse NBT to JSON
	j := gonbt.CompactJson(tag)

	return []byte(j), nil
}
