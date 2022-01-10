package main

import (
	"fmt"
	"log"

	rcon "github.com/Aton-Kish/gorcon"
	"github.com/caarlos0/env/v6"
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

	res, err := c.Command("/seed")
	if err != nil {
		log.Fatalf("%+v", err)
	}

	fmt.Println(res) // Seed: [...]
}
