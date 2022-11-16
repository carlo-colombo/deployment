package tiddlywiki

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

type Tiddler struct {
	Title     string `json:"title"`
	Tags      string `json:"tags"`
	Text      string `json:"text"`
	Pixelfed  string `json:"pixelfed"`
	Published string `json:"published"`
	item      gofeed.Item
	Err       error
	doc       *goquery.Document
}

func (t Tiddler) setTitle(title string) Tiddler {
	if t.Err != nil {
		return t
	}
	t.Title = title
	return t
}

func (t Tiddler) setTags(item gofeed.Item) Tiddler {
	if t.Err != nil {
		return t
	}

	r := regexp.MustCompile("#([^ ]+)")

	for _, tag := range r.FindAllStringSubmatch(item.Title, -1) {
		t.Tags += " pf:" + tag[1]
	}

	return t
}

func NewTiddler(item gofeed.Item) Tiddler {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(item.Description))

	if err != nil {
		return Tiddler{
			Err: fmt.Errorf("failed to parse description: %w", err),
		}
	}

	return Tiddler{doc: doc, Pixelfed: "true", item: item}.
		setTags(item).
		setTitle(item.Title).
		addImageToText(item).
		addTitleToText(item).
		addPublished()
}

func (t Tiddler) addImageToText(item gofeed.Item) Tiddler {
	if t.Err != nil {
		return t
	}
	img := t.doc.Find("img")
	imgContent, err := goquery.OuterHtml(img)

	if err != nil {
		return Tiddler{Err: err}
	}

	t.Text += imgContent

	return t
}

func (t Tiddler) addTitleToText(item gofeed.Item) Tiddler {
	if t.Err != nil {
		return t
	}
	desc := t.doc.Find("p").Text()

	t.Text += desc
	t.Text += fmt.Sprintf("<br/><br/>%s", item.Link)

	return t
}

func (t Tiddler) addPublished() Tiddler {
	if t.Err != nil {
		return t
	}
	p, err := time.Parse(time.RFC3339, t.item.Published)
	if err != nil {
		return Tiddler{Err: err}
	}

	t.Published = p.Format("20060102150405") + "000"
	return t
}
