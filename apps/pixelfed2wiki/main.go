package main

import (
	"fmt"
	"log"
	"os"

	"github.com/carlo-colombo/pixelfed2wiki/tiddlywiki"

	"github.com/mmcdole/gofeed"
)

func main() {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	client := tiddlywiki.NewClient(os.Args[2],
		os.Getenv("WIKI_USERNAME"),
		os.Getenv("WIKI_PASSWORD"))

	for _, value := range feed.Items {
		tiddler := tiddlywiki.NewTiddler(*value)

		err := client.CreateIfNew(tiddler)
		if err != nil {
			fmt.Printf("error while creating tiddler %s\n", tiddler.Title)
			continue
		}
	}
}
