package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"log"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

type buf struct {
	sync.Mutex
	buf      bytes.Buffer
	lastUnit string
}

type (
	// Item -
	Item struct {
		Names []string `toml:"service_name"`
		Tg    struct {
			Token string   `toml:"token"`
			Chats []string `toml:"chats"`
		} `toml:"telegram"`
	}
	// Conf -
	Conf struct {
		Token string   `toml:"tg_token"`
		Chats []string `toml:"tg_chats"`
		Item  []Item   `toml:"item"`
	}
)

var cfg Conf

// id -u bot
func main() {
	_, err := toml.DecodeFile(path.Join(os.Getenv("CONF_PATH"), "joura.conf"), &cfg)

	cmd := exec.Command("journalctl", "--user", "-f", "-o", "json")
	r, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}

	var b buf
	go scan(bufio.NewScanner(r), &b)
	go sender(&b)

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}

func scan(s *bufio.Scanner, b *buf) {

	var data struct {
		Exe  string          `json:"_EXE"`
		Msg  json.RawMessage `json:"MESSAGE"`
		Com  string          `json:"_COMM"`
		Unit string          `json:"_SYSTEMD_UNIT"`
	}
	var err error
	for s.Scan() {
		err = json.Unmarshal(s.Bytes(), &data)
		if err != nil {
			log.Println(err, s.Text())
		} else {
			if strings.HasPrefix(data.Unit, "tdot") {
				b.Lock()
				if b.lastUnit != data.Unit {
					b.buf.WriteString(data.Exe + "\n")
				}
				b.lastUnit = data.Unit
				var msg string
				err = json.Unmarshal(data.Msg, &msg)
				if err == nil {
					b.buf.WriteString(msg)
				} else {
					var bt []byte
					json.Unmarshal(data.Msg, &bt)
					b.buf.Write(bt)
				}
				b.buf.WriteString("\n")
				b.Unlock()
			}
		}
	}
}

func sender(b *buf) {
	var resp struct {
		OK      bool   `json:"ok"`
		ErrCode int    `json:"error_code"`
		Desc    string `json:"description"`
	}

	var data struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}

	var users []int64
	for _, v := range strings.Split(os.Getenv("IDS"), ",") {
		id, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			log.Println(err)
		} else {
			users = append(users, id)
		}
	}

	for {
		if b.buf.Len() > 0 {
			b.Lock()
			data.Text = strings.TrimSpace(b.buf.String())
			b.buf.Reset()
			b.lastUnit = ""
			b.Unlock()
			for _, id := range users {
				data.ChatID = id
				r, err := req.Post("https://api.telegram.org/bot"+os.Getenv("BOT")+"/sendMessage",
					req.BodyJSON(data),
				)
				if err != nil {
					log.Println(err)
					continue
				}
				err = r.ToJSON(&resp)
				if err != nil {
					log.Println(err)
					continue
				}
				if !resp.OK {
					log.Println(resp.ErrCode, resp.Desc)
				}
			}
			if users == nil {
				log.Println(data.Text)
			}
		}
		time.Sleep(time.Second)
	}
}
