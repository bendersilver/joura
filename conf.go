package joura

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path"
	"sort"
	"time"

	"github.com/BurntSushi/toml"
)

// PkgConfig -
type PkgConfig struct {
	time     time.Time
	buf      bytes.Buffer
	Pass     bool               `toml:"pass"`
	Telegram map[string][]int64 `toml:"telegram"`
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
	}

	// UserConfig -
	UserConfig struct {
		Defaut  map[string]*TG      `toml:"telegram"`
		Service map[string]*Service `toml:"service"`
	}
)

// SetPkgConfig -
func SetPkgConfig() error {
	var c UserConfig
	err := readConf(&c)
	if err != nil {
		return err
	}
	return writeConf(&c)
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

// readConf -
func readConf(c *UserConfig) error {

	_, err := toml.DecodeFile(path.Join(os.Getenv("CONF_PATH"), "user.conf"), &c)
	if err != nil {
		log.Fatal(err)
	}
	return nil
}

// writeConf -
func writeConf(c *UserConfig) error {
	var cfg = map[string]*PkgConfig{}

	// loop servises
	for name, sv := range c.Service {
		cfg[name] = new(PkgConfig)
		cfg[name].Telegram = map[string][]int64{}

		// loop telegram
		var token string
		for _, tg := range sv.Chats {
			if tele, ok := c.Defaut[tg]; ok {
				token = tele.Token
			} else {
				fmt.Printf("W service `%s`: telegram key `%s` not found\n", name, tg)
				continue
			}
			cfg[name].Telegram[token] = append(cfg[name].Telegram[token], c.Defaut[tg].Chats...)
		}
	}

	// clean
	for _, val := range cfg {
		for k, v := range val.Telegram {
			if len(v) == 0 {
				delete(val.Telegram, k)
			} else {
				val.Telegram[k] = unq(val.Telegram[k])
			}
		}
		if len(val.Telegram) == 0 {
			val.Pass = true
			val.Telegram = nil
		}
	}

	// write
	f, err := os.OpenFile(path.Join(os.Getenv("CONF_PATH"), "pkg.cfg"), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer f.Close()
	f.WriteString("# config auto generated; DO NOT EDIT\n\n")
	enc := toml.NewEncoder(f)
	enc.Indent = "\t"
	return enc.Encode(&cfg)
}
