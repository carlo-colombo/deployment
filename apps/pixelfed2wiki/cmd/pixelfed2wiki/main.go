package main

import (
	"fmt"
	"log"
	"os"

	"github.com/carlo-colombo/pixelfed2wiki/tiddlywiki"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/mmcdole/gofeed"
)

func main() {
	if len(os.Args) < 3 {
		log.Fatal("usage: pixelfed2wiki FEED_URL WIKI_URL")
	}
	feed, err := gofeed.NewParser().ParseURL(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	mc, err := minio.New("s3.nl-ams.scw.cloud", &minio.Options{Creds: credentials.NewStaticV4(os.Getenv("ACCESS_KEYID"), os.Getenv("SECRET_KEY"), ""), Secure: true})
	if err != nil {
		log.Fatal(err)
	}
	client := tiddlywiki.NewClient(os.Args[2], os.Getenv("WIKI_USERNAME"), os.Getenv("WIKI_PASSWORD"))
	for _, item := range feed.Items {
		for _, tiddler := range tiddlywiki.NewTiddlers(*item) {
			if tiddler.Err != nil {
				fmt.Printf("error while creating tiddler: %s\n", tiddler.Err)
				continue
			}
			if err := client.CreateIfNew(mc, tiddler); err != nil {
				fmt.Printf("error while saving tiddler: %s\n", err)
			}
		}
	}
}
