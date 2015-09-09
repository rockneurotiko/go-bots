package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

type UrlInfo struct {
	Url  string
	Name string
}

// youtube!

func scrape_youtube(uri UrlInfo) UrlInfo {
	var m struct {
		Data struct {
			Status  string `json:"status"`
			Content string `json:"content"`
		} `json:"data"`
	}
	var m2 struct {
		Data struct {
			Status  string `json:"status"`
			Content string `json:"content"`
		} `json:"data"`
	}

	basereq := "http://api.debianweb.ir/youtube/%s"
	downreq := "http://api.debianweb.ir/youtube/link/%s/%s"

	u, err := url.Parse(uri.Url)
	if err != nil {
		return uri
	}
	vals := u.Query()
	videoid := vals.Get("v")
	if videoid == "" {
		return uri
	}

	res, err := http.Get(fmt.Sprintf(basereq, videoid))
	if err != nil {
		return uri
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return uri
	}
	json.Unmarshal(bytes, &m)
	if m.Data.Status != "OK" {
		return uri
	}

	listformats := strings.Split(m.Data.Content, "\n")
	// Just pick thi firstone, but we'll have to
	picked := listformats[0]
	fid := strings.Split(picked, "|")[0]

	res, err = http.Get(fmt.Sprintf(downreq, videoid, fid))
	if err != nil {
		return uri
	}

	bytes, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return uri
	}
	json.Unmarshal(bytes, &m2)
	if m2.Data.Status != "OK" {
		return uri
	}

	if m2.Data.Content != "" {
		return UrlInfo{
			strings.TrimRight(m2.Data.Content, "\n"),
			uri.Name,
		}
	}

	return uri
}

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

func scrape_dropbox(uri UrlInfo) UrlInfo {
	return UrlInfo{
		strings.Replace(uri.Url, "dl=0", "raw=1", -1),
		uri.Name,
	}
}

func scrape_uri(uri UrlInfo) UrlInfo {
	checker := func(dom string, pref string) bool {
		return strings.HasPrefix(
			strings.TrimLeft(
				strings.TrimLeft(dom, "m."),
				"www."),
			pref)
	}
	u, err := url.Parse(uri.Url)
	if err != nil {
		return uri
	}
	dom := u.Host
	if checker(dom, "soundcloud.com") {
		return scrape_soundcloud(uri)
	}

	// if checker(dom, "youtube.") {
	// 	return scrape_youtube(uri)
	// }

	// if checker(dom, "youtu.be") {
	// 	uri.Url = fmt.Sprintf("%s?v=%s", uri.Url, strings.TrimLeft(u.Path, "/"))
	// 	return scrape_youtube(uri)
	// }

	if checker(dom, "dropbox.com") {
		return scrape_dropbox(uri)
	}
	return uri
}
