package tiddlywiki

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

type Tiddler struct {
	Title    string `json:"title"`
	Tags     string `json:"tags"`
	Text     string `json:"text"`
	Pixelfed string `json:"pixelfed"`
	err      error
	doc      *goquery.Document
}

func (t Tiddler) setTitle(title string) Tiddler {
	if t.err != nil {
		return t
	}
	t.Title = title
	return t
}

func (t Tiddler) setTags(item gofeed.Item) Tiddler {
	if t.err != nil {
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
			err: fmt.Errorf("failed to parse description: %w", err),
		}
	}

	return Tiddler{doc: doc, Pixelfed: "true"}.
		setTags(item).
		setTitle(item.Title).
		addImageToText(item).
		addTitleToText(item)
}

func (t Tiddler) addImageToText(item gofeed.Item) Tiddler {
	if t.err != nil {
		return t
	}
	img := t.doc.Find("img")
	imgContent, err := goquery.OuterHtml(img)

	if err != nil {
		return Tiddler{err: err}
	}

	t.Text = t.Text + imgContent

	return t
}

func (t Tiddler) addTitleToText(item gofeed.Item) Tiddler {
	if t.err != nil {
		return t
	}
	desc := t.doc.Find("p").Text()

	t.Text = t.Text + desc

	return t
}
