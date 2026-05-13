package tiddlywiki

import (
	"fmt"
	"io"
	"math"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/carlo-colombo/pixelfed2wiki/uploader"
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
	Thumbnail   string `json:"thumbnail"`
	item        gofeed.Item
	Err         error
	doc         *goquery.Document
	imageBytes  []byte
	mimeType    string
}

func NewTiddlers(item gofeed.Item) []Tiddler {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(item.Content))

	if err != nil {
		return []Tiddler{{
			Err: fmt.Errorf("failed to parse description: %w", err),
		}}
	}

	var tiddlers []Tiddler
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("src")
		suffix := ""
		if i > 0 {
			suffix = fmt.Sprintf(" /%d", i+1)
		}

		t := Tiddler{doc: doc, item: item}.
			fetchImage(src).
			setTags(item).
			setTitle(item.Title, suffix).
			setText(item).
			addPublished()
		tiddlers = append(tiddlers, t)
	})

	return tiddlers
}

func (t Tiddler) fetchImage(imageUrl string) Tiddler {
	if t.Err != nil {
		return t
	}

	t.Image = imageUrl

	resp, err := http.Get(t.Image)
	if err != nil {
		return Tiddler{Err: fmt.Errorf("cannot retrieve image '%s': %w", t.Image, err)}
	}
	defer resp.Body.Close()

	imageBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return Tiddler{Err: fmt.Errorf("cannot read image: %w", err)}
	}

	t.imageBytes = imageBytes
	t.mimeType = http.DetectContentType(imageBytes)

	return t
}

func (t Tiddler) UploadAndSetImage(u uploader.Uploader) Tiddler {
	if t.Err != nil {
		return t
	}

	imageUrl, err := u.UploadImage()
	if err != nil {
		return Tiddler{Err: fmt.Errorf("cannot upload image: %w", err)}
	}

	t.Image = imageUrl
	return t
}

func (t Tiddler) UploadAndSetThumbnail(u uploader.Uploader) Tiddler {
	if t.Err != nil {
		return t
	}

	imageUrl, err := u.UploadThumbnail()
	if err != nil {
		return Tiddler{Err: fmt.Errorf("cannot upload thumbnail: %w", err)}
	}

	t.Thumbnail = imageUrl
	return t
}

func (t Tiddler) setTitle(title string, suffix string) Tiddler {
	if t.Err != nil {
		return t
	}
	limit := 150 - len(suffix)
	t.Title = title[0:int(math.Min(float64(len(title)), float64(limit)))] + suffix
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

func (t Tiddler) setText(item gofeed.Item) Tiddler {
	if t.Err != nil {
		return t
	}
	t.Text = "<img src={{!!image}} /><br>"

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
