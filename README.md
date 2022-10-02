# Go RCON

A Minecraft RCON client

## Getting Started

Use go get to install the library

```go
go get github.com/Aton-Kish/gonbt
```

Import in your application

```go
import (
	rcon "github.com/Aton-Kish/gorcon"
)
```

## Usage

```go
package main

import (
	"fmt"
	"log"

	rcon "github.com/Aton-Kish/gorcon"
)

func main() {
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
```

## License

This library is licensed under the MIT License, see [LICENSE](./LICENSE).
