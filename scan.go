package joura

// #cgo LDFLAGS: -lsystemd
// #include <systemd/sd-journal.h>
// #include <stdlib.h>
import "C"
import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Joura -
type Joura map[string]*PkgConfig

// Start -
func (j Joura) Start() {
	var err error
	for range time.Tick(time.Second * 5) {
		for name, c := range j {
			if c.Pass {
				fmt.Printf("W exit service %s, chat empty\n", name)
				delete(j, name)
				continue
			}
			err = journalRead(c)
			if err != nil {
				fmt.Println(err)
			}
			c.send()
			// fmt.Println(name, c)

		}
	}
}

// New -
func New() (Joura, error) {
	var cfg Joura
	_, err := toml.DecodeFile(path.Join(os.Getenv("CONF_PATH"), "pkg.cfg"), &cfg)
	if err != nil {
		return nil, err
	}
	if len(cfg) == 0 {
		return nil, errors.New("E config empty")
	}
	for unit, c := range cfg {
		c.unit = unit
		if !strings.HasSuffix(c.unit, ".service") {
			c.unit += ".service"
		}
		c.match = C.CString("_SYSTEMD_UNIT=" + c.unit)
		c.time = C.uint64_t(time.Now().UnixMicro())

		if c.Level == 0 {
			c.Level = 8
		}
	}

	return cfg, nil
}
