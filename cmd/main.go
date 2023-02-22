package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bendersilver/joura"
)

// journalctl --user -n 10 -f -o cat

func main() {
	var file string
	flag.StringVar(&file, "c", "/etc/joura/joura.conf", "set configuration file (default: /etc/joura/joura.conf)")

	j, err := joura.New(file)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	j.Start()
}
