package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	godotenv "github.com/joho/godotenv"
	"github.com/rockneurotiko/go-homedir"
	"github.com/rockneurotiko/go-tgbot"
	"github.com/syndtr/goleveldb/leveldb"
)

// start - Start the bot
// help - Show this help
// cd - Change user directory
// ls - Show current user directory content
// download - Download specified file
// dw - Shortcut to /download :)

var availableCommands = map[string]string{
	"/start":              "Start the bot",
	"/help":               "Show this help",
	"/cd":                 "Change user directory",
	"/cd <name|id>":       "Change user directory to the specified",
	"/ls":                 "Show current user directory content",
	"/ls <name|id>":       "Show specified directory content",
	"/download <name|id>": "Download specified file",
	"/dw <name|id>":       "Shortcut to /download :)",
}

var bookid = map[string]string{}

var userpath = map[int]string{}
var usersallowed = map[string]bool{}

var base = ""
var dbdir = ""
var passphrasestart = ""

func start(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	if passphrasestart == "" {
		str := "Wellcome!"
		return &str
	}
	suplied := args[1]
	if suplied != passphrasestart {
		str := "Bad passphrase, sorry."
		return &str
	}
	id := string(msg.Chat.ID)
	usersallowed[id] = true
	saveInDb("user:"+id, "true")
	bot.Answer(msg).Text("Yay! You are now allowed to use me :-) <3").ReplyToMessage(msg.ID).End()
	return nil
}

func canUser(msg tgbot.Message) bool {
	if passphrasestart == "" {
		return true
	}
	_, ok := usersallowed[string(msg.Chat.ID)]
	return ok
}

func help(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	if !canUser(msg) {
		return nil
	}
	var buffer bytes.Buffer
	orderk := []string{}
	for cmd := range availableCommands {
		orderk = append(orderk, cmd)
	}
	sort.Strings(orderk)
	for _, cmd := range orderk {
		htext := availableCommands[cmd]
		buffer.WriteString(fmt.Sprintf("%s - %s\n", cmd, htext))
	}
	res := buffer.String()
	return &res
}

func getCleanPath(u int, p string) string {
	nbase, ok := userpath[u]
	if !ok {
		nbase = base
	}
	return filepath.Clean(filepath.Join(nbase, p))
}

func getFiles(p string) []os.FileInfo {
	files, _ := ioutil.ReadDir(p)
	nf := []os.FileInfo{}
	for _, f := range files {
		name := filepath.Base(f.Name())
		if !strings.HasPrefix(name, ".") && !strings.HasPrefix(name, "#") {
			nf = append(nf, f)
		}
	}
	return nf
}

func isValidFile(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return !fi.IsDir()
}

func isValidDir(p string) bool {
	fi, err := os.Stat(p)
	if err != nil {
		return false
	}
	return fi.IsDir()
}

func buildSafeSendPath(p string) string {
	return strings.Replace(p, base, "$HOME", 1)
}

func getPathFromInt(uid int, n int) (string, error) {
	nbase, ok := userpath[uid]
	if !ok {
		nbase = base
	}
	files := getFiles(nbase)
	if n > len(files) {
		return "", errors.New("Too big man...")
	}

	return files[n].Name(), nil
}

func cd(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	if !canUser(msg) {
		return nil
	}
	uid := msg.From.ID
	if len(args) == 1 {
		// cd base
		userpath[uid] = base
		str := buildSafeSendPath(base)
		str = fmt.Sprintf("%s:\n---------\n%s", str, buildList(base))
		return &str
	}
	p := args[1]

	i, err := strconv.Atoi(p)
	if err == nil {
		if i < 0 {
			str := "The number is negative... so bad..."
			return &str
		}
		p, err = getPathFromInt(uid, i)
		if err != nil {
			str := err.Error()
			return &str
		}
	}

	p = getCleanPath(uid, p)

	// Check if exist
	// cd p
	if !isValidDir(p) {
		msg := "Path " + buildSafeSendPath(p) + " is not valid (maybe is a file?)"
		return &msg
	}
	userpath[uid] = p
	pc := buildSafeSendPath(p)
	str := fmt.Sprintf("%s:\n---------\n%s", pc, buildList(p))
	newstr := divideAndConquer(str)
	for _, s := range newstr {
		bot.Answer(msg).Text(s).ReplyToMessage(msg.ID).End()
	}
	return nil
}

func ls(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	if !canUser(msg) {
		return nil
	}
	uid := msg.From.ID
	path := base
	if len(args) == 1 {
		// cd base
		p, ok := userpath[uid]
		path = p
		if !ok {
			path = base
			userpath[uid] = base
		}
	} else {
		p := args[1]

		i, err := strconv.Atoi(p)
		if err == nil {
			if i < 0 {
				str := "The number is negative... so bad..."
				return &str
			}
			p, err = getPathFromInt(uid, i)
			if err != nil {
				str := err.Error()
				return &str
			}
		}
		path = getCleanPath(uid, p)
	}
	if !isValidDir(path) {
		str := "Not valid dir: " + buildSafeSendPath(path)
		return &str
	}
	pc := buildSafeSendPath(path)
	str := fmt.Sprintf("%s:\n---------\n%s", pc, buildList(path))

	newstr := divideAndConquer(str)
	for _, s := range newstr {
		bot.Answer(msg).Text(s).ReplyToMessage(msg.ID).End()
	}
	return nil
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

func buildList(path string) string {
	files := getFiles(path)
	var buffer bytes.Buffer
	for i, f := range files {
		name := filepath.Base(f.Name())
		if f.IsDir() {
			buffer.WriteString(fmt.Sprintf("%d) [D] - %s\n", i, name))
		} else {
			buffer.WriteString(fmt.Sprintf("%d) [F] - %s\n", i, name))
		}
	}
	str := buffer.String()
	return str
}

func download(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	if !canUser(msg) {
		return nil
	}
	uid := msg.From.ID
	path := base
	filen := args[1]

	i, err := strconv.Atoi(filen)
	if err == nil {
		if i < 0 {
			str := "The number is negative... so bad..."
			return &str
		}
		filen, err = getPathFromInt(uid, i)
		if err != nil {
			str := err.Error()
			return &str
		}
	}

	path = getCleanPath(uid, filen)

	fid, ok := bookid[path]
	if ok {
		nmsg := bot.Answer(msg).Document(fid).ReplyToMessage(msg.ID).End()
		if nmsg.Ok {
			return nil
		}
	}

	if !isValidFile(path) {
		str := "Not valid file: " + buildSafeSendPath(path)
		return &str
	}

	bot.Answer(msg).Text("Sending file: " + buildSafeSendPath(path)).End()
	bot.Answer(msg).Action(tgbot.UploadDocument).End()
	nmsg := bot.Answer(msg).Document(path).ReplyToMessage(msg.ID).End()
	// fmt.Println(nmsg)
	if nmsg.Ok && nmsg.Result.Document != nil {
		res := *nmsg.Result.Document
		bookid[path] = res.FileID
		saveInDb("book:"+path, res.FileID)
	}
	return nil
}

func saveInDb(k string, v string) {
	db, err := leveldb.OpenFile(dbdir, nil)
	defer db.Close()
	if err != nil {
		return
	}
	err = db.Put([]byte(k), []byte(v), nil)
}

func readAllDb() {
	db, err := leveldb.OpenFile(dbdir, nil)
	defer db.Close()
	if err != nil {
		return
	}
	booksn := 0
	usersn := 0
	bookid = map[string]string{}
	usersallowed = map[string]bool{}
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := string(iter.Key())
		value := string(iter.Value())
		if strings.HasPrefix(key, "book:") {
			bookid[key] = value
			booksn += 1
		} else if strings.HasPrefix(key, "user:") {
			usersallowed[key] = value == "true"
			usersn += 1
		}
	}
	iter.Release()
	fmt.Printf("-----\nDatabase %s loaded!\n------\nBooks: %d\nUsers: %d\n-----\n", dbdir, booksn, usersn)
}

func main() {
	flag.StringVar(&base, "dir", "~/Libros", "working directory")
	flag.StringVar(&dbdir, "db", "./book.db", "database file")
	flag.StringVar(&passphrasestart, "pwd", "", "passphrase to start command")
	flag.Parse()
	ebase, err := homedir.Expand(base)
	if err != nil || ebase == "" {
		fmt.Println("Files path not valid")
		return
	}
	base = ebase
	edbdir, err := homedir.Expand(dbdir)
	if err != nil || edbdir == "" {
		fmt.Println("Database path not valid")
		return
	}
	dbdir = edbdir
	fmt.Println("Base dir: " + base)
	fmt.Println("DataBase dir: " + dbdir)
	if passphrasestart == "" {
		fmt.Println("Your bot is not password protected.")
	} else {
		fmt.Println("Your bot is password protected.\nThe users set the password with:\n/start <pwd>")
	}

	readAllDb()

	godotenv.Load("secrets.env")
	// Add a file secrets.env, with the key like:
	// TELEGRAM_KEY=yourtoken
	token := os.Getenv("TELEGRAM_KEY")
	bot := tgbot.New(token)

	bot.SimpleCommandFn(`help`, help)
	bot.MultiCommandFn([]string{`start`, `start (.+)`}, start)
	bot.MultiCommandFn([]string{`cd`, `cd ([\w/]+)`}, cd)
	bot.MultiCommandFn([]string{`ls`, `ls ([\w/]+)`}, ls)
	bot.MultiCommandFn([]string{`download ([\w/]+)`, `dw ([\w/]+)`}, download)

	bot.SimpleStart()
}
