package downloader

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/carbocation/go-instagram/instagram"
	"github.com/pivotal-golang/bytefmt"
	"github.com/pmylund/go-cache"
	"github.com/rockneurotiko/go-tgbot"
	"github.com/rockneurotiko/slideshare"
)

const MAX_SIZE = 52428800
const MAX_SEND_UPLOAD = 102400 // 100KB/s

var cacheids = cache.New(60*time.Minute, 5*time.Minute)
var cacheuserdownload = cache.New(MAX_SIZE/MAX_SEND_UPLOAD*time.Second, MAX_SIZE/MAX_SEND_UPLOAD*time.Second) // 2 minutes max to download 50MB and/or let enqueue a new one

type FileInfo struct {
	Size uint64
	Name string
}

func file_info(uri string) FileInfo { 
	defaultreturn := FileInfo{0, ""} 
        response, err := http.Head(uri)
	if err != nil {
		log.Printf("Error while doing HEAD %s: %s", uri, err.Error())
		return defaultreturn
	}
	defer response.Body.Close()

	// Verify if the response was ok
	if response.StatusCode != http.StatusOK {
		log.Printf("Server return non-200 status to HEAD: %v\nLet's try GET\n", response.Status)
		response, err = http.Get(uri)
		if err != nil {
			log.Printf("Error while doing GET %s: %s", uri, err.Error())
			return defaultreturn
		}
		if response.StatusCode != http.StatusOK {
			log.Printf("Server return non-200 status to GET: %v", response.Status)
			return defaultreturn
		}
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
	kind := Video
	format := ""

	if len(args) > 1 {
		urlstring = args[1]
	}

	if len(args) > 2 {
		name = args[2]
		switch name {
		case "audio":
			name = ""
			kind = Audio
		case "video":
			name = ""
			kind = Video
		case "format":
			name = ""
			format = args[3]
		}
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

	// check if see in cache or no
	urlcachekind := kind.WithUrl(originalurl)
	id, ok := cacheids.Get(urlcachekind)
	if ok {
		log.Println("Founded in cache!")
		bot.Answer(msg).Document(id).End()
		return nil
	}

	cacheuserdownload.Set(chatid, originalurl, cache.DefaultExpiration)
	defer cacheuserdownload.Delete(chatid)

	tmpu := scrape_uri(UrlInfo{
		Url:     urlstring,
		Name:    "",
		Format:  format,
		Cookies: make([]*http.Cookie, 0),
		Error:   "",
		Kind:    kind,
	}, kind, bot, msg.From.ID)

	if tmpu.Error != "" {
		bot.Answer(msg).Text(tmpu.Error).End()
		return nil
	}
	urlstring = tmpu.Url
	if tmpu.Kind != kind {
		kind = tmpu.Kind
	}

	// Get file info
	info := file_info(urlstring)
	if tmpu.Name != "" {
		info.Name = strings.Replace(tmpu.Name, " ", "_", -1)
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
Size: %s`, moreinfo, originalurl, info.Name, prettysize)).DisablePreview(true).End()
		return nil
	}

	// Notify that we are going to enqueue it
	bot.Answer(msg).Text(fmt.Sprintf(`✔Your download has been enqueued and will be sended ASAP✔

URL: %s
Name: %s
Size: %s`, originalurl, info.Name, prettysize)).End()

	log.Printf("Downloading URL: %s.\n", urlstring)

	// The guy are downloading
	wr, c := NewWorkRequest(msg, urlstring, originalurl, kind, info.Name, bot, tmpu.Cookies)

	// Enqueue ^^
	WorkQueue <- wr

	answer := WorkAnswer{OkDownloading}
	ntimes := uint64(0)
	measured := size / MAX_SEND_UPLOAD
	times_sleep := uint64(7)

	// Only send uploading document if size is > 5MB
	if size > MAX_SEND_UPLOAD {
		// This is just to send that we are uploading document
		bot.Answer(msg).Action(tgbot.UploadDocument).End()
	Download:
		for {
			select {
			case answer = <-c:
				break Download
			case <-time.After(time.Second * time.Duration(times_sleep)):
				ntimes += times_sleep
				if ntimes >= measured+1 {
					answer = WorkAnswer{Timeout}
					break Download
				}
				bot.Answer(msg).Action(tgbot.UploadDocument).End()
			}
		}
	} else {
		answer = <-c
	}

	msgerror := ""
	if answer.Result == ErrorDownloading {
		msgerror = "Some error happened while trying to download your URL..."
	}

	if answer.Result == Timeout {
		msgerror = "My download estimation has timed out, but maybe the server is in a huge load, so you will maybe receive the file later ;-)"
	}

	if msgerror != "" {
		bot.Answer(msg).Text(msgerror).End()
	}

	return nil
}

func about(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	m := `This bot is open source and has been created by @rockneurotiko, I hope that you like it ;-)

The source code can be found in: https://github.com/rockneurotiko/go-bots/tree/master/downloader

The icon is made by Dirtyworks (License: CC BY 3.0)

Thanks for using it and rate in @storebot: https://telegram.me/storebot?start=simple_downloader_bot`
	bot.Answer(msg).Text(m).DisablePreview(true).End()
	return nil
}

func contact(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	m := `You can contact with the developer in Telegram talking to @rockneurotiko

Please, don't hesitate in contact if you find something weird or you have some suggestion :)

Thanks for using it and rate in @storebot: https://telegram.me/storebot?start=simple_downloader_bot`
	bot.Answer(msg).Text(m).DisablePreview(true).End()
	return nil
}

func googleformats(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	docf := strings.Join(doccheck.ValidFormats, ", ")
	presf := strings.Join(presentationcheck.ValidFormats, ", ")
	spreadf := strings.Join(spreadsheetcheck.ValidFormats, ", ")

	fmsg := fmt.Sprintf(`- Google Drive: The file format, you can't choose this one.
- Google Docs: %s
- Google Presentations: %s
- Google Spreadsheets: %s

Ask for the custom format when you send the URL, like:
<URL> format <format>`, docf, presf, spreadf)
	bot.Answer(msg).Text(fmsg).End()
	return nil
}

func supported(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.Answer(msg).Text(`Currently known supported (let me know if you know other):
>>= Direct link of a file
>>= Youtube (video or audio)
>>= Soundcloud
>>= Google Drive, Docs, Presentations, Spreadsheets
>>= Uploadboy (This is in testing!)
>>= SlideShare presentations
>>= Almost any site supported (But probably not every site :P) by youtube-dl: https://rg3.github.io/youtube-dl/supportedsites.html
`).DisablePreview(true).End()
	return nil
}

func help(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	bot.Answer(msg).Text(fmt.Sprintf(`Hi! I'm Downloader Bot! How are you %s?

I'll help you to download files :)

You can only download one file at a time, and there are a general queue to not flood my free server, so maybe it take some time to download it (you will see that are downloading when the bot sends "Uploading document").

Type /supported to see the currently knowed sites from you can download

You can download in some different ways:

>>= Just send the URL
https://soundcloud.com/monstercat/tristam-braken-flight

>>= Send the URL with a file name (the format are mandatory and there are just an space after the URL and before the file name :P)
https://soundcloud.com/monstercat/tristam-braken-flight monstercat_awesome.mp3

>>= Send the URL with "audio" at the end to try to send it instead of the video. Currently I know that works in youtube (let me know if you find other)
https://www.youtube.com/watch?v=xjnbC8fwslM audio

>>= Send the URL with "format <formatid>" at the end to send it in that format, currently only work with some Google services, use the command /googleformats to see the formats available
https://docs.google.com/document/d/1s_BZK92pzK6zby6LhDnB0giVNJR-yDxsoALvuAEX7d8/edit?usp=sharing format pdf

If you like it you can vote this bot in @storebot: https://telegram.me/storebot?start=simple_downloader_bot
`, msg.From.FirstName)).DisablePreview(true).End()
	return nil
}

// var bpsspeed uint64 = 0

var youtubedl string = ""
var instaclient *instagram.Client
var slideshareclient *slideshare.Service

func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)
	go func(c chan os.Signal) {
		<-c
		ac := make(chan bool, 1)
		StopDispatcher <- ac
		<-ac
		os.Exit(0)
	}(c)
}

func BuildBot(token string, workers int, youtubeurl string, instaid string, slideapi string, slidesecret string) *tgbot.TgBot {
	youtubedl = youtubeurl
	slideshareclient = &slideshare.Service{slideapi, slidesecret}
	instaclient = instagram.NewClient(nil)
	instaclient.ClientID = instaid
	get_all_domains_available()
	StartDispatcher(workers)
	setupSignalHandler()

	// fmt.Println("Start upstream test")
	// cfg, err := stdn.GetConfig()
	// if err == nil && len(cfg.Servers) > 0 {
	// 	for i, s := range cfg.Servers {
	// 		fmt.Println("Testing server", i)
	// 		bps, err := s.Upstream(3)
	// 		if err == nil {
	// 			bpsspeed = bps
	// 			fmt.Println("Founded speed!", bps)
	// 			break
	// 		}
	// 	}
	// }
	// fmt.Println("Finished upstream test")

	bot := tgbot.New(token).
		AnyMsgFn(func(bot tgbot.TgBot, msg tgbot.Message) {
		m := fmt.Sprintf("<%d", msg.From.ID)
		if msg.From.Username != nil {
			m = fmt.Sprintf("%s(%s)", m, *msg.From.Username)
		}
		m = fmt.Sprintf("%s>:", m)
		if msg.Text != nil {
			m = fmt.Sprintf("%s %s", m, *msg.Text)
		}
		log.Println(m)
	}).
		SimpleCommandFn(`help`, help).
		SimpleCommandFn(`supported`, supported).
		SimpleCommandFn(`start`, help).
		SimpleCommandFn(`about`, about).
		SimpleCommandFn(`contact`, contact).
		SimpleCommandFn(`googleformats`, googleformats).
		MultiRegexFn([]string{
		`^([^/]\S+)$`,
		`^([^/]\S+) (audio|video)$`,
		`^([^/]\S+) (format) (doc|pdf|txt|html|odt|pptx|xlsx)$`,
		`^([^/]\S+) ([a-zA-Z0-9_-]+\..+)$`,
	}, down)
	return bot
}
