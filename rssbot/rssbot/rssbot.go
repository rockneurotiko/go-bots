package rssbot

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"           // Read HTML
	rss "github.com/jteeuwen/go-pkg-rss"       // Subscribe to RSS
	"github.com/jteeuwen/go-pkg-xmlx"          // Dependency of RSS
	"github.com/pmylund/go-cache"              // Cache not use DB everytime
	"github.com/rockneurotiko/go-tgbot"        // Telegram Bot Library :)
	"github.com/syndtr/goleveldb/leveldb"      // Data base
	"github.com/syndtr/goleveldb/leveldb/util" // Utils for data base
	"gopkg.in/fatih/set.v0"                    // Set data structure
)

type User struct {
	id      int
	Options OptionsUser
}

// DB keys like:
// user:<id>
// user:<id>:<url>
// settings:<id>:<prop>
// rss:<url>
// rss:<url>:<id>
var dbdir = ""
var dblock = &sync.Mutex{}

// Properties for settings:
// Send or not images.

// Default time one day, clean every 5 minutes
// upgrades: Store users too
var rsscache = cache.New(60*time.Minute, 5*time.Minute)
var cachelock = &sync.RWMutex{}

var opscache = make(map[int]OptionsUser)
var opslock = &sync.RWMutex{}

// start - Start the bot
// help - Show this help
// sub  - Subscribe to that RSS
// list - Return your RSS subscriptions
// delete - Remove your subscription of the RSS <id> (an integer)
// rm - Remove your subscription of the RSS <id> (an integer)

var availableCommands = map[string]string{
	"/start":       "Start the bot",
	"/help":        "Show this help",
	"/sub <url>":   "Subscribe to that RSS, the URL need to have 'http://' or 'https://'",
	"/list":        "Return your RSS subscriptions",
	"/delete <id>": "Remove your subscription of the RSS <id> (an integer)",
	"/rm <id>":     "Remove your subscription of the RSS <id> (an integer)",
}

var helptoptext = `This is the available commands:`
var helpbottomtext = `Please, if you like and use this bot, consider vote in https://telegram.me/storebot?start=RSSNewsBot

Also you have any suggestion or issue you can contact with the main developer of this bot: @rock_neurotiko

All the code of this bot is Open Source, you can see it or contribute in https://github.com/rockneurotiko/go-bots/tree/master/rssbot`

func buildHelp() string {
	var buffer bytes.Buffer

	buffer.WriteString(helptoptext + "\n\n")

	orderk := []string{}
	for cmd := range availableCommands {
		orderk = append(orderk, cmd)
	}
	sort.Strings(orderk)
	for _, cmd := range orderk {
		htext := availableCommands[cmd]
		buffer.WriteString(fmt.Sprintf("%s - %s\n", cmd, htext))
	}

	buffer.WriteString("\n" + helpbottomtext)
	return buffer.String()
}

func help(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	fmt.Printf("%d asked for help\n", msg.Chat.ID)
	bot.Answer(msg).Text(buildHelp()).ReplyToMessage(msg.ID).End()
	return nil
}

func remove(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	id := fmt.Sprintf("%d", msg.Chat.ID)
	key := buildKey("user", id, "") + ":"
	n := args[1]
	i, e := strconv.Atoi(n)
	if e != nil {
		return nil
	}

	fmt.Printf("%d asked to remove %d\n", msg.Chat.ID, i)

	k, _, err := getNthDb(key, i)
	if err != nil {
		bot.Answer(msg).Text(fmt.Sprintf("Some error happened:\nError: %s", err.Error())).ReplyToMessage(msg.ID).End()
		return nil
	}
	url := strings.TrimLeft(k, key)
	urlkey := buildKey("rss", url, id)

	removeFromCacheUsers(url, msg.Chat.ID)
	multiDeleteDb([]string{k, urlkey})
	bot.Answer(msg).Text(fmt.Sprintf("RSS %s removed!", url)).ReplyToMessage(msg.ID).End()
	return nil
}

func subs(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	if len(args) != 2 {
		bot.Answer(msg).Text("Usage of sub command:\n/sub <RSS_url>").ReplyToMessage(msg.ID).End()
		return nil
	}
	fmt.Printf("%d asked to subscribe to %s\n", msg.Chat.ID, args[1])
	go botPollSubscribe(bot, msg, args[1], 5, charsetReader)
	return nil
}

func list(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	fmt.Printf("%d asked for his list.\n", msg.Chat.ID)
	id := fmt.Sprintf("%d", msg.Chat.ID)
	userkey := buildKey("user", id, "")
	allusern := loadFromDbPrefix(userkey + ":")
	var buffer bytes.Buffer
	buffer.WriteString(fmt.Sprintf("You are subscribed to %d RSS.\n---------\n", len(allusern)))
	i := 0
	for urlkey := range allusern {
		parts := strings.Split(urlkey, ":")
		url := strings.Join(parts[2:], ":")
		buffer.WriteString(fmt.Sprintf("%d) %s\n", i, url))
		i++
	}
	longSend(bot, msg.Chat.ID, buffer.String())
	return nil
}

func readAllDbRss(bot tgbot.TgBot) {
	allrss := loadFromDbPrefix("rss:")
	for urlkey := range allrss {
		splitted := strings.Split(urlkey, ":")
		if len(splitted) != 3 {
			continue
		}
		uri := strings.Join(splitted[1:], ":")

		go func(uri string, firsttime bool) {
			// Add a random wait for every url?
			feed := rss.New(5, true, chanHandler, botItemHandler(bot, true))
			for {
				if err := feed.Fetch(uri, charsetReader); err != nil {
					fmt.Fprintf(os.Stderr, "[e] %s: %s\n", uri, err)
					return
				}
				if firsttime {
					firsttime = false
					cachelock.Lock()
					rsscache.Set(urlkey, "true", cache.DefaultExpiration)
					cachelock.Unlock()
				}
				// Usually every 5 mins
				<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))
			}
		}(uri, true)
	}
}

// BuildBot ...
func BuildBot(dbd string, token string) *tgbot.TgBot {
	dbdir = dbd

	bot := tgbot.New(token).
		MultiCommandFn([]string{`sub (https?://.+)`, `sub ?.*`}, subs).
		MultiCommandFn([]string{`delete (\d+)`, `rm (\d+)`}, remove).
		SimpleCommandFn(`list`, list).
		SimpleCommandFn(`help`, help).
		SimpleCommandFn(`start`, help)

	bot.DefaultDisableWebpagePreview(true)

	// Start all saved RSS
	readAllDbRss(*bot)
	return bot
}

func buildKey(base string, id string, extra string) string {
	res := fmt.Sprintf("%s:%s", base, id)
	if extra != "" {
		res = fmt.Sprintf("%s:%s", res, extra)
	}
	return res
}

func deleteFromDb(k string) bool {
	dblock.Lock()
	defer dblock.Unlock()
	db, err := leveldb.OpenFile(dbdir, nil)
	defer db.Close()
	if err != nil {
		return false
	}
	err = db.Delete([]byte(k), nil)
	return err == nil
}

func multiDeleteDb(ks []string) {
	dblock.Lock()
	defer dblock.Unlock()
	db, err := leveldb.OpenFile(dbdir, nil)
	defer db.Close()
	if err != nil {
		return
	}
	for _, k := range ks {
		db.Delete([]byte(k), nil)
	}
	return
}

func loadFromDb(k string) string {
	dblock.Lock()
	defer dblock.Unlock()
	db, err := leveldb.OpenFile(dbdir, nil)
	defer db.Close()
	if err != nil {
		return ""
	}
	data, err := db.Get([]byte(k), nil)
	if err != nil {
		return ""
	}
	return string(data)
}

func getNthDb(p string, n int) (string, string, error) {
	dblock.Lock()
	defer dblock.Unlock()
	db, err := leveldb.OpenFile(dbdir, nil)
	defer db.Close()
	if err != nil {
		fmt.Println("Wut")
		return "", "", err
	}
	i := 0
	iter := db.NewIterator(util.BytesPrefix([]byte(p)), nil)
	for iter.Next() {
		// Use key/value.
		if i == n {
			k := string(iter.Key())
			v := string(iter.Value())
			iter.Release()
			return k, v, nil
		}
		i++
	}
	iter.Release()
	return "", "", fmt.Errorf("The number %d is not valid", n)
}

func loadFromDbPrefix(p string) map[string]string {
	dblock.Lock()
	defer dblock.Unlock()
	res := make(map[string]string, 0)
	db, err := leveldb.OpenFile(dbdir, nil)
	defer db.Close()
	if err != nil {
		return res
	}
	iter := db.NewIterator(util.BytesPrefix([]byte(p)), nil)
	for iter.Next() {
		// Use key/value.
		res[string(iter.Key())] = string(iter.Value())
	}
	iter.Release()
	return res
}

func saveInDb(mult map[string]string) { //, k string, v string) {
	dblock.Lock()
	defer dblock.Unlock()
	db, err := leveldb.OpenFile(dbdir, nil)
	defer db.Close()
	if err != nil {
		return
	}
	for k, v := range mult {
		err = db.Put([]byte(k), []byte(v), nil)
	}
}

func isValidFile(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

func saveAllValues(uri string, id string) {
	urlkey := buildKey("rss", uri, "")
	urluserkey := buildKey("rss", uri, id)
	userkey := buildKey("user", id, "")
	userurlkey := buildKey("user", id, uri)
	saveInDb(map[string]string{
		urlkey:     "true",
		urluserkey: "true",
		userurlkey: "true",
		userkey:    "true",
	})
}

func checkCache(uri string, id string) bool {
	urlkey := buildKey("rss", uri, "")
	cachelock.RLock()
	defer cachelock.RUnlock()
	_, ok := rsscache.Get(urlkey)
	if ok {
		saveAllValues(uri, id)
		return true
	}
	return false
}

func checkDb(uri string, id string) bool {
	urlkey := buildKey("rss", uri, "")
	val := loadFromDb(urlkey)
	if val != "" {
		saveAllValues(uri, id)
		cachelock.Lock()
		rsscache.Set(urlkey, "true", cache.DefaultExpiration)
		cachelock.Unlock()
		return true
	}
	return false
}

func removeFromCacheUsers(uri string, id int) {
	cachelock.Lock()
	defer cachelock.Unlock()
	if x, found := rsscache.Get(uri); found {
		switch val := x.(type) {
		case *set.Set:
			val.Remove(id)
		default:
			fmt.Println("Error rsscache: ", uri, val)
		}
		// foo := x.(*set.Set)
		// foo.Remove(id)
	}
}

func appendToCacheUsers(uri string, id int) {
	cachelock.Lock()
	defer cachelock.Unlock()
	if x, found := rsscache.Get(uri); found {
		switch val := x.(type) {
		case *set.Set:
			val.Add(id)
		default:
			fmt.Println("Error rsscache: ", uri, val)
		}
		// foo := x.(*set.Set)
		// foo.Add(id)
	}
}

func botPollSubscribe(bot tgbot.TgBot, msg tgbot.Message, uri string, timeout int, cr xmlx.CharsetFunc) {
	// If already a rss, only add him to the db
	// First, in cache:
	id := fmt.Sprintf("%d", msg.Chat.ID)
	ok := checkCache(uri, id)
	if ok {
		appendToCacheUsers(uri, msg.Chat.ID)
		bot.Answer(msg).Text(fmt.Sprintf("You have been subscribed to %s", uri)).ReplyToMessage(msg.ID).End()
		return
	}
	ok = checkDb(uri, id)
	if ok {
		appendToCacheUsers(uri, msg.Chat.ID)
		bot.Answer(msg).Text(fmt.Sprintf("You have been subscribed to %s", uri)).ReplyToMessage(msg.ID).End()
		return
	}

	firsttime := true
	// Adding new RSS
	feed := rss.New(timeout, true, chanHandler, botItemHandler(bot, true))

	for {
		if err := feed.Fetch(uri, cr); err != nil {
			fmt.Fprintf(os.Stderr, "[e] %s: %s\n", uri, err)
			bot.Answer(msg).Text(fmt.Sprintf("Bad RSS: %s, maybe the URL is bad.\nError msg: %s", uri, err.Error())).ReplyToMessage(msg.ID).End()
			return
		}
		if firsttime {
			saveAllValues(uri, id)
			appendToCacheUsers(uri, msg.Chat.ID)
			cachelock.Lock()
			rsscache.Set(buildKey("rss", uri, ""), "true", cache.DefaultExpiration)
			cachelock.Unlock()
			firsttime = false
			bot.Answer(msg).Text(fmt.Sprintf("You have been subscribed to %s", uri)).ReplyToMessage(msg.ID).End()
		}
		// Usually every 5 mins
		<-time.After(time.Duration(feed.SecondsTillUpdate() * 1e9))

	}
}

func botItemHandler(bot tgbot.TgBot, firsttime bool) rss.ItemHandlerFunc {
	return func(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
		// fmt.Printf("%d new item(s) in %s, firsttime: %v\n", len(newitems), feed.Url, firsttime)

		if firsttime {
			firsttime = false
			return
		}

		newst := ExtractNews(newitems)

		sendToAll(bot, feed.Url, newst)
	}
}

// NewStruct represent a "new" document
type NewStruct struct {
	Text        string
	Images      []string
	Title       string
	Description string
	Date        string
	Author      string
	LinksText   string
	Links       []string
}

// BuildText from the data
func (ns *NewStruct) BuildText() string {
	if ns.Text != "" {
		return ns.Text
	}

	headerlinks := `[`
	for _, l := range ns.Links {
		headerlinks = fmt.Sprintf("%s %s", headerlinks, l)
	}
	headerlinks = headerlinks + ` ]`

	descr := ns.Description
	if descr != "" {
		descr = "\n" + descr + "\n---------"
	}
	by := ns.Author
	if by == "" {
		by = "No author defined."
	}

	ns.Text = fmt.Sprintf(`%s
---------
%s
---------%s
By: %s
Date: %s
---------`, headerlinks, ns.Title, descr, by, ns.Date)

	return ns.Text
}

func ExtractNews(newitems []*rss.Item) []NewStruct {
	var newst []NewStruct
	for _, new := range newitems {
		// init
		// linkstr := ""
		linkslist := make([]string, 0)
		var images []string
		descrip := ""

		// get all links
		if new.Links != nil {
			links := new.Links
			for _, l := range links {
				l2 := *l
				linkslist = append(linkslist, l2.Href)
				// linkstr += fmt.Sprintf(" - (%s)", l2.Href)
			}
		}

		// Read HTML
		read := strings.NewReader(new.Description)
		doc, err := goquery.NewDocumentFromReader(read)

		if err == nil {
			doc.Find("img[src]").Each(func(i int, s *goquery.Selection) {
				val, ok := s.Attr("src")
				if ok {
					images = append(images, val)
				}
			})
			descrip = doc.Text()
		}

		new.Title, descrip = analyzeTitleDescrip(new.Title, descrip)

		// itemstr := fmt.Sprintf("%s%s\n%s", new.Title, linkstr, descrip)
		newst = append(newst, NewStruct{"", images, new.Title, descrip, new.PubDate, new.Author.Name, "", linkslist})

		// newst = append(newst, NewStruct{itemstr, images})
	}
	return newst
}

func analyzeTitleDescrip(title string, descrip string) (string, string) {
	title = strings.TrimSpace(title)
	descrip = strings.TrimSpace(descrip)
	if strings.HasSuffix(title, "...") && strings.HasPrefix(descrip, title[:len(title)-3]) {
		title = descrip
		descrip = ""
	} else if title == descrip {
		descrip = ""
	}

	return title, descrip
}

func downloadImage(url string) (img image.Image, err error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	img, _, err = image.Decode(resp.Body)
	if err != nil {
		return
	}

	return
}

type OptionsUser struct {
	SendImage bool
}

func SaveSettingsToUser(i int, nops OptionsUser) {
	opslock.Lock()
	opscache[i] = nops
	opslock.Unlock()
	id := fmt.Sprintf("%d", i)
	opsmap := make(map[string]string)
	basek := buildKey("settings", id, "")

	opsmap[buildKey(basek, "send-image", "")] = fmt.Sprintf("%v", nops.SendImage)
	saveInDb(opsmap)
}

func LoadSettingsFromUser(i int) OptionsUser {
	id := fmt.Sprintf("%d", i)
	opslock.RLock()
	x, found := opscache[i]
	opslock.RUnlock()
	if found {
		return x
	}

	opsmap := make(map[string]string)
	basek := buildKey("settings", id, "")
	ops := []string{"send-image"}
	for _, k := range ops {
		key := buildKey(basek, k, "")
		value := loadFromDb(key)
		opsmap[k] = value
	}
	finalops := OptionsUser{
		SendImage: opsmap["send-image"] == "true",
	}
	return finalops
}

func sendToAll(bot tgbot.TgBot, uri string, newst []NewStruct) {
	var allones []int
	// useroptions := make(map[int]OptionsUser)
	// In cache?
	cachelock.RLock()
	x, found := rsscache.Get(uri)
	cachelock.RUnlock()
	if found {
		switch val := x.(type) {
		case *set.Set:
			allones = set.IntSlice(val)
		default:
			fmt.Println("Error in type in rsscache: ", val)
		}
	} else {
		// Search in db and send
		urikey := buildKey("rss", uri, "")
		alloness := loadFromDbPrefix(urikey + ":")
		users := set.New()

		for key := range alloness {
			splitted := strings.Split(key, ":")
			i, e := strconv.Atoi(splitted[len(splitted)-1])
			if e == nil {
				allones = append(allones, i)
				users.Add(i)
			}
		}
		cachelock.Lock()
		rsscache.Set(uri, users, cache.DefaultExpiration)
		cachelock.Unlock()
	}

	// Right now we are doing: for all user, send every new.
	// Maybe do it in the other way? For every new, send to all
	imagesids := make(map[string]string)
	for _, i := range allones {
		useroptions := LoadSettingsFromUser(i)
		// bot.Send(i).Text(fmt.Sprintf("%d new items for: %s", len(newst), uri)).End()
		for _, n := range newst {
			// Send text
			longSend(bot, i, n.BuildText())

			// handle options!
			if !useroptions.SendImage {
				continue
			}
			// Then images :)
			for _, im := range n.Images {
				// Search in cache
				id, ok := imagesids[im]
				if ok && id != "" {
					bot.Send(i).Photo(id).End()
					continue
				} else if ok && id == "" {
					continue
				}

				// If don't in cache, download, send and write in cache :)
				img, err := downloadImage(im)
				if err != nil {
					imagesids[im] = ""
					continue
				}

				sended := bot.Send(i).Photo(img).End()
				if sended.Ok && sended.Result.Photo != nil && len(*sended.Result.Photo) > 0 {
					newimg := *sended.Result.Photo
					imagesids[im] = newimg[0].FileID
				}
			}
		}
	}
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

func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
	// fmt.Printf("%d new channel(s) in %s\n", len(newchannels), feed.Url)
}

func charsetReader(charset string, r io.Reader) (io.Reader, error) {
	if charset == "ISO-8859-1" || charset == "iso-8859-1" {
		return r, nil
	}
	return nil, errors.New("Unsupported character set encoding: " + charset)
}
