package rssbot

import (
	"fmt"
	"sync"
	"time"

	"gopkg.in/fatih/set.v0"

	"github.com/pmylund/go-cache"
)

// Default time one day, clean every 5 minutes
// upgrades: Store users too
// var rsscache = cache.New(60*time.Minute, 5*time.Minute)
// var cachelock = &sync.RWMutex{}

type rssCache struct {
	*sync.RWMutex
	Cache *cache.Cache
}

var autorsscache = rssCache{
	&sync.RWMutex{},
	cache.New(60*time.Minute, 5*time.Minute),
}

func setRssValue(key string, value interface{}) {
	autorsscache.Lock()
	autorsscache.Cache.Set(key, value, cache.DefaultExpiration)
	autorsscache.Unlock()
}

func getRssValue(key string) (interface{}, bool) {
	autorsscache.RLock()
	defer autorsscache.RUnlock()
	return autorsscache.Cache.Get(key)
}

func delRssKey(key string) {
	autorsscache.Lock()
	defer autorsscache.Unlock()
	autorsscache.Cache.Delete(key)
}

func addRssInnerUser(key string, id int) {
	autorsscache.Lock()
	defer autorsscache.Unlock()
	if x, found := autorsscache.Cache.Get(key); found {
		switch val := x.(type) {
		case *set.Set:
			val.Add(id)
		default:
			fmt.Println("Error rsscache: ", key, val)
		}
	}
}

func removeRssInnerUser(key string, id int) {
	autorsscache.Lock()
	defer autorsscache.Unlock()
	if x, found := autorsscache.Cache.Get(key); found {
		switch val := x.(type) {
		case *set.Set:
			val.Remove(id)
		default:
			fmt.Println("Error rsscache: ", key, val)
		}
	}
}
