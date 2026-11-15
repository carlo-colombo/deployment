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
	"github.com/mmcdole/gofeed"
)

type Tiddler struct {
	Title                string `json:"title"`
	Tags                 string `json:"tags"`
	Text                 string `json:"text"`
	Pixelfed             string `json:"pixelfed"`
	Published            string `json:"published"`
	Image                string `json:"image"`
	Description          string `json:"description"`
	Thumbnail            string `json:"thumbnail"`
	AdditionalImages     string `json:"additional-images,omitempty"`
	AdditionalThumbnails string `json:"additional-thumbnails,omitempty"`
	item                 gofeed.Item
	Err                  error `json:"-"`
	doc                  *goquery.Document
	media                []photoAsset
}

type photoAsset struct {
	source string
	data   []byte
}

var splitTag = regexp.MustCompile(`(?i)(^|[^[:alnum:]_])#split($|[^[:alnum:]_])`)

func NewTiddlers(item gofeed.Item) []Tiddler {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(item.Content))
	if err != nil {
		return []Tiddler{{Err: fmt.Errorf("failed to parse description: %w", err)}}
	}

	var imageURLs []string
	doc.Find("img").Each(func(_ int, s *goquery.Selection) {
		if src, exists := s.Attr("src"); exists && src != "" {
			imageURLs = append(imageURLs, src)
		}
	})

	if len(imageURLs) == 0 {
		return nil
	}

	splitPost := isSplitPost(item, doc)
	groups := [][]string{imageURLs}
	if splitPost {
		groups = make([][]string, len(imageURLs))
		for i, imageURL := range imageURLs {
			groups[i] = []string{imageURL}
		}
	}

	tiddlers := make([]Tiddler, 0, len(groups))
	for i, group := range groups {
		suffix := ""
		if splitPost && i > 0 {
			suffix = fmt.Sprintf(" /%d", i+1)
		}

		t := Tiddler{doc: doc, item: item}.
			setTags(item).
			setTitle(item.Title, suffix).
			setText(item).
			addPublished()
		if !splitPost && len(group) > 1 {
			t.Text = strings.TrimPrefix(t.Text, "<img src={{!!image}} /><br>")
		}
		if t.Err == nil {
			for photoIndex, imageURL := range group {
				data, err := downloadImage(imageURL)
				if err != nil {
					t.Err = fmt.Errorf("post %q (%s), photo %d (%s): %w", item.Title, item.Link, photoIndex+1, imageURL, err)
					break
				}
				t.media = append(t.media, photoAsset{source: imageURL, data: data})
				if photoIndex == 0 {
					t.Image = imageURL
				} else {
					t.addAdditionalImage(imageURL)
				}
			}
		}
		tiddlers = append(tiddlers, t)
	}

	return tiddlers
}

func isSplitPost(item gofeed.Item, doc *goquery.Document) bool {
	return splitTag.MatchString(item.Title) || splitTag.MatchString(doc.Text())
}

func downloadImage(imageURL string) ([]byte, error) {
	resp, err := http.Get(imageURL)
	if err != nil {
		return nil, fmt.Errorf("cannot retrieve image: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("cannot retrieve image: %s", resp.Status)
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("cannot read image: %w", err)
	}
	return imageBytes, nil
}

func (t *Tiddler) addAdditionalImage(imageURL string) {
	if t.AdditionalImages != "" {
		t.AdditionalImages += " "
	}
	t.AdditionalImages += imageURL
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
		if strings.EqualFold(tag[1], "split") {
			continue
		}
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
