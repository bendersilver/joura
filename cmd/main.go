package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bendersilver/joura"
	"github.com/coreos/go-systemd/v22/sdjournal"
)

// journalctl --user -n 10 -f -o cat

type buf struct {
	sync.Mutex
	buf      bytes.Buffer
	lastUnit string
}

// Joura -
type Joura struct {
	Service map[string]*joura.PkgConfig
	cmd     []string
}

func fatal(err error) {
	fmt.Println("F", err)
	os.Exit(1)
}

// id -u bot
func main() {
	r, err := sdjournal.NewJournalReader(sdjournal.JournalReaderConfig{
		Since: time.Duration(-15) * time.Minute,
		Matches: []sdjournal.Match{
			{
				Field: sdjournal.SD_JOURNAL_FIELD_SYSTEMD_UNIT,
				Value: "NetworkManager-dispatcher",
			},
		},
		Formatter: func(entry *sdjournal.JournalEntry) (string, error) {
			b, err := json.Marshal(map[string]any{
				"msg":  entry.Fields["MESSAGE"],
				"unit": entry.Fields["UNIT"],
				"time": entry.RealtimeTimestamp,
			})
			return string(b), err
		},
	})
	if err != nil {
		fmt.Println(err)
		fatal(err)
	}
	defer r.Close()

	b, err := ioutil.ReadAll(r)
	if err != nil {
		fmt.Println(err)
		fatal(err)
	}
	fmt.Printf("%s", b)
	return
	// 2023-02-17 11:42:28.814366
	fmt.Println(time.Now().Add(-time.Hour * 12).Format("'2006-01-02 15:04:05'"))
	fmt.Println(strings.Join([]string{"journalctl", "--user", "-o", "json", "--since", time.Now().Add(-time.Hour * 12).Format("'2006-01-02 15:04:05'")}, " "))

	cmd := exec.Command("journalctl", "--user", "-o", "json", "--since", time.Now().Add(-time.Hour*12).Format("'2006-01-02 15:04:05'"))
	rd, err := cmd.StdoutPipe()
	if err != nil {
		fmt.Println("1", err)
		fatal(err)
	}

	err = cmd.Start()
	if err != nil {
		fmt.Println("2", err)
		fatal(err)
	}

	b, err = ioutil.ReadAll(rd)
	if err != nil {
		fmt.Println("2", err)
		fatal(err)
	}
	fmt.Printf("%s", b)

	// var data struct {
	// 	Msg       json.RawMessage `json:"MESSAGE"`
	// 	Unit      string          `json:"_SYSTEMD_UNIT"`
	// 	Timestamp int64           `json:"_SOURCE_REALTIME_TIMESTAMP,string"`
	// }

	// d := json.NewDecoder(rd)
	// for d.More() {
	// 	err = d.Decode(&data)
	// 	if err != nil {
	// 		fmt.Println("W", err)
	// 	}
	// 	fmt.Printf("read %v\n", data)
	// 	fmt.Println(time.UnixMicro(data.Timestamp))
	// }

	err = cmd.Wait()
	if err != nil {

		fmt.Println("3", err)
		fatal(err)
	}
	os.Exit(1)
	// err := joura.SetPkgConfig()
	// if err != nil {
	// 	fatal(err)
	// }
	// joura.New()
	// fatal(fmt.Errorf("end"))

	// var j Joura
	// j.cmd = []string{"journalctl", "-f", "-o", "json", "-n", "0"}

	// _, err := toml.DecodeFile(path.Join(os.Getenv("CONF_PATH"), "pkg.cfg"), &j.Service)
	// if err != nil {
	// 	fatal(err)
	// }

	// for k, v := range j.Service {
	// 	if v.Pass {
	// 		delete(j.Service, k)
	// 		continue
	// 	}
	// 	fmt.Println(k, v)
	// 	j.cmd = append(j.cmd, "-u", k)
	// }

	// fmt.Println(j)
	// os.Exit(1)

	// cmd := exec.Command("journalctl", "--user", "-f", "-o", "json")
	// r, err := cmd.StdoutPipe()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// var b buf
	// go scan(bufio.NewScanner(r), &b)
	// go sender(&b)

	// err = cmd.Start()
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// err = cmd.Wait()
	// if err != nil {
	// 	log.Fatal(err)
	// }
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
	// var resp struct {
	// 	OK      bool   `json:"ok"`
	// 	ErrCode int    `json:"error_code"`
	// 	Desc    string `json:"description"`
	// }

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
				// r, err := req.Post("https://api.telegram.org/bot"+os.Getenv("BOT")+"/sendMessage",
				// 	req.BodyJSON(data),
				// )
				// if err != nil {
				// 	log.Println(err)
				// 	continue
				// }
				// err = r.ToJSON(&resp)
				// if err != nil {
				// 	log.Println(err)
				// 	continue
				// }
				// if !resp.OK {
				// 	log.Println(resp.ErrCode, resp.Desc)
				// }
			}
			if users == nil {
				log.Println(data.Text)
			}
		}
		time.Sleep(time.Second)
	}
}
