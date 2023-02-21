package main

import (
	"fmt"
	"os"

	"github.com/bendersilver/joura"
)

// journalctl --user -n 10 -f -o cat

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "apply":
			err := joura.SetPkgConfig()
			if err != nil {
				fmt.Println(err)
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
