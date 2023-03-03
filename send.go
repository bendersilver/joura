package joura

import (
	"fmt"
	"os"
	"time"

	"github.com/bendersilver/jlog"
	"github.com/bendersilver/nanobot"
)

func (p *service) send() error {
	if p.buf.Len() == 0 {
		return nil
	}
	defer func() {
		p.buf.Reset()
		p.bufferFull = false
	}()

	var clean bool
	var body nanobot.Body
	var res *nanobot.Result

	hostname, _ := os.Hostname()
	body.Text = fmt.Sprintf("*%s | %s*\n\n", p.unit, hostname) + p.buf.String()
	body.Mode = "Markdown"

BASE:
	for bot, chats := range p.Telegram {
		var ix int
		for {
			if len(chats) <= ix {
				break
			}
			body.ChatID = chats[ix]
			res = bot.SendMessage(&body)

			switch res.Status {
			case nanobot.OK:
			case nanobot.BadChat:
				clean = true
				if len(chats) == ix+1 {
					chats = chats[:ix]
				} else {
					chats = append(chats[:ix], chats[ix+1:]...)
				}
				p.Telegram[bot] = chats
				jlog.Warningf("chat %d: %d %s\n", body.ChatID, res.Code, res.Desc)
				continue
			case nanobot.BadToken:
				clean = true
				delete(p.Telegram, bot)
				jlog.Warningf("token XXXXXXXXXX: %d %s\n", res.Code, res.Desc)
				continue BASE
			default:
				jlog.Warningf("%d %s\n", res.Code, res.Desc)
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
