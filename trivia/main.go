package main

import (
	"math/rand"
	"sync"

	"github.com/rockneurotiko/go-tgbot"
)

type question struct {
	Question string   `json:"question"`
	Answers  []string `json:"answers"`
	Solution int      `json:"solution"`
}

var questions = []question{
	question{"Current year?", []string{"1994", "2015", "2016", "2017"}, 2},
}

type usersAnsweringStruct struct {
	*sync.RWMutex
	Users map[int]int
}

func (users *usersAnsweringStruct) get(user int) (int, bool) {
	users.RLock()
	i, ok := users.Users[user]
	users.RUnlock()
	return i, ok
}

func (users *usersAnsweringStruct) set(user int, value int) {
	users.Lock()
	users.Users[user] = value
	users.Unlock()
}

func (users *usersAnsweringStruct) del(user int) {
	users.Lock()
	delete(users.Users, user)
	users.Unlock()
}

var usersAnswering = usersAnsweringStruct{&sync.RWMutex{}, make(map[int]int)}

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

func questionHandler(bot tgbot.TgBot, msg tgbot.Message, text string) *string {
	r := rand.Intn(len(questions))
	choosen := questions[r]

	keyl := buildKeyboard(choosen.Answers)
	rkm := tgbot.ReplyKeyboardMarkup{
		Keyboard:        keyl,
		ResizeKeyboard:  false,
		OneTimeKeyboard: true,
		Selective:       false,
	}

	usersAnswering.set(msg.Chat.ID, r)

	bot.Answer(msg).Text(choosen.Question).ReplyToMessage(msg.ID).Keyboard(rkm).End()

	return nil
}

func maybeAnswerHandler(bot tgbot.TgBot, msg tgbot.Message) {
	if msg.Text == nil {
		return
	}

	text := *msg.Text

	i, ok := usersAnswering.get(msg.Chat.ID)

	usersAnswering.del(msg.Chat.ID) // We can safely remove right now

	if !ok || i < 0 || i >= len(questions) {
		bot.Answer(msg).Text("You need to start a /question first").End()
		return
	}

	choosen := questions[i]
	goodone := choosen.Answers[choosen.Solution]

	if text == goodone {
		bot.Answer(msg).Text("SUCCESS!").End()
		return
	}

	bot.Answer(msg).Text("WRONG!").End()
}

func main() {
	token := ""

	bot := tgbot.New(token).
		SimpleCommandFn(`question`, questionHandler).
		NotCalledFn(maybeAnswerHandler)

	bot.SimpleStart()
}
