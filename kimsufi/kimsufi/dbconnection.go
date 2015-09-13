package kimsufi

import (
	"fmt"
	"sort"
	"sync"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
)

type dbSync struct {
	*sync.Mutex
	Db *leveldb.DB
}

var autodb dbSync

func (self dbSync) loadFromDb(k string) string {
	data, err := self.Db.Get([]byte(k), nil)
	if err != nil {
		return ""
	}
	return string(data)
}

func getKeysAndSort(dic map[string]string) []string {
	var res []string
	for k := range dic {
		res = append(res, k)
	}
	sort.Strings(res)
	return res
}

func getNthDb(p string, n int) (string, error) {
	full := autodb.loadFromDbPrefix(p)
	list := getKeysAndSort(full)

	if n < 0 || n >= len(list) {
		return "", fmt.Errorf("The number %d is not valid", n)
	}
	return list[n], nil
}

func (self dbSync) loadFromDbPrefix(p string) map[string]string {
	res := make(map[string]string, 0)
	iter := self.Db.NewIterator(util.BytesPrefix([]byte(p)), nil)
	for iter.Next() {
		// Use key/value.
		res[string(iter.Key())] = string(iter.Value())
	}
	iter.Release()
	return res
}

func (self dbSync) saveInDb(mult map[string]string) {
	batch := new(leveldb.Batch)
	for k, v := range mult {
		batch.Put([]byte(k), []byte(v))
	}
	self.Db.Write(batch, nil)

}

func (self dbSync) deleteFromDb(k string) bool {
	err := self.Db.Delete([]byte(k), nil)
	return err == nil
}

func (self dbSync) multiDeleteDb(ks []string) {
	batch := new(leveldb.Batch)
	for _, k := range ks {
		batch.Delete([]byte(k))
	}
	self.Db.Write(batch, nil)
}
