package rssbot

import (
	"fmt"
	"sync"
)

// Properties for settings:
// Send or not images.

var opscache = make(map[int]optionsUser)
var opslock = &sync.RWMutex{}

type optionsUser struct {
	SendImage bool
}

func saveSettingsToUser(i int, nops optionsUser) {
	opslock.Lock()
	opscache[i] = nops
	opslock.Unlock()
	id := fmt.Sprintf("%d", i)
	opsmap := make(map[string]string)
	basek := buildKey("settings", id, "")

	opsmap[buildKey(basek, "send-image", "")] = fmt.Sprintf("%v", nops.SendImage)
	saveInDb(opsmap)
}

func loadSettingsFromUser(i int) optionsUser {
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
	finalops := optionsUser{
		SendImage: opsmap["send-image"] == "true",
	}
	return finalops
}
