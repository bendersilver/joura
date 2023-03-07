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
	defer p.buf.Reset()

	// var clean bool
	var body nanobot.Body
	var res *nanobot.Result

	hostname, _ := os.Hostname()
	body.Text = fmt.Sprintf("%s | %s\n\n", p.unit, hostname) + p.buf.String() + "```\n"
	body.Mode = "Markdown"

BASE:
	for tk, t := range p.tg {
		var ix int
		for {
			if len(t.chats) <= ix {
				break
			}
			body.ChatID = t.chats[ix]
			res = t.bot.SendMessage(&body)

			switch res.Status {
			case nanobot.OK:
			case nanobot.BadChat:
				if len(t.chats) == ix+1 {
					t.chats = t.chats[:ix]
				} else {
					t.chats = append(t.chats[:ix], t.chats[ix+1:]...)
				}
				jlog.Warningf("token '%s...' chat %d: %d %s", tk[:10], body.ChatID, res.Code, res.Desc)
				continue
			case nanobot.BadToken:
				delete(p.tg, tk)
				jlog.Warningf("token '%s...': %d %s", tk[:10], res.Code, res.Desc)
				continue BASE
			default:
				jlog.Warningf("%d %s", res.Code, res.Desc)
			}
			ix++
			time.Sleep(time.Second / 20)
		}
	}
	return nil
}
