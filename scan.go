package joura

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
)

// Joura -
type Joura struct {
	service map[string]*PkgConfig
	cmd     []string
}

// New -
func New() (*Joura, error) {
	err := SetPkgConfig()
	if err != nil {
		return nil, err
	}

	var j Joura
	_, err = toml.DecodeFile(path.Join(os.Getenv("CONF_PATH"), "pkg.cfg"), &j.service)
	if err != nil {
		return nil, err
	}

	// var dt = time.Now().UTC()

	for k, v := range j.service {
		if v.Pass {
			delete(j.service, k)
		} else {
			j.cmd = append(j.cmd, "-u", k)
			v.time = time.UnixMicro(1176500001751063) //dt
		}
	}

	var data struct {
		Msg       json.RawMessage `json:"MESSAGE"`
		Unit      string          `json:"_SYSTEMD_UNIT"`
		Timestamp int64           `json:"_SOURCE_REALTIME_TIMESTAMP,string"`
	}

	var c = []string{"--user", "-o", "json", "--since", "2023-02-16 17:20:10"}
	// var cmd = []string{"-o", "json", "--utc", "--since", "YYYY-MM-DD HH:MM:SS", "-u", "unit_name"}
	for range j.service {
		fmt.Println(time.Now().Format("'2006-01-02 15:04:05'"))
		fmt.Println(time.Now().UTC().Format("'2006-01-02 15:04:05'"))
		// cmd[4], cmd[6] = v.Time.Format("'2006-01-02 15:04:05'"), k
		// fmt.Println("W", v.time.Format("'2006-01-02 15:04:05'"))
		// c[5] = v.time.Format("'2006-01-02 15:04:05'")
		cmd := exec.Command("journalctl", c...)
		rd, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatal(err)
		}
		if err := cmd.Start(); err != nil {
			log.Fatal(err)
		}
		d := json.NewDecoder(rd)
		for d.More() {
			err = d.Decode(&data)
			if err != nil {
				fmt.Println("W", err)
			}
			fmt.Printf("read %v\n", data)
			fmt.Println(time.UnixMicro(data.Timestamp))
		}

		if err := cmd.Wait(); err != nil {
			log.Fatal(err)
		}

		// s := bufio.NewScanner(bytes.NewReader(b))
		// for s.Scan() {
		// 	json.
		// 		fmt.Printf("I\t%s\t%s\n", k, s.Text())
		// }
	}
	// fmt.Printf("%s\n", b)

	return nil, nil
}

// Scan -
func (j *Joura) Scan() {
	// journalctl -o json -u nginx --utc --since "2023-02-16 10:15:00"
	// var cmd = []string{"-o", "json", "--utc", "--since", "YYYY-MM-DD HH:MM:SS", "-u", "unit_name"}
	var cmd = []string{"--user", "-n", "10"}
	for range j.service {
		// cmd[4], cmd[6] = v.Time.Format("'2006-01-02 15:04:05'"), k
		b, err := exec.Command("journalctl", cmd...).Output()
		if err != nil {
			fmt.Println("W", err)
		}
		fmt.Printf("%s\n", b)

		log.Println(strings.Join(cmd, " "))
	}

	// var data struct {
	// 	Msg  json.RawMessage `json:"MESSAGE"`
	// 	Unit string          `json:"_SYSTEMD_UNIT"`
	// }
	// var err error
	// var item *PkgConfig
	// var ok bool
	// var msg string
	// for s.Scan() {
	// 	err = json.Unmarshal(s.Bytes(), &data)
	// 	if err != nil {
	// 		fmt.Println("W", err)
	// 	} else {
	// 		if len(data.Msg) > 4000 {
	// 			continue
	// 		}
	// 		if item, ok = j.service[data.Unit]; ok {
	// 			item.Lock()
	// 			err = json.Unmarshal(data.Msg, &msg)
	// 			if err == nil {
	// 				item.buf.WriteString(msg[:10])
	// 			} else {
	// 				item.buf.Write(data.Msg)
	// 			}
	// 			item.buf.WriteString("\n")
	// 			if item.buf.Len() > 1024 {
	// 				item.buf.Truncate(1024)
	// 			}
	// 			item.Unlock()
	// 		}
	// 	}
	// }
}

// Sender -
// func (j *Joura) Sender() {
// 	for {
// 		for k, v := range j.service {
// 			log.Println(v.buf.String())
// 		}
// 		time.Sleep(time.Second)
// 	}
// }
