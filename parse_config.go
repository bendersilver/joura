package joura

// #cgo LDFLAGS: -lsystemd
// #include <systemd/sd-journal.h>
// #include <stdlib.h>
import "C"
import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/bendersilver/jlog"
	"github.com/bendersilver/nanobot"
	"gopkg.in/yaml.v3"
)

type (
	// Teleg -
	Teleg struct {
		Token string  `yaml:"token"`
		Chats []int64 `yaml:"chats"`
	}
	// Srvc -
	Srvc struct {
		Chats []string `yaml:"chats"`
		Level int      `yaml:"log_level"`
	}
)

func parseConfig(files ...string) (Joura, error) {
	var err error
	var c struct {
		T map[string]Teleg `yaml:"telegram"`
		S map[string]Srvc  `yaml:"service"`
	}
	// telemap = make(map[string]Teleg)
	// svrcmap = make(map[string]Teleg)
	for _, f := range files {
		b, err := ioutil.ReadFile(f)
		if err != nil {
			jlog.Warning(err)
			continue
		}
		err = yaml.Unmarshal(b, &c)
		if err != nil {
			jlog.Warning(err)
			continue
		}
	}

	var cfg Joura = make(map[string]*service)
	for name, srvc := range c.S {
		cfg[name] = new(service)
		cfg[name].time = C.uint64_t(time.Now().UnixMicro() - 1000)
		cfg[name].tg = make(map[string]*tgBot)
		if srvc.Level == 0 {
			srvc.Level = 8
		}
		cfg[name].level = srvc.Level
		// loop telegram
		for _, tg := range srvc.Chats {
			var b tgBot
			if tele, ok := c.T[tg]; ok {
				b.bot, err = nanobot.New(tele.Token)
				if err != nil {
					if len(tele.Token) > 7 {
						tele.Token = tele.Token[:10]
					}
					jlog.Errorf("token '%s...' error: %s", tele.Token, err.Error())
					continue
				}
				b.chats = append(b.chats, tele.Chats...)
				cfg[name].tg[tele.Token] = &b
			} else {
				jlog.Warningf("service '%s': telegram key '%s' not found. pass", name, tg)
				continue
			}
			b.clean()
		}
	}

	for k, v := range cfg {
		for tk, b := range v.tg {
			if len(b.chats) == 0 {
				delete(v.tg, tk)
				jlog.Warningf("service '%s', bot token '%s...': empty chats. pass", k, tk[:10])
			}
		}
		if len(v.tg) == 0 {
			delete(cfg, k)
			jlog.Warningf("service '%s' empty bot chats. pass", k)
		} else {
			v.unit = k
			v.match = []string{
				"_SYSTEMD_UNIT=" + k,
				"UNIT=" + k,
			}
			if !strings.HasSuffix(k, ".service") {
				v.match = append(v.match,
					"_SYSTEMD_UNIT="+k+".service",
					"UNIT="+k+".service",
				)
			}
			jlog.Noticef("start watch service '%s'", k)
		}
	}
	if len(cfg) == 0 {
		jlog.Warning("joura started empty")
	}
	return cfg, nil

}
