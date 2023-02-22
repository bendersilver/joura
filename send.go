package joura

import (
	"fmt"
	"os"
	"time"

	"github.com/imroc/req"
)

func (p *service) send() error {
	if p.buf.Len() == 0 {
		return nil
	}
	hostname, _ := os.Hostname()

	var clean bool
	var uri string
	var data = struct {
		ChatID int64  `json:"chat_id"`
		Text   string `json:"text"`
	}{
		Text: fmt.Sprintf("%s | %s\n\n", p.unit, hostname) + p.buf.String(),
	}
	defer p.buf.Reset()

	var teleRsp struct {
		OK      bool   `json:"ok"`
		ErrCode int    `json:"error_code"`
		Desc    string `json:"description"`
	}
BASE:
	for token, chats := range p.Telegram {
		uri = "https://api.telegram.org/bot" + token + "/sendMessage"
		var ix int
		for {
			if len(chats) <= ix {
				break
			}
			data.ChatID = chats[ix]
			rsp, err := req.Post(uri, req.BodyJSON(data))
			if err != nil {
				return err
			}
			err = rsp.ToJSON(&teleRsp)
			if err != nil {
				return err
			}
			if !teleRsp.OK {
				switch teleRsp.ErrCode {
				case 403, 400:
					clean = true
					if len(chats) == ix+1 {
						chats = chats[:ix]
					} else {
						chats = append(chats[:ix], chats[ix+1:]...)
					}

					p.Telegram[token] = chats
					fmt.Printf("chat %d: %d %s\n", data.ChatID, teleRsp.ErrCode, teleRsp.Desc)
					continue
				case 401:
					clean = true
					delete(p.Telegram, token)
					fmt.Printf("token %s....: %d %s\n", token[:10], teleRsp.ErrCode, teleRsp.Desc)
					continue BASE
				default:
					fmt.Printf("%+v", teleRsp)
				}
			}
			ix++
			time.Sleep(time.Second / 20)
		}
	}
	if clean {
		p.clean()
	}
	return nil
}
