package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"github.com/mmcdole/gofeed"
)

type Tiddler struct {
	Title    string            `json:"title"`
	Tags     string            `json:"tags"`
	Text     string            `json:"text"`
	Pixelfed string            `json:"pixelfed"`
	err      error             `json:"err"`
	doc      *goquery.Document `json:"doc"`
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

func dump(data interface{}) {
	b, _ := json.MarshalIndent(data, "", "  ")
	fmt.Print(string(b))
}

func main() {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	client := &http.Client{}
	for _, value := range feed.Items {
		t := NewTiddler(*value)

		if t.err != nil {
			log.Fatal(t.err)
		}

		tjson, err := json.Marshal(t)

		path := fmt.Sprintf("%s/recipes/default/tiddlers/%s",
			os.Args[2],
			url.PathEscape(t.Title))

		resp, err := http.Get(path)

		if err != nil || resp.StatusCode != 404 {
			fmt.Printf("(%d)skipping %s\n", resp.StatusCode, t.Title)
			continue
		}

		req, err := http.NewRequest("PUT", path, bytes.NewBuffer(tjson))

		if err != nil {
			log.Fatal(err)
		}

		req.Header.Add("x-requested-with", "TiddlyWiki")
		req.SetBasicAuth(
			os.Getenv("WIKI_USERNAME"),
			os.Getenv("WIKI_PASSWORD"))

		if err != nil {
			log.Fatal(err)
		}

		resp, err = client.Do(req)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%v\n", resp)
	}
}
