package pocket

import (
	"bytes"
	"fmt"
	"sync"
	"text/template"

	"github.com/motemen/go-pocket/api"
	"github.com/motemen/go-pocket/auth"
)

var defaultItemTemplate = template.Must(template.New("item").Parse(
	`[{{.ItemID | printf "%9d"}}] {{.Title}} <{{.URL}}>
------------------------------
`, // explicit return
))

type bySortID []api.Item

func (s bySortID) Len() int           { return len(s) }
func (s bySortID) Less(i, j int) bool { return s[i].SortId < s[j].SortId }
func (s bySortID) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

type Urls struct {
	*sync.RWMutex
	Data map[int][]api.Item
}

func (self *Urls) getUrls(id int) (*[]api.Item, bool) {
	self.RLock()
	defer self.RUnlock()
	d, ok := self.Data[id]
	return &d, ok
}

func (self *Urls) replaceUrls(id int, urls *[]api.Item) {
	self.Lock()
	defer self.Unlock()
	self.Data[id] = *urls
}

var useritems = &Urls{&sync.RWMutex{}, make(map[int][]api.Item)}

type Auths struct {
	*sync.RWMutex
	Ids map[int]Tokens
}

type Tokens struct {
	RequestToken  *auth.RequestToken
	Authorization *auth.Authorization
}

func (self Auths) getTokens(id int) (Tokens, bool) {
	u, ok := groupstates.getUser(id)
	return u.Toks, ok

	// self.RLock()
	// defer self.RUnlock()
	// t, o := self.Ids[id]
	// return t, o
}

func (self *Auths) addAuthorization(id int, rt *auth.Authorization) {
	u, ok := groupstates.getUser(id)
	if !ok {
		return
	}

	u.Toks.Authorization = rt
	groupstates.addUser(id, u)

	u.saveInDb()
	// self.Lock()
	// defer self.Unlock()
	// toks, ok := self.Ids[id]
	// if !ok {
	// 	// something bad!
	// 	return
	// }
	// self.Ids[id] = Tokens{toks.RequestToken, rt}

}

func (self *Auths) addRequestToken(id int, rt *auth.RequestToken) {
	u, ok := groupstates.getUser(id)
	if !ok {
		u = UserState{id, Tokens{
			RequestToken:  rt,
			Authorization: nil,
		}}
	}

	u.Toks.RequestToken = rt
	groupstates.addUser(id, u)

	u.saveInDb()

	// self.Lock()
	// defer self.Unlock()
	// toks, ok := self.Ids[id]
	// if !ok {
	// 	self.Ids[id] = Tokens{rt, nil}
	// } else {
	// 	self.Ids[id] = Tokens{rt, toks.Authorization}
	// }
}

var auths = Auths{&sync.RWMutex{}, make(map[int]Tokens, 0)}

func getBlockN(items []api.Item, n int, initial string) (string, []string) {
	w := initial
	liststosend := make([]string, 0)

	count := 1
	count2 := 1
	for _, item := range items {
		interm := new(bytes.Buffer)
		err := defaultItemTemplate.Execute(interm, item)
		if err != nil {
			fmt.Println("Error executing")
			// WTF
			return "", []string{}
		}
		temp := interm.String()

		if len(w+temp) >= 4096 {
			if count >= n {
				return w, liststosend
			} else {
				liststosend = []string{initial + temp}
				w = initial + temp // reset string with the new one
				count++
			}
		} else {
			liststosend = append(liststosend, temp)
			w = w + temp
		}

		// if nitems > 0 && nitems == count2 && count >= n {
		// 	return w
		// }

		count2++
	}

	return w, liststosend
}

func buildListKeyboard(prev bool, next bool, actual int) [][]string {
	res := make([]string, 0)
	if prev && actual > 1 {
		res = append(res, fmt.Sprintf("/list %d", actual-1))
	}
	res = append(res, "/stop list")
	if next {
		res = append(res, fmt.Sprintf("/list %d", actual+1))
	}
	return [][]string{res}
}
