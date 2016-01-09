package downloader

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/rockneurotiko/go-tgbot"
	"github.com/rockneurotiko/gorequest"
)

type GDrive struct {
	Reg           *regexp.Regexp
	To            string
	ValidFormats  []string
	DefaultFormat string
}

var (
	drivecheck = GDrive{
		// Reg:           regexp.MustCompile(`https://drive.google.com/file/d/([a-zA-Z0-9-_]+)/edit?usp=sharing`),
		Reg:           regexp.MustCompile(`https:\/\/drive\.google\.com\/file\/d\/([a-zA-Z0-9-_]+)\/(edit|view)\?usp=sharing`),
		To:            "https://drive.google.com/uc?export=download&id=%s",
		ValidFormats:  []string{""},
		DefaultFormat: "",
	}

	doccheck = GDrive{
		Reg:           regexp.MustCompile(`https:\/\/docs\.google\.com\/document\/d\/([a-zA-Z0-9-_]+)\/edit\?usp=sharing`),
		To:            "https://docs.google.com/document/d/%s/export?format=%s",
		ValidFormats:  []string{"doc", "pdf", "txt", "html", "odt"},
		DefaultFormat: "doc",
	}

	presentationcheck = GDrive{
		Reg:           regexp.MustCompile(`https:\/\/docs\.google\.com\/presentation\/d\/([a-zA-Z0-9-_]+)\/edit\?usp=sharing`),
		To:            "https://docs.google.com/presentation/d/%s/export/%s",
		ValidFormats:  []string{"pptx", "pdf"},
		DefaultFormat: "pptx",
	}

	spreadsheetcheck = GDrive{
		Reg:           regexp.MustCompile(`https:\/\/docs\.google\.com\/spreadsheets\/d\/([a-zA-Z0-9-_]+)\/edit\?usp=sharing`),
		To:            "https://docs.google.com/spreadsheets/d/%s/export?format=%s",
		ValidFormats:  []string{"xlsx", "pdf"},
		DefaultFormat: "xlsx",
	}
	drivenoopen = regexp.MustCompile(`https://drive.google.com/open?id=([a-zA-Z0-9-_]+)`)

	// drivereg    = regexp.MustCompile(`https://drive.google.com/file/d/([a-zA-Z0-9-_]+)/edit?usp=sharing`)
	// driveto     = "https://drive.google.com/uc?export=download&id=%s"

	// docreg = regexp.MustCompile(`https://docs.google.com/document/d/([a-zA-Z0-9-_]+)/edit?usp=sharing`)
	// docto  = "https://docs.google.com/document/d/%s/export?format=%s" // doc | pdf | txt | html | odt

	// presentationreg = regexp.MustCompile(`https://docs.google.com/presentation/d/([a-zA-Z0-9-_]+)/edit?usp=sharing`)
	// presentationto  = "https://docs.google.com/presentation/d/%s/export/%s" // pptx | pdf

	// spreadsheetreg = regexp.MustCompile(`https://docs.google.com/spreadsheets/d/([a-zA-Z0-9-_]+)/edit?usp=sharing`)
	// spreadsheetto  = "https://docs.google.com/spreadsheets/d/%s/export?format=%s" // xlsx | pdf
)

type UrlInfo struct {
	Url     string
	Name    string
	Format  string
	Cookies []*http.Cookie
	Error   string
	Kind    TypeMedia
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
	Image
	Document
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
	// fmt.Println(fmt.Sprintf(basereq, uri.Url))
	res, err := http.Get(fmt.Sprintf(basereq, uri.Url))
	if err != nil || res.StatusCode != 200 {
		fmt.Println(err)
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
		return UrlInfo{Url: yta.BestVideo.Url, Name: newname, Kind: Video}
	} else if tm == Audio && yta.BestAudio.Url != "" {
		newname := fmt.Sprintf("%s.%s", yta.Title, yta.BestAudio.Extension)
		return UrlInfo{Url: yta.BestAudio.Url, Name: newname, Kind: Audio}
	} else if yta.BestVideo.Url != "" {
		newname := fmt.Sprintf("%s.%s", yta.Title, yta.BestVideo.Extension)
		return UrlInfo{Url: yta.BestVideo.Url, Name: newname, Kind: Video}
	} else if yta.BestAudio.Url != "" {
		newname := fmt.Sprintf("%s.%s", yta.Title, yta.BestAudio.Extension)
		return UrlInfo{Url: yta.BestAudio.Url, Name: newname, Kind: Audio}
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
		return generic_yt_dl(uri, kind) // fallback!
	}

	vals := u.Query()
	videoid := vals.Get("v")
	if videoid == "" {
		return generic_yt_dl(uri, kind) // fallback!
	}

	res, err := http.Get(fmt.Sprintf(basereq, videoid))
	if err != nil {
		return generic_yt_dl(uri, kind) // fallback!
	}
	byteArray, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return generic_yt_dl(uri, kind) // fallback!
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
		return UrlInfo{Url: yta.BestVideo.Url, Name: newname, Kind: Video}
	} else if kind == Audio && yta.BestAudio.Url != "" {
		newname := fmt.Sprintf("%s.%s", yta.Title, yta.BestAudio.Extension)
		return UrlInfo{Url: yta.BestAudio.Url, Name: newname, Kind: Audio}
	}
	return uri
}

func other_soundcloud(uri UrlInfo, kind TypeMedia) UrlInfo {
	request := gorequest.New().
                Post("http://soundflush.com/").
                Send(struct{
                TrackUrl string `json:"track_url"`
                BtnDl string `json:"btn_download"`
        }{ uri.Url, "Download", })
	request.TargetType = "form"
	_, body, err := request.End()
	
	if len(err) > 0 {
		return new_soundcloud(uri, kind)
	}

	reg := regexp.MustCompile(`<a .*(?:download=".+"|href="([^"]+)").*(?:download=".+"|href="([^"]+)").*>`)
	allu := reg.FindStringSubmatch(body)
        if len(allu) <= 2 {
                return new_soundcloud(uri, kind)
        }
	uri.Url = allu[2]
	return uri
}

func new_soundcloud(uri UrlInfo, kind TypeMedia) UrlInfo {
	// csrftoken := "opHVmkUdlIFSBRrkOIc6O5bNBvYIxnwZ"
	// baseu := "http://9soundclouddownloader.com/download-sound-track?csrfmiddlewaretoken=%s&sound-url=%s"
	baseu := "http://9soundclouddownloader.com/download-sound-track?sound-url=%s?ref=chrome-soundcloud-downloader"
	// gorequest.New().
	// 	Get(fmt.Sprintf(baseu, csrftoken, uri.Url)).
	// 	AddCookies([]*http.Cookie{
	// 	&http.Cookie{
	// 		Name: "_atuvc",
	// 		Value: "16|42",
	// 	},
	// 	&http.Cookie{
        //                 Name: "_atuvs",
        //                 Value: "562b78cd7543e787005",
        //         },
	// 	&http.Cookie{
        //                 Name: "csrftoken",
        //                 Value: csrftoken,
        //         },
	// }).End()
	// resp, err := http.Get(fmt.Sprintf(baseu, csrftoken, uri.Url))
	resp, err := http.Get(fmt.Sprintf(baseu, uri.Url))
	if err != nil {
		return scrape_soundcloud(uri, kind)
	}

	buf := new(bytes.Buffer)
        buf.ReadFrom(resp.Body)
        resp.Body.Close()
        body := buf.String()

	reg := regexp.MustCompile(`<a rel="nofollow" href="([^"]+)"`)
	allu := reg.FindStringSubmatch(body)
	if len(allu) <= 1 {
		return scrape_soundcloud(uri, kind)
	}
	uri.Url = strings.Replace(allu[1], "amp;", "", -1)
	// uri.Url = allu[1]
	return uri
}

func scrape_soundcloud(uri UrlInfo, kind TypeMedia) UrlInfo {
	var m struct {
		WaveForm  string `json:"waveform_url"`
		Permalink string `json:"permalink"`
	}
	baseu := "https://www.appendipity.com/scs/scsdata.php?url=%s&callback=?"
	urlbase := "http://media.soundcloud.com/stream/%s"
	namebase := "%s.mp3"

	resp, err := http.Get(fmt.Sprintf(baseu, uri.Url))
	if err != nil {
		return generic_yt_dl(uri, kind)
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
			Url:  fmt.Sprintf(urlbase, code),
			Name: fmt.Sprintf(namebase, m.Permalink),
			Kind: kind,
		}
	}
	return uri
}

func scrape_dropbox(uri UrlInfo) UrlInfo {
	return UrlInfo{
		Url:  strings.Replace(uri.Url, "dl=0", "raw=1", -1),
		Name: uri.Name,
		Kind: Document,
	}
}

func uploadboy(uri UrlInfo) UrlInfo {
	up, err := url.Parse(uri.Url)
	if err != nil {
		fmt.Println("Error parsing URL")
		return uri
	}

	fid := strings.TrimSuffix(strings.TrimRight(strings.TrimLeft(up.RequestURI(), "/"), "/"), ".html")

	_, body, errs := gorequest.New().
		DisableKeepAlives(true).
		CloseRequest(true).
		Get(uri.Url).
		End()
	if errs != nil {
		fmt.Println("Error in GET request")
		return uri
	}

	rfname := regexp.MustCompile(`<input (?:type="hidden"|name="fname"|value="([^"]+)") (?:value="([^"]+)"|name="fname"|type="hidden") (?:name="fname"|type="hidden"|value="([^"]+)")>`)
	if !rfname.MatchString(body) {
		fmt.Println("Error, first match don't do it")
		return uri
	}

	all := rfname.FindStringSubmatch(body)
	if len(all) < 2 {
		fmt.Println("Error, all < 2,", all)
		return uri
	}

	fname := ""
	for _, fni := range all[1:] {
		if fname == "" {
			fname = fni
		}
	}
	if fname == "" {
		fmt.Println("I didn't detect any fname :(")
		return uri
	}

	uri.Name = fname

	data1 := struct {
		Op         string `json:"op"`
		UsrLogin   string `json:"usr_login"`
		ID         string `json:"id"`
		Fname      string `json:"fname"`
		Referer    string `json:"referer"`
		MethodFree string `json:"method_free"`
	}{
		Op:         "download1",
		UsrLogin:   "",
		ID:         fid,
		Fname:      fname,
		Referer:    "",
		MethodFree: "Free+Download",
	}

	r := gorequest.New().
		DisableKeepAlives(true).
		CloseRequest(true).
		Post(uri.Url).
		Send(data1)
	r.TargetType = "form"
	_, body, errs = r.End()
	if errs != nil {
		fmt.Println("Error in POST download1")
		return uri
	}

	if !strings.Contains(body, "Create") {
		fmt.Println("Error, body don't have Create")
		return uri
	}

	data := struct {
		Op            string `json:"op"`
		ID            string `json:"id"`
		Rand          string `json:"rand"`
		Referer       string `json:"referer"`
		UsrRes        string `json:"usr_resolution"`
		UsrOS         string `json:"usr_os"`
		UsrBrowser    string `json:"usr_browser"`
		MethodFree    string `json:"method_free"`
		MethodPremium string `json:"method_premium"`
		DownScript    string `json:"down_script"`
	}{
		Op:            "download2",
		ID:            fid,
		Rand:          "",
		Referer:       uri.Url,
		UsrRes:        "1366x768",
		UsrOS:         "Linux+x86_64",
		UsrBrowser:    "Firefox+41",
		MethodFree:    "Free+Download",
		MethodPremium: "",
		DownScript:    "1",
	}

	r = gorequest.New().
		DisableKeepAlives(true).
		CloseRequest(true).
		Post(uri.Url).
		Send(data)
	r.TargetType = "form"
	_, body, errs = r.End()
	if errs != nil {
		fmt.Println("Error in POST download2")
		return uri
	}
	rurl := regexp.MustCompile(`<span\s*class="hidden-xs" style="[^"]+">\s*<a href="([^"]+)">\s*[^<]+\s*<\/a>\s*<\/span>`)
	if !rurl.MatchString(body) {
		fmt.Println("Error in match2")
		return uri
	}
	allu := rurl.FindStringSubmatch(body)
	if len(allu) > 1 {
		uri.Url = allu[1]
	}
	return uri
}

func insta(uri UrlInfo) UrlInfo {
	iurl := uri.Url
	if instaclient == nil {
		uri.Error = "I don't have instagram configured"
		return uri
	}
	re := regexp.MustCompile("https?://instagram.com/p/([a-zA-Z0-9-_]+)")
	matches := re.FindStringSubmatch(iurl)
	if len(matches) <= 1 {
		uri.Error = "Bad instagram URL"
		return uri
	}
	shortcode := matches[1]
	media, err := instaclient.Media.GetShortcode(shortcode)
	if err != nil || media == nil {
		uri.Error = err.Error()
		return uri
	}
	newurl := ""
	if media.Type == "image" {
		if media.Images.StandardResolution != nil {
			uri.Kind = Image
			newurl = media.Images.StandardResolution.URL
		} else if media.Images.LowResolution != nil {
			uri.Kind = Image
			newurl = media.Images.LowResolution.URL
		} else {
			uri.Error = "No image founded"
			return uri
		}
	} else if media.Type == "video" {
		if media.Videos.StandardResolution != nil {
			uri.Kind = Video
			newurl = media.Videos.StandardResolution.URL
		} else if media.Videos.LowResolution != nil {
			uri.Kind = Video
			newurl = media.Videos.LowResolution.URL
		} else {
			uri.Error = "No video founded"
			return uri
		}
	}
	uri.Url = newurl
	return uri
}

func drivehandler(uri UrlInfo) (UrlInfo, bool) {
	chooseone := func(l []string, ask string, def string) string {
		if ask != "" && stringInSlice(ask, l) {
			return ask
		}
		return def
	}
	allg := []GDrive{drivecheck, doccheck, presentationcheck, spreadsheetcheck}
	for _, g := range allg {
		if g.Reg.MatchString(uri.Url) {
			vals := g.Reg.FindStringSubmatch(uri.Url)
			id := vals[1]
			// Check in g.ValidFormats an/or check default
			format := chooseone(g.ValidFormats, uri.Format, g.DefaultFormat)
			newurl := ""
			if format == "" {
				newurl = fmt.Sprintf(g.To, id)
			} else {
				newurl = fmt.Sprintf(g.To, id, format)
			}
			uri.Url = newurl
			return uri, true
		}
	}
	return uri, false
}

func handleslideshare(uri UrlInfo) UrlInfo {
	surl, err := url.ParseRequestURI(uri.Url)
	if err != nil {
		return uri
	}
	nurl := fmt.Sprintf("http://www.slideshare.net%s", surl.EscapedPath())
	s, err := slideshareclient.GetSlideshowUrl(nurl, false)
	if err != nil {
		return uri
	}
	if s.Download && s.DownloadUrl != "" {
		uri.Url = s.DownloadUrl
	}
	return uri
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

	uri, ok := drivehandler(uri)
	if ok {
		return uri
	}

	dom := u.Host
	if checker(dom, "dropbox") {
		return scrape_dropbox(uri)
	}

	if checker(dom, "slideshare.net") {
		return handleslideshare(uri)
	}

	if checker(dom, "soundcloud") {
		go advice()
		return other_soundcloud(uri, kind)
	}

	if checker(dom, "youtube") {
		go advice()
		return yt(uri, kind)
	}

	if checker(dom, "instagram") {
		go advice()
		return insta(uri)
	}

	if checker(dom, "uploadboy.com") {
		go advice()
		return uploadboy(uri)
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
