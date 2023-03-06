package main

import (
	"flag"
	"fmt"

	"github.com/bendersilver/jlog"
	"github.com/bendersilver/joura"
)

// Version -
const Version = "0.1.2"

func main() {

	var file, dir string
	var version bool

	flag.BoolVar(&version, "v", false, "print version info")
	flag.StringVar(&file, "c", "/etc/joura/default.conf", "configuration file")
	flag.StringVar(&dir, "d", "", "configuration dir")
	flag.Parse()

	if version {
		fmt.Println("joura", Version)
		return
	}

	j, err := joura.New(file, dir)
	if err != nil {
		jlog.Fatal(err)
	}
	j.Start()

}
