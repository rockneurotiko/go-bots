package downloader

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"gopkg.in/fatih/set.v0"
)

var listvaliddoms []string = []string{}

func get_all_domains_available() {
	s := set.New()
	basereq := fmt.Sprintf("%s/api/extractors", youtubedl)
	res, err := http.Get(basereq)
	if err != nil || res.StatusCode != 200 {
		return
	}
	bytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return
	}

	var m struct {
		Extractors []struct {
			Name    string `json:"name"`
			Working bool   `json:"working"`
		} `json:"extractors"`
	}
	json.Unmarshal(bytes, &m)
	for _, e := range m.Extractors {
		if e.Working {
			sname := strings.Split(e.Name, ":")
			if len(sname) == 0 {
				continue
			}
			name := sname[0]
			s.Add(strings.ToLower(name))
		}
	}
	listvaliddoms = set.StringSlice(s)
}
