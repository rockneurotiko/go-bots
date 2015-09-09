package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type UrlInfo struct {
	Url  string
	Name string
}

// youtube!

func scrape_soundcloud(uri UrlInfo) UrlInfo {
	var m struct {
		WaveForm  string `json:"waveform_url"`
		Permalink string `json:"permalink"`
	}
	baseu := "https://www.appendipity.com/scs/scsdata.php?url=%s&callback=?"
	urlbase := "http://media.soundcloud.com/stream/%s"
	namebase := "%s.mp3"

	resp, err := http.Get(fmt.Sprintf(baseu, uri.Url))
	if err != nil {
		return uri
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	body := buf.String()

	nbody := strings.TrimRight(strings.TrimLeft(body, "?()"), ");")

	json.Unmarshal([]byte(nbody), &m)
	splitted := strings.Split(m.WaveForm, "/")
	if len(splitted) > 0 {
		splitted2 := strings.Split(splitted[len(splitted)-1], "_")
		code := splitted2[0]
		return UrlInfo{
			fmt.Sprintf(urlbase, code),
			fmt.Sprintf(namebase, m.Permalink),
		}
	}
	return uri
}

func scrape_uri(uri UrlInfo) UrlInfo {
	u, err := url.Parse(uri.Url)
	if err != nil {
		return uri
	}
	dom := u.Host
	if dom == "soundcloud.com" {
		return scrape_soundcloud(uri)
	}

	return uri
}
