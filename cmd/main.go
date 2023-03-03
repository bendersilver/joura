package main

import (
	"flag"
	"fmt"

	"github.com/bendersilver/jlog"
	"github.com/bendersilver/joura"
)

const Version = "0.1.2"

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
		jlog.Fatal(err)
	}
	j.Start()

}
