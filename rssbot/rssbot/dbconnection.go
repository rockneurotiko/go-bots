package rssbot

import (
	"fmt"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

// DB keys like:
// user:<id>
// user:<id>:<url>
// settings:<id>:<prop>
// rss:<url>
// rss:<url>:<id>

type dbSync struct {
	*sync.Mutex
	Path string
}

var autodb dbSync

func loadFromDb(k string) string {
	autodb.Lock()
	defer autodb.Unlock()
	db, err := leveldb.OpenFile(autodb.Path, nil)
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

func getNthDb(p string, n int) (string, error) {
	full := loadFromDbPrefix(p)
	list := getKeysAndSort(full)

	if n < 0 || n >= len(list) {
		return "", fmt.Errorf("The number %d is not valid", n)
	}
	return list[n], nil
	// autodb.Lock()
	// defer autodb.Unlock()
	// db, err := leveldb.OpenFile(autodb.Path, nil)
	// defer db.Close()
	// if err != nil {
	// 	fmt.Println("Wut")
	// 	return "", "", err
	// }
	// i := 0
	// iter := db.NewIterator(util.BytesPrefix([]byte(p)), nil)
	// for iter.Next() {
	// 	// Use key/value.
	// 	if i == n {
	// 		k := string(iter.Key())
	// 		v := string(iter.Value())
	// 		iter.Release()
	// 		return k, v, nil
	// 	}
	// 	i++
	// }
	// iter.Release()
	// return "", "", fmt.Errorf("The number %d is not valid", n)
}

func loadFromDbPrefix(p string) map[string]string {
	autodb.Lock()
	defer autodb.Unlock()
	res := make(map[string]string, 0)
	db, err := leveldb.OpenFile(autodb.Path, nil)
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

func saveInDb(mult map[string]string) {
	autodb.Lock()
	defer autodb.Unlock()
	db, err := leveldb.OpenFile(autodb.Path, nil)
	defer db.Close()
	if err != nil {
		return
	}
	for k, v := range mult {
		err = db.Put([]byte(k), []byte(v), nil)
	}
}

func deleteFromDb(k string) bool {
	autodb.Lock()
	defer autodb.Unlock()
	db, err := leveldb.OpenFile(autodb.Path, nil)
	defer db.Close()
	if err != nil {
		return false
	}
	err = db.Delete([]byte(k), nil)
	return err == nil
}

func multiDeleteDb(ks []string) {
	autodb.Lock()
	defer autodb.Unlock()
	db, err := leveldb.OpenFile(autodb.Path, nil)
	defer db.Close()
	if err != nil {
		return
	}
	for _, k := range ks {
		db.Delete([]byte(k), nil)
	}
	return
}
