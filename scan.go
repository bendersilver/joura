package joura

// #cgo LDFLAGS: -lsystemd
// #include <systemd/sd-journal.h>
// #include <stdlib.h>
import "C"
import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bendersilver/jlog"
	"github.com/bendersilver/nanobot"
)

// Joura -
type Joura map[string]*service

type tgBot struct {
	chats []int64
	bot   *nanobot.Bot
}

// clean - remove duplicate
func (t *tgBot) clean() {
	u := make([]int64, 0, len(t.chats))
	m := make(map[int64]bool)

	for _, val := range t.chats {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	t.chats = u
}

// service -
type service struct {
	time  C.uint64_t
	unit  string
	match []string
	buf   bytes.Buffer
	level int
	tg    map[string]*tgBot
}

// Start -
func (j Joura) Start() {
	var err error
	for range time.Tick(time.Second * 5) {
		for name, c := range j {
			err = journalRead(c)
			if err != nil {
				jlog.Error(err)
			}
			err = c.send()
			if err != nil {
				jlog.Error(err)
			}
			// c.clean()
			if c.tg == nil {
				jlog.Warningf("service '%s': empty chats. pass", name)
				delete(j, name)
			}
		}
	}
}

// New -
func New(fileConf, dirConf string) (Joura, error) {
	var files []string
	if dirConf != "" {
		err := filepath.Walk(dirConf, func(path string, info os.FileInfo, err error) error {

			if err != nil {

				fmt.Println(err)
				return nil
			}

			if !info.IsDir() && filepath.Ext(path) == ".conf" {
				files = append(files, path)
			}

			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		files = append(files, fileConf)
	}
	return parseConfig(files...)
}
