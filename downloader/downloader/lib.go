package downloader

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/pivotal-golang/bytefmt"
	"github.com/pmylund/go-cache"
	"github.com/rockneurotiko/go-tgbot"
	stdn "github.com/traetox/speedtest/speedtestdotnet"
)

const MAX_SIZE = 52428800
const MAX_SEND_UPLOAD = 5242880

var cacheids = cache.New(60*time.Minute, 5*time.Minute)
var cacheuserdownload = cache.New(2*time.Minute, 2*time.Minute) // 2 minutes max to download 50MB and/or let enqueue a new one

type FileInfo struct {
	Size uint64
	Name string
}

func file_info(uri string) FileInfo {
	defaultreturn := FileInfo{0, ""}

	response, err := http.Head(uri)
	if err != nil {
		log.Printf("Error while downloading %s: %s", uri, err.Error())
		return defaultreturn
	}

	// Verify if the response was ok
	if response.StatusCode != http.StatusOK {
		log.Printf("Server return non-200 status: %v\n", response.Status)
		return defaultreturn
	}

	length, _ := strconv.Atoi(response.Header.Get("Content-Length"))
	sourceSize := uint64(length)

	filename := ""
	_, params, err := mime.ParseMediaType(response.Header.Get("Content-Disposition"))
	if err == nil {
		filename = params["filename"] // set to "foo.png"
	}

	if filename == "" {
		u, err := url.Parse(uri)
		if err == nil {
			s := strings.Split(u.Path, "/")
			if len(s) > 0 {
				filename = s[len(s)-1]
			}
		}
	}

	if filename == "" {
		filename = "defaultname"
	}

	// for k, v := range response.Header {
	// 	log.Println("key:", k, "value:", v)
	// }

	return FileInfo{sourceSize, filename}
}

func down(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	urlstring := args[0]
	name := ""

	if len(args) > 1 {
		urlstring = args[1]
	}

	if len(args) > 2 {
		name = args[2]
	}

	originalurl := urlstring

	chatid := fmt.Sprintf("%d", msg.Chat.ID)

	founded, ok := cacheuserdownload.Get(chatid)
	if ok && founded.(string) != "" {
		bot.Answer(msg).Text(fmt.Sprintf("🚫 You are already downloading %s, wait to finish the download.", founded.(string))).End()
		return nil
	}

	// parse URL
	_, err := url.ParseRequestURI(urlstring)
	if err != nil {
		bot.Answer(msg).Text(fmt.Sprintf("🚫 You have to send me an URL to download it and %s is not ;-)", urlstring)).End()
		return nil
	}

	tmpu := scrape_uri(UrlInfo{urlstring, ""})
	urlstring = tmpu.Url

	// id, ok := cacheids.Get(urlstring)
	id, ok := cacheids.Get(originalurl)
	if ok {
		log.Println("Founded in cache!")
		bot.Answer(msg).Document(id).End()
		return nil
	}

	// Get file info
	info := file_info(urlstring)
	if tmpu.Name != "" {
		info.Name = tmpu.Name
	}
	if name != "" {
		info.Name = name
	}

	size := info.Size
	prettysize := bytefmt.ByteSize(size)

	can_down := size > 0 && size < MAX_SIZE
	if !can_down {
		moreinfo := "The size is more than 50MB (Telegram Bots only can send 50MB)"
		if size <= 0 {
			moreinfo = "I didn't detected any size at all, if you think that I should contact me!"
		}
		bot.Answer(msg).Text(fmt.Sprintf(`✖✖ I can't download the file ✖✖

Reason: %s

URL: %s
Name: %s
Size: %s`, moreinfo, originalurl, info.Name, prettysize)).End()
		return nil
	}

	// Notify that we are going to enqueue it
	bot.Answer(msg).Text(fmt.Sprintf(`✔Your download has been enqueued and will be sended ASAP✔

URL: %s
Name: %s
Size: %s`, originalurl, info.Name, prettysize)).End()

	log.Printf("Downloading URL: %s.\n", urlstring)

	// The guy are downloading
	cacheuserdownload.Set(chatid, originalurl, cache.DefaultExpiration)

	wr, c := NewWorkRequest(msg, urlstring, originalurl, info.Name, bot)

	// Enqueue ^^
	WorkQueue <- wr

	answer := WorkAnswer{true}
	ntimes := uint64(0)
	measured := size / MAX_SEND_UPLOAD
	if bpsspeed > 0 {
		measured = (size * 8) / bpsspeed // in bits
	}

	// Only send uploading document if size is > 5MB
	if size > MAX_SEND_UPLOAD {
		// This is just to send that we are uploading document
		bot.Answer(msg).Action(tgbot.UploadDocument).End()
	Download:
		for {
			select {
			case answer = <-c:
				break Download
			case <-time.After(time.Second * 7):
				ntimes++
				if ntimes >= measured+1 {
					answer = WorkAnswer{false}
					break Download
				}
				bot.Answer(msg).Action(tgbot.UploadDocument).End()
			}
		}
	} else {
		answer = <-c
	}

	if !answer.Result {
		bot.Answer(msg).Text("Some error happened while trying to download your URL...").End()
	}

	cacheuserdownload.Delete(chatid)

	return nil
}

func tricks(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	// - Youtube: Uses an API to find the media download
	helptext := `I will try to do some automatic detections to find the media you want. Currently supported:
- SoundCloud: Uses an API to find the media download
- Dropbox: Change "dl=0" to "raw=1" to download
`
	bot.Answer(msg).Text(helptext).End()

	return nil
}

func help(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.Answer(msg).Text(fmt.Sprintf(`Hi! I'm Downloader Bot! How are you %s?

I'll help you to download files :)

You can only download one file at a time, and there are a general queue to not flood my free server, so maybe it take some time to download it (you will see that are downloading when the bot sends "Uploading document").

You can download in two ways:
- Send the URL, for example, this will send you the song:
https://soundcloud.com/monstercat/tristam-braken-flight
- Send the URL and a name of file, for example, this will send you the song with the name "monstercat_awesome.mp3":
https://soundcloud.com/monstercat/tristam-braken-flight monstercat_awesome.mp3

To try to be fast, I have a cache that wipes the URL's if hadn't been used in one hour. Thanks to that you can ask for an URL two times and the second time will be instant.
But, because of that, if someone ask an URL with name, and you ask the same URL, you will gate that name no matter what, until the cache is wiped :)

Other commands:
- /help - Show this help
- /start - Start the bot (and show this help too :P)
- /tricks - Show some URL tricks

Icon made by Dirtyworks (License: CC BY 3.0)

This bot is open source and has been created by @rockneurotiko, I hope that you like it ;-)

The source code can be founded in: https://github.com/rockneurotiko/go-bots/tree/master/downloader

If you like it you can vote this bot in @storebot: https://telegram.me/storebot?start=simple_downloader_bot
`, msg.From.FirstName)).DisablePreview(true).End()
	return nil
}

var bpsspeed uint64 = 0

func BuildBot(token string, workers int) *tgbot.TgBot {
	StartDispatcher(workers)

	fmt.Println("Start upstream test")
	cfg, err := stdn.GetConfig()
	if err == nil && len(cfg.Servers) > 0 {
		for i, s := range cfg.Servers {
			fmt.Println("Testing server", i)
			bps, err := s.Upstream(3)
			if err == nil {
				bpsspeed = bps
				fmt.Println("Founded speed!", bps)
				break
			}
		}
	}
	fmt.Println("Finished upstream test")
	bot := tgbot.New(token).
		SimpleCommandFn(`help`, help).
		SimpleCommandFn(`start`, help).
		SimpleCommandFn(`tricks`, tricks).
		// MultiCommandFn([]string{`down (\S+)`, `down (\S+) ([a-zA-Z0-9_-]+\..+)`}, down).
		MultiRegexFn([]string{`^([^/]\S+)$`, `^([^/]\S+) ([a-zA-Z0-9_-]+\..+)$`}, down).
		AnyMsgFn(func(bot tgbot.TgBot, msg tgbot.Message) {
		log.Println(msg)
	})

	return bot
}
