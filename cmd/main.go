package main

import (
	"fmt"
	"os"

	"github.com/bendersilver/joura"
)

// journalctl --user -n 10 -f -o cat

func main() {
	j, err := joura.New()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	j.Start()
}
