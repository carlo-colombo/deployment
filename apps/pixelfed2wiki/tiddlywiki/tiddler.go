package tiddlywiki

import (
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/mmcdole/gofeed"
)

type Tiddler struct {
	Title       string `json:"title"`
	Tags        string `json:"tags"`
	Text        string `json:"text"`
	Pixelfed    string `json:"pixelfed"`
	Published   string `json:"published"`
	Image       string `json:"image"`
	Description string `json:"description"`
	item        gofeed.Item
	Err         error
	doc         *goquery.Document
}

func (t Tiddler) setTitle(title string) Tiddler {
	if t.Err != nil {
		return t
	}
	t.Title = title[0:int(math.Min(float64(len(title)), 150))]
	return t
}

func (t Tiddler) setTags(item gofeed.Item) Tiddler {
	if t.Err != nil {
		return t
	}

	r := regexp.MustCompile(`#(\S+)`)

	tags := []string{}

	for _, tag := range r.FindAllStringSubmatch(item.Title, -1) {
		tags = append(tags, "pf:"+tag[1])
	}

	t.Tags = strings.Join(tags, " ")

	return t
}

func NewTiddler(item gofeed.Item) Tiddler {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(item.Description))

	if err != nil {
		return Tiddler{
			Err: fmt.Errorf("failed to parse description: %w", err),
		}
	}

	return Tiddler{doc: doc, item: item}.
		setTags(item).
		setTitle(item.Title).
		addImageToText(item).
		setText(item).
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
	t.Image = img.AttrOr("src", "")

	return t
}

func (t Tiddler) setText(item gofeed.Item) Tiddler {
	if t.Err != nil {
		return t
	}
	desc := t.doc.Find("p").Text()

	t.Text += desc
	t.Text += fmt.Sprintf("<br/><br/>%s", item.Link)

	t.Pixelfed = item.Link
	t.Description = desc

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
