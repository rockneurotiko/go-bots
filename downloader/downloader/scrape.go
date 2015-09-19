package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/rockneurotiko/go-tgbot"
)

type UrlInfo struct {
	Url  string
	Name string
}

// youtube!

type YoutubeStruct struct {
	Info struct {
		Acodec            string      `json:"acodec,omitempty"`
		AgeLimit          int         `json:"age_limit,omitempty"`
		Annotations       interface{} `json:"annotations,omitempty"`
		AutomaticCaptions struct {
		} `json:"automatic_captions,omitempty"`
		AverageRating float64     `json:"average_rating,omitempty"`
		Categories    []string    `json:"categories,omitempty"`
		Description   string      `json:"description,omitempty"`
		DislikeCount  int         `json:"dislike_count,omitempty"`
		DisplayID     string      `json:"display_id,omitempty"`
		Duration      int         `json:"duration,omitempty"`
		EndTime       interface{} `json:"end_time,omitempty"`
		Ext           string      `json:"ext,omitempty"`
		Extractor     string      `json:"extractor,omitempty"`
		ExtractorKey  string      `json:"extractor_key,omitempty"`
		Filesize      int         `json:"filesize,omitempty"`
		Format        string      `json:"format,omitempty"`
		FormatID      string      `json:"format_id,omitempty"`
		FormatNote    string      `json:"format_note,omitempty"`
		Formats       []struct {
			Abr         int    `json:"abr,omitempty"`
			Acodec      string `json:"acodec,omitempty"`
			Asr         int    `json:"asr,omitempty"`
			Container   string `json:"container,omitempty"`
			Ext         string `json:"ext,omitempty"`
			Filesize    int    `json:"filesize,omitempty"`
			Format      string `json:"format,omitempty"`
			FormatID    string `json:"format_id,omitempty"`
			FormatNote  string `json:"format_note,omitempty"`
			Fps         int    `json:"fps,omitempty"`
			Height      int    `json:"height,omitempty"`
			HTTPHeaders struct {
				Accept         string `json:"Accept,omitempty"`
				AcceptCharset  string `json:"Accept-Charset,omitempty"`
				AcceptEncoding string `json:"Accept-Encoding,omitempty"`
				AcceptLanguage string `json:"Accept-Language,omitempty"`
				UserAgent      string `json:"User-Agent,omitempty"`
			} `json:"http_headers,omitempty"`
			Preference int    `json:"preference,omitempty"`
			Tbr        int    `json:"tbr,omitempty"`
			URL        string `json:"url,omitempty"`
			Vcodec     string `json:"vcodec,omitempty"`
			Width      int    `json:"width,omitempty"`
		} `json:"formats,omitempty"`
		Fps         interface{} `json:"fps,omitempty"`
		Height      int         `json:"height,omitempty"`
		HTTPHeaders struct {
			Accept         string `json:"Accept,omitempty"`
			AcceptCharset  string `json:"Accept-Charset,omitempty"`
			AcceptEncoding string `json:"Accept-Encoding,omitempty"`
			AcceptLanguage string `json:"Accept-Language,omitempty"`
			UserAgent      string `json:"User-Agent,omitempty"`
		} `json:"http_headers,omitempty"`
		ID                 string      `json:"id,omitempty"`
		IsLive             bool        `json:"is_live,omitempty"`
		LikeCount          int         `json:"like_count,omitempty"`
		PlayerURL          string      `json:"player_url,omitempty"`
		Playlist           interface{} `json:"playlist,omitempty"`
		PlaylistIndex      interface{} `json:"playlist_index,omitempty"`
		RequestedSubtitles interface{} `json:"requested_subtitles,omitempty"`
		StartTime          interface{} `json:"start_time,omitempty"`
		Subtitles          struct {
		} `json:"subtitles,omitempty"`
		Tags       []string    `json:"tags,omitempty"`
		Tbr        interface{} `json:"tbr,omitempty"`
		Thumbnail  string      `json:"thumbnail,omitempty"`
		Thumbnails []struct {
			ID  string `json:"id,omitempty"`
			URL string `json:"url,omitempty"`
		} `json:"thumbnails,omitempty"`
		Title              string `json:"title,omitempty"`
		UploadDate         string `json:"upload_date,omitempty"`
		Uploader           string `json:"uploader,omitempty"`
		UploaderID         string `json:"uploader_id,omitempty"`
		URL                string `json:"url,omitempty"`
		Vcodec             string `json:"vcodec,omitempty"`
		ViewCount          int    `json:"view_count,omitempty"`
		WebpageURL         string `json:"webpage_url,omitempty"`
		WebpageURLBasename string `json:"webpage_url_basename,omitempty"`
		Width              int    `json:"width,omitempty"`
	} `json:"info"`
	URL              string `json:"url,omitempty"`
	YoutubeDlVersion string `json:"youtube-dl.version,omitempty"`
}

type internalBest struct {
	Extension string
	Size      int
	Url       string
	HW        int
}

type YoutubeAnalyze struct {
	Title        string
	AllForbidden bool
	BestVideo    internalBest
	BestAudio    internalBest
}

func get_bestone(yt YoutubeStruct) YoutubeAnalyze {
	yta := YoutubeAnalyze{
		Title:        yt.Info.Title,
		AllForbidden: false,
		BestVideo:    internalBest{"", 999999999999, "", 99999999999},
		BestAudio:    internalBest{"", 999999999999, "", 99999999999},
	}
	compare := func(old int, new int) bool {
		if old < MAX_SIZE && new > old && new < MAX_SIZE {
			return true
		}
		if old > MAX_SIZE {
			return true
		}
		return false
	}

	for _, f := range yt.Info.Formats {
		// tenga tamanho, sea menor que el maximo, y el maximo sea mayor que 50MB
		if f.Filesize > 0 {
			if strings.Contains(f.FormatNote, "audio") {
				// Si es mejor O es vorbis y esta bien :)
				if compare(yta.BestAudio.Size, f.Filesize) ||
					(f.Acodec == "vorbis" && f.Filesize < MAX_SIZE) {
					yta.BestAudio.Size = f.Filesize
					yta.BestAudio.Url = f.URL
					yta.BestAudio.Extension = f.Ext
				}
			} else if !strings.Contains(f.FormatNote, "DASH") {
				if compare(yta.BestVideo.Size, f.Filesize) {
					yta.BestVideo.Size = f.Filesize
					yta.BestVideo.Url = f.URL
					yta.BestVideo.Extension = f.Ext
				}
			}
		} else if !strings.Contains(f.FormatNote, "audio") && f.Height > 0 && f.Width > 0 {
			fi := file_info(f.URL)
			if fi.Size == 0 && yta.BestVideo.Size > MAX_SIZE && f.Height*f.Width < yta.BestVideo.HW {
				yta.BestVideo.HW = f.Height * f.Width
				yta.BestVideo.Url = f.URL
				yta.BestVideo.Extension = f.Ext
			} else if yta.BestVideo.Size > MAX_SIZE || (int(fi.Size) > yta.BestVideo.Size && int(fi.Size) < MAX_SIZE) {
				yta.BestVideo.HW = f.Height * f.Width
				yta.BestVideo.Url = f.URL
				yta.BestVideo.Extension = f.Ext
				yta.BestVideo.Size = int(fi.Size)
			} else if yta.BestVideo.Size > MAX_SIZE { // if we are not there yet!
				yta.BestVideo.HW = f.Height * f.Width
				yta.BestVideo.Url = f.URL
				yta.BestVideo.Extension = f.Ext
				if fi.Size > 0 {
					yta.BestVideo.Size = int(fi.Size)
				}
			}
		} else {
			fi := file_info(f.URL)
			if fi.Size > 0 {
				if yta.BestVideo.Size < MAX_SIZE {
					if fi.Size < MAX_SIZE && int(fi.Size) > yta.BestVideo.Size {
						yta.BestVideo.Size = int(fi.Size)
						yta.BestVideo.Url = f.URL
						yta.BestVideo.Extension = f.Ext
					}
				} else {
					yta.BestVideo.Size = int(fi.Size)
					yta.BestVideo.Url = f.URL
					yta.BestVideo.Extension = f.Ext
				}
			}
		}
	}
	return yta
}

type TypeMedia int

const (
	Video TypeMedia = iota
	Audio
)

func (self TypeMedia) WithUrl(url string) string {
	kind := "video"
	if self == Audio {
		kind = "audio"
	}
	return fmt.Sprintf("%s::%s", url, kind)
}

func generic_yt_dl(uri UrlInfo, tm TypeMedia) UrlInfo {
	basereq := fmt.Sprintf("%s/api/info?url=%s", youtubedl, "%s")
	res, err := http.Get(fmt.Sprintf(basereq, uri.Url))
	if err != nil || res.StatusCode != 200 {
		return uri
	}
	bytes, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return uri
	}

	var m YoutubeStruct
	json.Unmarshal(bytes, &m)

	yta := get_bestone(m)
	if tm == Video && yta.BestVideo.Url != "" {
		newname := fmt.Sprintf("%s.%s", yta.Title, yta.BestVideo.Extension)
		return UrlInfo{yta.BestVideo.Url, newname}
	} else if tm == Audio && yta.BestAudio.Url != "" {
		newname := fmt.Sprintf("%s.%s", yta.Title, yta.BestAudio.Extension)
		return UrlInfo{yta.BestAudio.Url, newname}
	} else if yta.BestVideo.Url != "" {
		newname := fmt.Sprintf("%s.%s", yta.Title, yta.BestVideo.Extension)
		return UrlInfo{yta.BestVideo.Url, newname}
	} else if yta.BestAudio.Url != "" {
		newname := fmt.Sprintf("%s.%s", yta.Title, yta.BestAudio.Extension)
		return UrlInfo{yta.BestAudio.Url, newname}
	}
	return uri
}

// ---------------------
type Autogenerated struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Content struct {
			Uploader    string `json:"uploader"`
			UploaderID  string `json:"uploader_id"`
			UploadDate  string `json:"upload_date"`
			Description string `json:"description"`
			Thumbnail   string `json:"thumbnail"`
			Title       string `json:"title"`
			ID          string `json:"id"`
			Formats     []struct {
				FormatID string `json:"format_id"`
				Format   string `json:"format"`
				Ext      string `json:"ext"`
				Width    int    `json:"width,omitempty"`
				Height   int    `json:"height,omitempty"`
				Filesize int    `json:"filesize,omitempty"`
				URL      string `json:"url"`
				ShortURL string `json:"short_url"`
			} `json:"formats"`
		} `json:"content"`
		Status string `json:"status"`
	} `json:"data"`
}

func get_bestone_yt(yt Autogenerated) YoutubeAnalyze {
	yta := YoutubeAnalyze{
		Title:        yt.Data.Content.Title,
		AllForbidden: true,
		BestVideo:    internalBest{"", 999999999999, "", 99999999999},
		BestAudio:    internalBest{"", 999999999999, "", 99999999999},
	}

	compare := func(old int, new int) bool {
		if old < MAX_SIZE && new > old && new < MAX_SIZE {
			return true
		}
		if old > MAX_SIZE {
			return true
		}
		return false
	}
	for _, f := range yt.Data.Content.Formats {
		// tenga tamanho, sea menor que el maximo, y el maximo sea mayor que 50MB
		if f.Filesize > 0 {
			if strings.Contains(f.Format, "audio") {
				// Si es mejor O es vorbis y esta bien :)
				if compare(yta.BestAudio.Size, f.Filesize) {
					yta.BestAudio.Size = f.Filesize
					yta.BestAudio.Url = f.URL
					yta.BestAudio.Extension = f.Ext
				}
			} else if !strings.Contains(f.Format, "DASH") {
				if compare(yta.BestVideo.Size, f.Filesize) {
					yta.BestVideo.Size = f.Filesize
					yta.BestVideo.Url = f.URL
					yta.BestVideo.Extension = f.Ext
				}
			}
		} else if !strings.Contains(f.Format, "audio") && f.Height > 0 && f.Width > 0 {
			fi := file_info(f.URL)
			if fi.Size > 0 && fi.Name != "" {
				yta.AllForbidden = false
			}
			if fi.Size == 0 && yta.BestVideo.Size > MAX_SIZE && f.Height*f.Width < yta.BestVideo.HW {
				yta.BestVideo.HW = f.Height * f.Width
				yta.BestVideo.Url = f.URL
				yta.BestVideo.Extension = f.Ext
			} else if yta.BestVideo.Size > MAX_SIZE || (int(fi.Size) > yta.BestVideo.Size && int(fi.Size) < MAX_SIZE) {
				yta.BestVideo.HW = f.Height * f.Width
				yta.BestVideo.Url = f.URL
				yta.BestVideo.Extension = f.Ext
				yta.BestVideo.Size = int(fi.Size)
			} else if yta.BestVideo.Size > MAX_SIZE { // if we are not there yet!
				yta.BestVideo.HW = f.Height * f.Width
				yta.BestVideo.Url = f.URL
				yta.BestVideo.Extension = f.Ext
				if fi.Size > 0 {
					yta.BestVideo.Size = int(fi.Size)
				}
			}
		}
	}
	return yta
}

func yt(uri UrlInfo, kind TypeMedia) UrlInfo {
	basereq := "http://api.debianweb.ir/yt/%s"

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
	byteArray, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return uri
	}

	var m Autogenerated
	json.Unmarshal(byteArray, &m)
	if m.Code != 200 || m.Data.Status != "OK" {
		log.Println("Fallback to my server!")
		return generic_yt_dl(uri, kind) // fallback!
	}

	yta := get_bestone_yt(m)

	if yta.AllForbidden {
		return generic_yt_dl(uri, kind)
	}

	if kind == Video && yta.BestVideo.Url != "" {
		newname := fmt.Sprintf("%s.%s", yta.Title, yta.BestVideo.Extension)
		return UrlInfo{yta.BestVideo.Url, newname}
	} else if kind == Audio && yta.BestAudio.Url != "" {
		newname := fmt.Sprintf("%s.%s", yta.Title, yta.BestAudio.Extension)
		return UrlInfo{yta.BestAudio.Url, newname}
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
	resp.Body.Close()
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

func scrape_uri(uri UrlInfo, kind TypeMedia, bot tgbot.TgBot, id int) UrlInfo {
	advice := func() {
		bot.Send(id).Text(`I'm going to analyze the URL and try to extract the information you want, please wait :)`).End()
	}
	checker := func(dom string, pref string) bool {
		get_three := func(d string) (string, string, string) {
			s := strings.Split(d, ".")
			if len(s) == 1 {
				return "*", s[0], "*"
			}
			if len(s) == 2 {
				return "*", s[0], s[1]
			}
			if len(s) >= 3 {
				return s[0], s[1], s[2]
			}
			return "*", "*", "*"
		}
		subg, domg, extg := get_three(dom)
		subc, domc, extc := get_three(pref)

		return (subg == subc || subc == "*") &&
			(domg == domc || domc == "*") &&
			(extg == extc || extc == "*")
	}

	u, err := url.Parse(uri.Url)
	if err != nil {
		return uri
	}

	dom := u.Host
	if checker(dom, "dropbox") {
		return scrape_dropbox(uri)
	}

	// if checker(dom, "soundcloud") {
	// 	return scrape_soundcloud(uri)
	// }

	if checker(dom, "youtube") {
		go advice()
		return yt(uri, kind)
	}

	if strings.HasSuffix(dom, ".be") && checker(dom, "youtu") {
		go advice()
		uri.Url = fmt.Sprintf("%s?v=%s", uri.Url, strings.TrimLeft(u.Path, "/"))
		return yt(uri, kind)
	}

	for _, dc := range listvaliddoms {
		if checker(dom, dc) {
			go advice()
			return generic_yt_dl(uri, kind)
		}
	}

	return uri
}
