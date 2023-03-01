package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/bendersilver/joura"
)

const Version = "0.1.1"

func main() {
	var file string
	var version bool

	flag.BoolVar(&version, "v", false, "print version info")
	flag.StringVar(&file, "c", "/etc/joura/default.conf", "configuration file")
	flag.Parse()

	if version {
		fmt.Println("joura", Version)
		return
	}

	j, err := joura.New(file)

	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	j.Start()

}
