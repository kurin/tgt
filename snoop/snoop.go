package main

import (
	"fmt"
	"io"
	"log"
	"os"

	"github.com/kurin/tgt/packet"
)

func main() {
	for {
		m, err := packet.Next(os.Stdin)

		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(m)
	}
}
