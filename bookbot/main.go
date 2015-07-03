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
// dl - Shourcut to /download :)
// hidekeyboard - Hide the keyboard

var availableCommands = map[string]string{
	"/start":         "Start the bot",
	"/start <pwd>":   "Start the bot, use it when the bot have password",
	"/help":          "Show this help",
	"/cd":            "Change user directory",
	"/cd <id>":       "Change user directory to the specified",
	"/ls":            "Show current user directory content",
	"/ls <id>":       "Show specified directory content",
	"/download <id>": "Download specified file",
	"/dw <id>":       "Shortcut to /download :)",
	"/dl <id>":       "Another shortcut to /download :)",
	"/hidekeyboard":  "Hide the keyboard for you!",
}

var bookid = make(map[string]string)

var userpath = make(map[int]string)
var usersallowed = make(map[string]bool)

var base = ""
var dbdir = ""
var passphrasestart = ""

func buildKeyboard(ops []string) [][]string {
	keylayout := [][]string{{}}
	for _, k := range ops {
		if len(keylayout[len(keylayout)-1]) == 2 {
			keylayout = append(keylayout, []string{k})
		} else {
			keylayout[len(keylayout)-1] = append(keylayout[len(keylayout)-1], k)
		}
	}
	return keylayout
}

func start(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	if len(args) <= 1 || passphrasestart == "" {
		res := buildHelp()
		bot.Answer(msg).Text(res).ReplyToMessage(msg.ID).End()
		return nil
	}

	suplied := args[1]
	if suplied != passphrasestart {
		str := "Bad passphrase, sorry."
		bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
		return nil
	}
	id := fmt.Sprintf("%v", msg.Chat.ID)
	usersallowed[id] = true
	saveInDb("user:"+id, "true")
	bot.Answer(msg).Text("Yay! You are now allowed to use me :-) <3").ReplyToMessage(msg.ID).End()
	return nil
}

func hidekeyboard(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	rkm := tgbot.ReplyKeyboardHide{HideKeyboard: true, Selective: true}
	bot.Answer(msg).Text("Hidden!").KeyboardHide(rkm).End()
	return nil
}

func canUser(msg tgbot.Message) bool {
	if passphrasestart == "" {
		return true
	}

	str := fmt.Sprintf("%v", msg.Chat.ID)
	v, ok := usersallowed[str]
	return ok && v
}

func buildHelp() string {
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
	return buffer.String()
}

func help(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	if !canUser(msg) {
		return nil
	}
	res := buildHelp()
	bot.Answer(msg).Text(res).ReplyToMessage(msg.ID).End()
	return nil
}

func getCleanPath(u int, p string) string {
	nbase, ok := userpath[u]
	if !ok {
		nbase = base
	}
	if p == ".." && nbase == base {
		p = ""
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

func cdhome(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	if !canUser(msg) {
		return nil
	}
	uid := msg.From.ID
	// cd base
	userpath[uid] = base
	str := buildSafeSendPath(base)
	lists, keyb := buildListKey(base)
	str = fmt.Sprintf("%s:\n---------\n%s", str, lists)
	keyl := buildKeyboard(keyb)
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keyl,
		ResizeKeyboard:  false,
		OneTimeKeyboard: true,
		Selective:       true}
	bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).Keyboard(rkm).End()
	return nil
}

func cdother(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	if !canUser(msg) {
		return nil
	}
	uid := msg.From.ID
	p := args[1]

	i, err := strconv.Atoi(p)
	if err == nil {
		if i < 0 {
			str := "The number is negative... so bad..."
			bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
			return nil
		}
		p, err = getPathFromInt(uid, i)
		if err != nil {
			str := err.Error()
			bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
			return nil
		}
	}

	p = getCleanPath(uid, p)

	// Check if exist
	// cd p
	if !isValidDir(p) {
		str := "Path " + buildSafeSendPath(p) + " is not valid (maybe is a file?)"
		bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
		return nil
	}
	userpath[uid] = p
	pc := buildSafeSendPath(p)
	lists, keyb := buildListKey(p)
	str := fmt.Sprintf("%s:\n---------\n%s", pc, lists)
	keyl := buildKeyboard(keyb)
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keyl,
		ResizeKeyboard:  false,
		OneTimeKeyboard: true,
		Selective:       true}

	newstr := divideAndConquer(str)
	for _, s := range newstr {
		bot.Answer(msg).Text(s).ReplyToMessage(msg.ID).Keyboard(rkm).End()
	}
	return nil
}

func lskey(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	if !canUser(msg) {
		return nil
	}
	uid := msg.From.ID
	p, ok := userpath[uid]
	path := p
	if !ok {
		path = base
		userpath[uid] = base
	}

	if !isValidDir(path) {
		str := "Not valid dir: " + buildSafeSendPath(path)
		bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
		return nil
	}
	pc := buildSafeSendPath(path)
	str, keyb := buildListKey(path)
	str = fmt.Sprintf("%s:\n---------\n%s", pc, str)

	keyl := buildKeyboard(keyb)
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keyl,
		ResizeKeyboard:  false,
		OneTimeKeyboard: true,
		Selective:       true}

	newstr := divideAndConquer(str)
	for _, s := range newstr {
		bot.Answer(msg).Text(s).ReplyToMessage(msg.ID).Keyboard(rkm).End()
	}
	return nil
}

func lsother(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	uid := msg.From.ID
	p := args[1]

	i, err := strconv.Atoi(p)
	if err == nil {
		if i < 0 {
			str := "The number is negative... so bad..."
			bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
			return nil
		}
		p, err = getPathFromInt(uid, i)
		if err != nil {
			str := err.Error()
			bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
			return nil
		}
	}
	p = getCleanPath(uid, p)

	if !isValidDir(p) {
		str := "Not valid dir: " + buildSafeSendPath(p)
		bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
		return nil
	}

	pc := buildSafeSendPath(p)
	str := fmt.Sprintf("%s:\n---------\n%s", pc, buildList(p))
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        [][]string{{fmt.Sprintf(`/cd %d`, i)}},
		ResizeKeyboard:  false,
		OneTimeKeyboard: true,
		Selective:       true}
	newstr := divideAndConquer(str)
	for _, s := range newstr {
		bot.Answer(msg).Text(s).ReplyToMessage(msg.ID).Keyboard(rkm).End()
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

func buildListKey(path string) (string, []string) {
	files := getFiles(path)
	kf := []string{}
	var buffer bytes.Buffer
	for i, f := range files {
		name := filepath.Base(f.Name())
		if f.IsDir() {
			buffer.WriteString(fmt.Sprintf("%d) [D] - %s\n", i, name))
			kf = append(kf, fmt.Sprintf("/cd %d (%s)", i, name))
		} else {
			buffer.WriteString(fmt.Sprintf("%d) [F] - %s\n", i, name))
			kf = append(kf, fmt.Sprintf("/dl %d (%s)", i, name))
		}
	}
	if path != base {
		backd := filepath.Clean(filepath.Join(path, ".."))
		kf = append(kf, fmt.Sprintf(`/cd .. (%s)`, backd))
	}
	kf = append(kf, `/hidekeyboard`)

	str := buffer.String()
	return str, kf
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

func downloadkey(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	if !canUser(msg) {
		return nil
	}
	uid := msg.From.ID
	path := base
	filen := args[1]

	i, err := strconv.Atoi(filen)
	if err != nil {
		// Not number, that's impossible (reg expr)
		return nil
	}

	if i < 0 {
		// That's impossible too
		str := "The number is negative... so bad..."
		bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
		return nil
	}
	filen, err = getPathFromInt(uid, i)
	if err != nil {
		str := err.Error()
		bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
		return nil
	}

	path = getCleanPath(uid, filen)

	fid, ok := bookid[path]
	if ok {
		kh := tgbot.ReplyKeyboardHide{true, true}
		nmsg := bot.Answer(msg).Document(fid).ReplyToMessage(msg.ID).KeyboardHide(kh).End()
		if nmsg.Ok {
			return nil
		}
	}

	if !isValidFile(path) {
		str := "Not valid file: " + buildSafeSendPath(path)
		bot.Answer(msg).Text(str).ReplyToMessage(msg.ID).End()
		return nil
	}

	bot.Answer(msg).Text("Sending file: " + buildSafeSendPath(path)).End()
	bot.Answer(msg).Action(tgbot.UploadDocument).End()
	nmsg := bot.Answer(msg).Document(path).ReplyToMessage(msg.ID).End()

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
	bookid = make(map[string]string)
	usersallowed = make(map[string]bool)
	iter := db.NewIterator(nil, nil)
	for iter.Next() {
		// Remember that the contents of the returned slice should not be modified, and
		// only valid until the next call to Next.
		key := string(iter.Key())
		value := string(iter.Value())
		if strings.HasPrefix(key, "book:") {
			bookid[key[5:]] = value
			booksn++
		} else if strings.HasPrefix(key, "user:") {
			usersallowed[key[5:]] = value == "true"
			usersn++
		}
	}
	iter.Release()
	fmt.Printf("-----\nDatabase %s loaded!\n------\nBooks: %d\nUsers: %d\n-----\n", dbdir, booksn, usersn)
}

func main() {
	var envdir string
	flag.StringVar(&base, "dir", "~/Libros", "working directory")
	flag.StringVar(&dbdir, "db", "./book.db", "database file")
	flag.StringVar(&passphrasestart, "pwd", "", "passphrase to start command")
	flag.StringVar(&envdir, "env", "secrets.env", "Environment file (secret.env)")
	flag.Parse()
	ebase, err := homedir.Expand(base)
	if err != nil || ebase == "" || !isValidDir(ebase) {
		fmt.Println("Files path not valid")
		return
	}
	base = ebase

	edbdir, err := homedir.Expand(dbdir)
	if err != nil || edbdir == "" || !isValidDir(edbdir) {
		fmt.Println("Database path not valid")
		return
	}
	dbdir = edbdir

	eenvdir, err := homedir.Expand(envdir)
	if err != nil || eenvdir == "" || !isValidFile(eenvdir) {
		fmt.Println("Environment path not valid")
		return
	}
	envdir = eenvdir

	fmt.Println("Base dir: " + base)
	fmt.Println("DataBase dir: " + dbdir)
	fmt.Println("Environment dir: " + envdir)
	if passphrasestart == "" {
		fmt.Println("Your bot is not password protected.")
	} else {
		fmt.Println("Your bot is password protected.\nThe users set the password with:\n/start <pwd>")
	}

	readAllDb()

	godotenv.Load(envdir)
	// Add a file secrets.env, with the key like:
	// TELEGRAM_KEY=yourtoken
	token := os.Getenv("TELEGRAM_KEY")
	bot := tgbot.New(token)

	bot.SimpleCommandFn(`help`, help)
	bot.SimpleCommandFn(`hidekeyboard`, hidekeyboard)
	bot.SimpleCommandFn(`cd`, cdhome)
	bot.SimpleCommandFn(`ls`, lskey)

	bot.CommandFn(`cd (\d+|..)(?:.*)`, cdother) // cd with number or ..
	bot.CommandFn(`ls (\d+)(?:.*)`, lsother)

	bot.MultiCommandFn([]string{`start`, `start (.+)`}, start)
	bot.MultiCommandFn([]string{
		`download (\d+)(?:.*)`,
		`dw (\d+)(?:.*)`,
		`dl (\d+)(?:.*)`},
		downloadkey)

	bot.SimpleStart()
}
