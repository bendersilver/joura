package joura

// #cgo LDFLAGS: -lsystemd
// #include <systemd/sd-journal.h>
// #include <stdlib.h>
import "C"
import (
	"bytes"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Joura -
type Joura map[string]*service

// service -
type service struct {
	time     C.uint64_t
	match    *C.char
	unit     string
	buf      bytes.Buffer
	level    int
	Telegram map[string][]int64
}

func (s *service) clean() {
	for k, v := range s.Telegram {
		if len(v) == 0 {
			delete(s.Telegram, k)
		} else {
			s.Telegram[k] = unq(s.Telegram[k])
		}
	}
	if len(s.Telegram) == 0 {
		s.Telegram = nil
	}
}

func unq(input []int64) []int64 {
	u := make([]int64, 0, len(input))
	m := make(map[int64]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	sort.Slice(u[:], func(i, j int) bool { return u[i] < u[j] })
	return u
}

// UserConfig
type (
	// TG -
	TG struct {
		Token string  `toml:"token"`
		Chats []int64 `toml:"chats"`
	}

	// Service -
	Service struct {
		Chats []string           `toml:"chats"`
		Tg    map[string][]int64 `toml:"tele_chats"`
		Level int                `toml:"log_level"`
	}

	// UserConfig -
	UserConfig struct {
		Defaut  map[string]*TG      `toml:"telegram"`
		Service map[string]*Service `toml:"service"`
	}
)

// Start -
func (j Joura) Start() {
	var err error
	for range time.Tick(time.Second * 5) {
		for name, c := range j {
			err = journalRead(c)
			if err != nil {
				fmt.Println(err)
			}
			err = c.send()
			if err != nil {
				fmt.Println(err)
			}
			c.clean()
			if c.Telegram == nil {
				fmt.Printf("W service `%s`: empty chats. pass\n", name)
				delete(j, name)
			}
		}
	}
}

// New -
func New() (Joura, error) {
	var c UserConfig
	_, err := toml.DecodeFile(path.Join(os.Getenv("CONF_PATH"), "user.conf"), &c)
	if err != nil {
		return nil, err
	}
	var cfg Joura = make(map[string]*service)
	// loop servises
	for name, sv := range c.Service {
		cfg[name] = new(service)
		cfg[name].Telegram = map[string][]int64{}
		if sv.Level == 0 {
			sv.Level = 8
		}
		cfg[name].level = sv.Level

		// loop telegram
		var token string
		for _, tg := range sv.Chats {
			if tele, ok := c.Defaut[tg]; ok {
				token = tele.Token
			} else {
				fmt.Printf("W service `%s`: telegram key `%s` not found. pass\n", name, tg)
				continue
			}
			cfg[name].Telegram[token] = append(cfg[name].Telegram[token], c.Defaut[tg].Chats...)
		}
		cfg[name].clean()
		if cfg[name].Telegram == nil {
			fmt.Printf("W service `%s`: empty chats. pass\n", name)
			delete(cfg, name)
		} else {
			cfg[name].unit = name
			if !strings.HasSuffix(cfg[name].unit, ".service") {
				cfg[name].unit += ".service"
			}
			cfg[name].match = C.CString("_SYSTEMD_UNIT=" + cfg[name].unit)
			cfg[name].time = C.uint64_t(time.Now().UnixMicro())
		}
	}
	return cfg, nil
}
