package rssbot

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	rss "github.com/jteeuwen/go-pkg-rss" // Subscribe to RSS
	"github.com/rockneurotiko/go-tgbot"
	"gopkg.in/fatih/set.v0"
)

func buildKey(base string, id string, extra string) string {
	res := fmt.Sprintf("%s:%s", base, id)
	if extra != "" {
		res = fmt.Sprintf("%s:%s", res, extra)
	}
	return res
}

func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	// fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func charsetReader(charset string, r io.Reader) (io.Reader, error) {
	if charset == "ISO-8859-1" || charset == "iso-8859-1" {
		return r, nil
	}
	return nil, errors.New("Unsupported character set encoding: " + charset)
}

func isValidFile(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

func longSend(bot tgbot.TgBot, i int, text string) {
	newstr := divideAndConquer(text)
	for _, s := range newstr {
		bot.Send(i).Text(s).End()
	}
}

func divideAndConquer(str string) []string {
	newstr := []string{}
	for {
		if len(str) < 4096 {
			newstr = append(newstr, str)
			break
		}
		newstr = append(newstr, str[:4096])
		str = str[4096:]
	}
	return newstr

}

func sendWakingUpMessage(bot tgbot.TgBot, send bool) {
	if !send {
		return
	}

	text := fmt.Sprintf(`%s

Actual version: %s`, messageUpdate, version)

	if lastchangelog != "" {
		text = text + fmt.Sprintf(`

Changelog:

%s`, lastchangelog)
	}
	users := getAllActiveUsers()

	for _, u := range users {
		go longSend(bot, u, text)
	}
}

func getAllActiveUsers() []int {
	var res []int
	_, _, _, users, _ := getStats()
	for id, n := range users {
		if n <= 0 {
			continue
		}
		idint, e := strconv.Atoi(id)
		if e == nil {
			res = append(res, idint)
		}
	}
	return res
}

func getStats() (chats *set.Set, users *set.Set, rss *set.Set, nperuser map[string]int, subscribed map[string]int) {
	chats = set.New()
	users = set.New()
	rss = set.New()
	nperuser = map[string]int{}
	subscribed = map[string]int{}

	allv := loadFromDbPrefix("")
	for k := range allv {
		if strings.HasPrefix(k, "user") && len(strings.Split(k, ":")) == 2 {
			id := strings.TrimLeft(k, "user:")
			if i, e := strconv.Atoi(id); e == nil {
				nperuser[id] = 0
				if i > 0 {
					users.Add(i)
				} else {
					chats.Add(i)
				}
			}
		} else if strings.HasPrefix(k, "user") {
			uid := strings.Split(k, ":")[1]

			vu, oku := nperuser[uid]
			if oku {
				nperuser[uid] = vu + 1
			} else {
				nperuser[uid] = 1
			}

			surl := strings.Join(strings.Split(k, ":")[2:], ":")
			v, ok := subscribed[surl]
			if ok {
				subscribed[surl] = v + 1
			} else {
				subscribed[surl] = 1
			}
		}

		if strings.HasPrefix(k, "rss") && len(strings.Split(k, ":")) == 3 {
			rss.Add(strings.TrimLeft(k, "rss:"))
		}
	}
	return
}
