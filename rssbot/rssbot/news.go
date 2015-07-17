package rssbot

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
	rss "github.com/jteeuwen/go-pkg-rss" // Subscribe to RSS
)

// NewStruct represent a "new" document
type NewStruct struct {
	Text        string
	Images      []string
	Title       string
	Description string
	Date        string
	Author      string
	LinksText   string
	Links       []string
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
		var linkslist []string
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
		// finaltext := fmt.Sprintf("%s<br>%s", new.Description, content)
		read := strings.NewReader(content)
		doc, err := goquery.NewDocumentFromReader(read)

		if err == nil {
			doc.Find("img").Each(func(i int, s *goquery.Selection) {
				val, ok := s.Attr("src")
				if ok {
					images = append(images, val)
				}
			})

			descrip = doc.Text()

			doc2, err2 := goquery.NewDocumentFromReader(strings.NewReader(descrip))
			if err2 == nil {
				doc2.Find("img").Each(func(i int, s *goquery.Selection) {
					val, ok := s.Attr("src")
					if ok {
						images = append(images, val)
					}
				})
				descrip = doc2.Text()
			}
		}

		new.Title, descrip = analyzeTitleDescrip(new.Title, descrip)

		// itemstr := fmt.Sprintf("%s%s\n%s", new.Title, linkstr, descrip)
		newst = append(newst, NewStruct{"", images, new.Title, descrip, new.PubDate, new.Author.Name, "", linkslist})

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
