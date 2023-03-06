package joura

import (
	"io/ioutil"

	"github.com/bendersilver/jlog"
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

func parseConfig(files ...string) {
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

	// var j Joura

	// for name, item := range c.S {

	// }

	jlog.Info("%+v", c)
}
