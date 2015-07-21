package rssbot

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	rss "github.com/jteeuwen/go-pkg-rss" // Subscribe to RSS
)

// NewStruct represent a "new" document
type NewStruct struct {
	Text          string
	Images        []string
	Title         string
	Description   string
	Date          string
	Author        string
	LinksText     string
	Links         []string
	InternalLinks []string
}

// BuildText from the data
func (ns *NewStruct) BuildText() string {
	if ns.Text != "" {
		return ns.Text
	}

	headerlinks := `[`
	for _, l := range ns.Links {
		headerlinks = fmt.Sprintf("%s %s", headerlinks, l)
	}
	headerlinks = headerlinks + ` ]`

	descr := ns.Description
	if descr != "" {
		descr = "\n" + descr + "\n---------"
	}
	by := ns.Author
	if by == "" {
		by = "No author defined."
	}

	ns.Text = fmt.Sprintf(`%s
---------
%s
---------%s
By: %s
Date: %s
---------`, headerlinks, ns.Title, descr, by, ns.Date)

	return ns.Text
}

// ExtractNews will return the proper structures from items
func ExtractNews(newitems []*rss.Item) []NewStruct {
	var newst []NewStruct
	// avoid create a channel (the reverseRss) if no items
	if len(newitems) <= 0 {
		return newst
	}
	for new := range reverseRss(newitems) {
		// init
		// linkstr := ""
		var (
			linkslist     []string
			linksinternal []string
		)
		// linkslist := make([]string, 0)
		var images []string
		descrip := ""

		// get all links
		if new.Links != nil {
			links := new.Links
			for _, l := range links {
				l2 := *l
				linkslist = append(linkslist, l2.Href)
				// linkstr += fmt.Sprintf(" - (%s)", l2.Href)
			}
		}

		// Read HTML
		content := new.Description
		if new.Content != nil {
			content = new.Content.Text
		}

		f := func(reader io.Reader, list []string, search string, attr string) (string, []string) {
			text := ""
			doc, err := goquery.NewDocumentFromReader(reader)
			if err == nil {
				doc.Find(search).Each(func(i int, s *goquery.Selection) {
					val, ok := s.Attr(attr)
					if ok {
						list = append(list, val)
					}
				})
				text = doc.Text()
			}
			return text, list
		}

		cleanlinks := func(links []string) []string {
			suffixlist := []string{
				"/embed/simple", // vine
			}
			singleregex := map[*regexp.Regexp]string{
				regexp.MustCompile(`https?://(?:www\.)?youtube\.com/embed/([\w-_]+)\??`): "https://www.youtube.com/watch?v=%s", // youtube embed
			}
			for i, l := range links {
				for _, suff := range suffixlist {
					if strings.HasSuffix(l, suff) {
						links[i] = strings.TrimRight(l, suff)
					}
				}
				for k, v := range singleregex {
					if k.MatchString(l) {
						matchs := k.FindStringSubmatch(l)
						if len(matchs) <= 1 {
							continue
						}
						links[i] = fmt.Sprintf(v, matchs[1])
					}
				}
			}
			return links
		}

		descrip, images = f(strings.NewReader(content), images, "img", "src")
		descrip, images = f(strings.NewReader(descrip), images, "img", "src")
		_, linksinternal = f(strings.NewReader(content), linksinternal, "iframe", "src")
		_, linksinternal = f(strings.NewReader(content), linksinternal, "video", "src")
		_, linksinternal = f(strings.NewReader(content), linksinternal, "source", "src")

		linksinternal = cleanlinks(linksinternal)

		new.Title, descrip = analyzeTitleDescrip(new.Title, descrip)

		// itemstr := fmt.Sprintf("%s%s\n%s", new.Title, linkstr, descrip)
		newst = append(newst, NewStruct{
			Text:          "",
			Images:        images,
			Title:         new.Title,
			Description:   descrip,
			Date:          new.PubDate,
			Author:        new.Author.Name,
			LinksText:     "",
			Links:         linkslist,
			InternalLinks: linksinternal,
		})

		// newst = append(newst, NewStruct{itemstr, images})
	}
	return newst
}

func analyzeTitleDescrip(title string, descrip string) (string, string) {
	title = strings.TrimSpace(title)
	descrip = strings.TrimSpace(descrip)
	if strings.HasSuffix(title, "...") && strings.HasPrefix(descrip, title[:len(title)-3]) {
		title = descrip
		descrip = ""
	} else if title == descrip {
		descrip = ""
	}

	return title, descrip
}
