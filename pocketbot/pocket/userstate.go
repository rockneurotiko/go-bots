package pocket

import (
	"fmt"
	"sync"
)

type GroupUserState struct {
	*sync.RWMutex
	Users map[int]UserState
}

var groupstates = GroupUserState{
	&sync.RWMutex{},
	make(map[int]UserState),
}

func (self GroupUserState) getUser(i int) (UserState, bool) {
	self.RLock()
	defer self.RUnlock()
	u, ok := self.Users[i]
	return u, ok
}

func (self *GroupUserState) addUser(i int, user UserState) {
	self.Lock()
	defer self.Unlock()
	self.Users[i] = user
}

func (self *GroupUserState) addNUsers(users []UserState) {
	self.Lock()
	defer self.Unlock()
	for _, v := range users {
		self.Users[v.Id] = v
	}
}

type UserState struct {
	Id   int
	Toks Tokens
}

func (self UserState) deleteDb() {
	i := fmt.Sprintf("%d", self.Id)

	basekey := fmt.Sprintf("user:%s:%s", i, "%s")
	idkey := fmt.Sprintf(basekey, "id")
	atokenkey := fmt.Sprintf(basekey, "accesstoken")
	unamekey := fmt.Sprintf(basekey, "username")
	rtokenkey := fmt.Sprintf(basekey, "requesttoken")

	autodb.multiDeleteDb([]string{idkey, atokenkey, unamekey, rtokenkey})
}

func (self UserState) saveInDb() {
	i := fmt.Sprintf("%d", self.Id)
	requesttoken := ""
	accesstoken := ""
	username := ""
	if self.Toks.RequestToken != nil {
		requesttoken = self.Toks.RequestToken.Code
	}
	if self.Toks.Authorization != nil {
		accesstoken = self.Toks.Authorization.AccessToken
		username = self.Toks.Authorization.Username
	}

	basekey := fmt.Sprintf("user:%s:%s", i, "%s")
	idkey := fmt.Sprintf(basekey, "id")
	atokenkey := fmt.Sprintf(basekey, "accesstoken")
	unamekey := fmt.Sprintf(basekey, "username")
	rtokenkey := fmt.Sprintf(basekey, "requesttoken")

	autodb.saveInDb(map[string]string{
		idkey:     i,
		atokenkey: accesstoken,
		unamekey:  username,
		rtokenkey: requesttoken,
	})
}
