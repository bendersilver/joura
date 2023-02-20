package main

import (
	"fmt"
	"os"

	"github.com/bendersilver/joura"
)

// journalctl --user -n 10 -f -o cat

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "aply" {
			err := joura.SetPkgConfig()
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				fmt.Println("OK")
			}
		}
	} else {
		j, err := joura.New()

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		j.Start()
	}
}
