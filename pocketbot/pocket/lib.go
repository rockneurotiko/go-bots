package pocket

import (
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/motemen/go-pocket/api"
	"github.com/motemen/go-pocket/auth"
	"github.com/rockneurotiko/go-tgbot"
	"github.com/syndtr/goleveldb/leveldb"
)

/*
Commands:

help - Shows the help
stop - Stop!!!!
auth - Authenticate me please!
list - Give me my urls!
sync - I've just updated something elsewhere, update yours please
add - Add this URL ;-)
delete - Delete this URL ;-)

*/

type Messages struct {
	NeedAuth string
}

var default_messages = Messages{
	"You need to authenticate first: /auth",
}

type ConfigJ struct {
	Token       string `json:"token"`
	ConsumerKey string `json:"consumer_key"`
	DbPath      string `json:"db_path"`
}

func debug(base string, others ...interface{}) {
	doit := true
	if doit {
		fmt.Printf(base, others...)
	}
}

func (self ConfigJ) buildClient(auth Tokens) *api.Client {
	return api.NewClient(self.ConsumerKey, auth.Authorization.AccessToken)
}

func (self ConfigJ) help(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	fmt.Printf("<%s(%d)> is asking for general help\n", msg.From.FirstName, msg.From.ID)

	authtext := "You are authenticated! You can use all my power!"
	auth, ok := auths.getTokens(msg.From.ID)
	if !ok || auth.Authorization == nil {
		authtext = "Seems that you are not authenticate, use the /auth command to authenticate and make a real use of me :)"
	}

	textans := fmt.Sprintf(`Hi! I'm Pocket Bot! How are you you %s?
I'll help you to see and administrate your saved links of Pocket.

%s

This is the list of commands, you can write /help <command> to see the specific help, for example, type "/help auth" to see the help specific to the auth command:

/help
/stop
/auth
/list
/sync
/add
/delete

This bot has been created by @rockneurotiko, I hope that you like it ;-)`, msg.From.FirstName, authtext)
	bot.Answer(msg).Text(textans).End()
	return nil
}

func (self ConfigJ) specific_help(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	fmt.Printf("<%s(%d)> is asking for specific help of %s\n", msg.From.FirstName, msg.From.ID, args[1])
	bot.Answer(msg).Text(fmt.Sprintf("Sorry, I didn't writed the help to %s yet...", args[1])).End()
	return nil
}

func (self ConfigJ) start(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	fmt.Println(args)
	if len(args) <= 1 {
		auth, ok := auths.getTokens(msg.From.ID)
		if !ok || auth.Authorization == nil {
			self.help(bot, msg, "")
			return nil
		}
		return nil // Error, not parameter :S
	}

	rid := args[1]

	i, e := strconv.Atoi(rid)
	if e != nil {
		// Not a number!
		return nil
	}

	fmt.Printf("<%s(%d)> is asking to authorize %d\n", msg.From.FirstName, msg.From.ID, i)

	if i != msg.From.ID {
		// Other guy is trying to fuck other xD
		return nil
	}

	tokens, ok := auths.getTokens(i)
	fmt.Println(tokens, ok)
	if !ok || tokens.RequestToken == nil {
		// First need to /auth
		bot.Answer(msg).Text(default_messages.NeedAuth).End()
		return nil
	}

	auth, error := auth.ObtainAccessToken(self.ConsumerKey, tokens.RequestToken)

	if error != nil {
		// :(
		return nil
	}

	debug("%+v\n", auth)

	auths.addAuthorization(msg.From.ID, auth)
	bot.Answer(msg).Text("Thanks for authenticating!!").End()
	return nil
}

func (self ConfigJ) stop(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	if len(args) > 1 {
		cmd := args[1]
		if cmd == "list" {
			rkm := tgbot.ReplyKeyboardHide{HideKeyboard: true, Selective: true}
			bot.Answer(msg).Text("Stopped :)").KeyboardHide(rkm).End()
			return nil
		}
	}
	return nil
}

func (self ConfigJ) internalSync(id int, auth Tokens) ([]api.Item, error) {
	client := self.buildClient(auth)
	res, err := client.Retrieve(nil)
	if err != nil {
		return nil, err
	}

	items := []api.Item{}
	for _, item := range res.List {
		items = append(items, item)
	}

	sort.Sort(bySortID(items))

	useritems.replaceUrls(id, &items)

	return items, nil
}

func (self ConfigJ) safeRetrieve(id int, auth Tokens) ([]api.Item, error) {
	url, ok := useritems.getUrls(id)
	if !ok {
		return self.internalSync(id, auth)
	}
	return *url, nil
}

func (self ConfigJ) sync(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	fmt.Printf("<%s(%d)> is asking to sync\n", msg.From.FirstName, msg.From.ID)
	auth, ok := auths.getTokens(msg.From.ID)
	if !ok || auth.Authorization == nil {

		bot.Answer(msg).Text(default_messages.NeedAuth).End()
		return nil
	}

	_, err := self.internalSync(msg.From.ID, auth)

	if err != nil {
		bot.Answer(msg).Text("Something went wrong while syncing.").End()
		return nil
	}
	bot.Answer(msg).Text("Sync completed!!!").End()
	return nil
}

func (self ConfigJ) list(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	getn := "1"
	nitemslast := "0"
	nitems := "0"
	if len(args) > 1 {
		getn = args[1]
	}

	if len(args) > 2 {
		nitems = args[2]
	}

	if len(args) > 3 {
		nitemslast = args[3]
	}

	fmt.Printf("<%s(%d)> is asking him list number %s, from item %s to item %s\n", msg.From.FirstName, msg.From.ID, getn, nitems, nitemslast)

	auth, ok := auths.getTokens(msg.From.ID)
	if !ok || auth.Authorization == nil {
		bot.Answer(msg).Text(default_messages.NeedAuth).End()
		return nil
	}

	i, err := strconv.Atoi(getn)
	if err != nil {
		fmt.Println("Not number")
		return nil
	}
	if i <= 0 {
		fmt.Println("Negative number or zero")
		return nil
	}

	iitems, err := strconv.Atoi(nitems)
	if err != nil {
		fmt.Println("Not number")
		return nil
	}
	if iitems < 0 {
		fmt.Println("Negative number or zero")
		return nil
	}

	iitemslast, err := strconv.Atoi(nitemslast)
	if err != nil {
		fmt.Println("Not number")
		return nil
	}
	if iitemslast < 0 {
		fmt.Println("Negative number or zero")
		return nil
	}

	if iitems > 0 && iitemslast > 0 && iitems > iitemslast {
		bot.Answer(msg).Text("With the range mode of list you can't ask for inverse").End()
		return nil
	}

	items, err := self.safeRetrieve(msg.From.ID, auth)

	if err != nil {
		fmt.Println("Error retrieving")
		return nil
	}

	textsend, listarts := getBlockN(items, i, "")

	rangeinit := 0
	rangeend := iitems

	if iitemslast > 0 {
		rangeinit = iitems - 1
		rangeend = iitemslast
	}

	if rangeend > 0 {
		if rangeinit < 0 {
			rangeinit = 0
		}
		if rangeend >= len(listarts) {
			rangeend = len(listarts) - 1
		}
		if rangeinit <= rangeend {
			textsend = strings.Join(listarts[rangeinit:rangeend], "")
		} else {
			textsend = ""
		}
	}

	if textsend == "" {
		if rangeinit > 0 {
			fmt.Println("Range out of length")
			bot.Answer(msg).Text(fmt.Sprintf("Your range is bigger than the page.\nYou asked for page %d, items from %d to %d, and the page has %d items.", i, iitems, iitemslast, len(listarts))).End()
			return nil
		}
		fmt.Println("Some error")
		return nil
	}

	sending := bot.Answer(msg).Text(textsend).DisablePreview(true)

	if iitems == 0 && iitemslast == 0 {
		kbl := buildListKeyboard(i > 1, len(textsend) > 3800, i)
		rkm := tgbot.ReplyKeyboardMarkup{
			Keyboard:        kbl,
			ResizeKeyboard:  false,
			OneTimeKeyboard: true,
			Selective:       true,
		}
		sending.Keyboard(rkm)
	}

	sending.End()

	return nil
}

func (self ConfigJ) remove(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	id := args[1]

	fmt.Printf("<%s(%d)> is asking to remove %s\n", msg.From.FirstName, msg.From.ID, id)

	auth, ok := auths.getTokens(msg.From.ID)

	if !ok {
		bot.Answer(msg).Text(default_messages.NeedAuth).End()
		return nil
	}

	idint, e := strconv.Atoi(id)

	if e != nil {
		bot.Answer(msg).Text("That's not an id").End()
		return nil
	}

	client := self.buildClient(auth)

	result, error := client.Modify(&api.Action{Action: "delete", ItemID: idint})

	if error != nil {
		return nil
	}

	if len(result.ActionResults) > 0 && !result.ActionResults[0] {
		bot.Answer(msg).Text(fmt.Sprintf("I can't remove ited %d", idint)).End()
		// Error
		return nil
	}

	self.internalSync(msg.From.ID, auth)
	bot.Answer(msg).Text(fmt.Sprintf("The item %d has been removed", idint)).End()
	return nil
}

func (self ConfigJ) add(bot tgbot.TgBot, msg tgbot.Message, args []string, kargs map[string]string) *string {
	urlmaybe := args[1]

	fmt.Printf("<%s(%d)> is asking to add %s\n", msg.From.FirstName, msg.From.ID, urlmaybe)

	auth, ok := auths.getTokens(msg.From.ID)

	if !ok {
		bot.Answer(msg).Text(default_messages.NeedAuth).End()
		return nil
	}

	_, err := url.Parse(urlmaybe)

	if err != nil {
		bot.Answer(msg).Text("You have to add an url: /add <url>.\nExample: /add https://google.com").End()
		return nil
	}

	if (!strings.HasPrefix(urlmaybe, "http://") || !strings.HasPrefix(urlmaybe, "https://")) && !strings.Contains(urlmaybe, "://") {
		urlmaybe = fmt.Sprintf("http://%s", urlmaybe)
	}

	client := self.buildClient(auth)

	err = client.Add(&api.AddOption{URL: urlmaybe})

	if err != nil {
		// Something wrong
		bot.Answer(msg).Text(fmt.Sprintf("Some error happened while adding %s", urlmaybe)).End()
		return nil
	}

	bot.Answer(msg).Text(fmt.Sprintf("Url %s has been added", urlmaybe)).End()
	self.internalSync(msg.From.ID, auth)
	return nil
}

func (self ConfigJ) auth_f(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	fmt.Println(text)

	redirectURL := fmt.Sprintf("https://telegram.me/getpocketbot?start=%d", msg.From.ID)
	requestToken, err := auth.ObtainRequestToken(self.ConsumerKey, redirectURL)

	if err != nil {
		bot.Answer(msg).Text("That's pretty embarrassing, but seems that the Pocket server is down...").End()
		// Some error!!
		fmt.Println(err)
		return nil
	}

	debug("Token: %+v\nError: %+v\n", requestToken, err)

	url := auth.GenerateAuthorizationURL(requestToken, redirectURL)

	debug("%s\n", url)

	auths.addRequestToken(msg.From.ID, requestToken)

	bot.Answer(msg).Text(fmt.Sprintf("Follow the new URL, authorize it, and when you are back to Telegram, press the Start button: \n%s", url)).DisablePreview(true).End()

	return nil
}

func loadUserAuths() {
	allkeys := autodb.loadFromDbPrefix("user:")

	users := make(map[string]UserState)

	for urlkey, val := range allkeys {
		splitted := strings.Split(urlkey, ":")
		if len(splitted) != 3 {
			continue
		}

		id := splitted[1]
		field := splitted[2]

		idint, e := strconv.Atoi(id)
		if e != nil {
			continue
		}

		user, ok := users[id]
		if !ok {
			user = UserState{idint, Tokens{}}
		}
		if field == "id" {
			user.Id = idint
		}
		if field == "accesstoken" {
			uname := ""
			if user.Toks.Authorization != nil {
				uname = user.Toks.Authorization.Username
			}
			user.Toks.Authorization = &auth.Authorization{val, uname}
		}
		if field == "username" {
			atok := ""
			if user.Toks.Authorization != nil {
				atok = user.Toks.Authorization.AccessToken
			}
			user.Toks.Authorization = &auth.Authorization{atok, val}
		}
		if field == "requesttoken" {
			user.Toks.RequestToken = &auth.RequestToken{val}
		}
		users[id] = user
	}

	usersadd := make([]UserState, 0)
	for _, v := range users {
		usersadd = append(usersadd, v)
	}

	groupstates.addNUsers(usersadd)
}

func BuildBot(token string, localconfig ConfigJ, db *leveldb.DB) *tgbot.TgBot {
	autodb = dbSync{
		&sync.Mutex{},
		localconfig.DbPath,
		db,
	}

	loadUserAuths()

	bot, err := tgbot.NewWithError(token)

	for err != nil {
		fmt.Printf("Some error while building the bot: %s\n", err.Error())
		bot, err = tgbot.NewWithError(token)
	}

	fmt.Println("Bot created! Let's go!")

	bot.SimpleCommandFn(`help`, localconfig.help).
		CommandFn(`help (start|stop|list|auth|sync|add|delete|edit|search|tag|(?:un)?archive|(?:un)?favorite)`, localconfig.specific_help).
		MultiCommandFn([]string{`start`, `start (.*)`}, localconfig.start).
		CommandFn(`stop ?(list)?`, localconfig.stop).
		SimpleCommandFn(`auth`, localconfig.auth_f).
		MultiCommandFn([]string{`list`,
		`list (\d+)`,
		`list (\d+) (\d+)`,
		`list (\d+) (\d+) - (\d+)`}, localconfig.list).
		SimpleCommandFn(`sync`, localconfig.sync).
		CommandFn(`add (.+)`, localconfig.add).      // add
		CommandFn(`delete (.+)`, localconfig.remove) // delete
	// edit
	// search
	// tag

	// archive
	// unarchive
	// favorite
	// unfavorite

	// time.Now().UTC()

	return bot
}
