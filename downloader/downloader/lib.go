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
		log.Fatal("Error while downloading", uri, ":", err)
		return defaultreturn
	}

	// Verify if the response was ok
	if response.StatusCode != http.StatusOK {
		log.Fatal("Server return non-200 status: %v\n", response.Status)
		return defaultreturn
	}

	length, _ := strconv.Atoi(response.Header.Get("Content-Length"))
	sourceSize := uint64(length)

	filename := "defaultname"
	_, params, err := mime.ParseMediaType(response.Header.Get("Content-Disposition"))
	if err == nil {
		filename = params["filename"] // set to "foo.png"
	}

	if filename == "defaultname" {
		u, err := url.Parse(uri)
		if err != nil {
			filename = strings.TrimLeft(u.Path, "/")
		}
	}

	// for k, v := range response.Header {
	// 	log.Println("key:", k, "value:", v)
	// }

	return FileInfo{sourceSize, filename}
}

func down(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	urlstring := args[0]
	if len(args) > 1 {
		urlstring = args[1]
	}

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

	id, ok := cacheids.Get(urlstring)
	if ok {
		log.Println("Founded in cache!")
		bot.Answer(msg).Document(id).End()
		return nil
	}

	// Get file info
	info := file_info(urlstring)
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
Size: %s`, moreinfo, urlstring, info.Name, prettysize)).End()
		return nil
	}

	// Notify that we are going to enqueue it
	bot.Answer(msg).Text(fmt.Sprintf(`✔Your download has been enqueued and will be sended ASAP✔

URL: %s
Name: %s
Size: %s`, urlstring, info.Name, prettysize)).End()

	log.Printf("Downloading URL: %s.\n", urlstring)

	// The guy are downloading
	cacheuserdownload.Set(chatid, urlstring, cache.DefaultExpiration)

	wr, c := NewWorkRequest(msg, urlstring, info.Name, bot)

	// Enqueue ^^
	WorkQueue <- wr

	// Only send uploading document if size is > 5MB
	if size > MAX_SEND_UPLOAD {
		// This is just to send that we are uploading document
		bot.Answer(msg).Action(tgbot.UploadDocument).End()
	Download:
		for {
			select {
			case <-c:
				break Download
			case <-time.After(time.Second * 7):
				bot.Answer(msg).Action(tgbot.UploadDocument).End()
			}
		}
	} else {
		<-c
	}

	cacheuserdownload.Delete(chatid)

	return nil
}

func tricks(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	dropboxhelp := `Usually dropbox gives an URL like:
https://www.dropbox.com/blablabla?dl=0

This link works for browser, but not for direct download, change the last ?dl=0 for ?raw=1. The previous link becomes:
https://www.dropbox.com/blablabla?raw=1
`
	bot.Answer(msg).Text(fmt.Sprintf(`Here are some tricks that you can use in your links:

Dropbox: %s`, dropboxhelp)).End()

	return nil
}

func help(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.Answer(msg).Text(fmt.Sprintf(`Hi! I'm Downloader Bot! How are you %s?

I'll help you to download files :)

You can only download one file at a time, and there are a general queue to not flood my free server, so maybe it take some time to download it (you will see that are downloading when the bot sends "Uploading document").

You can download in two ways:
- Send the URL
- Send: /down URL

Other commands:
- /help - Show this help
- /start - Start the bot (and show this help too :P)
- /tricks - Show some URL tricks

Icon made by Dirtyworks (License: CC BY 3.0)

This bot is open source and has been created by @rock_neurotiko, I hope that you like it ;-)

The source code can be founded in: https://github.com/rockneurotiko/go-bots/tree/master/downloader

If you like it you can vote this bot in @storebot: https://telegram.me/storebot?start=simple_downloader_bot
`, msg.From.FirstName)).DisablePreview(true).End()
	return nil
}

func BuildBot(token string, workers int) *tgbot.TgBot {
	bot := tgbot.New(token).
		SimpleCommandFn(`help`, help).
		SimpleCommandFn(`start`, help).
		SimpleCommandFn(`tricks`, tricks).
		CommandFn(`down (.+)`, down).
		RegexFn(`^([^/].+)`, down).
		AnyMsgFn(func(bot tgbot.TgBot, msg tgbot.Message) {
		log.Println(msg)
	})

	StartDispatcher(workers) // more?

	return bot
}
