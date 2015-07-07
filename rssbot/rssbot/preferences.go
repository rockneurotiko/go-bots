package rssbot

import (
	"fmt"
	"strings"
	"sync"

	"github.com/rockneurotiko/go-tgbot"
)

var (
	preferencesNameDescr = map[string]prefname{
		"image": prefname{
			"Send images.",
			"If enabled, the bot will send you the images of a new.",
			[][]string{{"/enable", "/disable"}},
		},
	}
	settinglock     = &sync.Mutex{}
	settingchanging = make(map[int]string)
)

type prefname struct {
	Name        string
	Description string
	Keyboard    [][]string
}

func keyboardFromPreference(pr prefname) tgbot.ReplyKeyboardMarkup {
	keylayout := pr.Keyboard
	return tgbot.ReplyKeyboardMarkup{
		Keyboard:        keylayout,
		ResizeKeyboard:  false,
		OneTimeKeyboard: true,
		Selective:       true,
	}
}

func textFromKeyboard(pr prefname) string {
	opst := ""
	for _, v := range pr.Keyboard {
		for _, v2 := range v {
			opst = opst + "\n" + v2
		}
	}
	return opst
}

func textFromOption(pr prefname) string {
	opts := textFromKeyboard(pr)
	return fmt.Sprintf(`%s
%s
Options: %s
/cancel`, pr.Name, pr.Description, opts)
}

func changePreferenceTo(id int, to string) bool {
	settinglock.Lock()
	value := settingchanging[id]
	delete(settingchanging, id)
	settinglock.Unlock()
	pr := preferencesNameDescr[value]
	valid := false
first:
	for _, v := range pr.Keyboard {
		for _, v2 := range v {
			if strings.HasSuffix(v2, to) {
				valid = true
				break first
			}
		}
	}
	if !valid {
		return false
	}

	newopt := transformSetting(id, value, to)
	saveSettingsToUser(id, newopt)
	return true
}

func transformSetting(i int, value string, new string) optionsUser {
	sets := loadSettingsFromUser(i)
	switch value {
	case "image":
		sets.SendImage = new == "enable"
	default:
		return sets
	}
	return sets
}

func getPreference(id int) string {
	settinglock.Lock()
	value := settingchanging[id]
	settinglock.Unlock() // faster do this than defer
	return value
}

func setPreference(id int, value string) {
	settinglock.Lock()
	settingchanging[id] = value
	settinglock.Unlock()
}

func deletePreference(id int) {
	settinglock.Lock()
	delete(settingchanging, id)
	settinglock.Unlock()
}
