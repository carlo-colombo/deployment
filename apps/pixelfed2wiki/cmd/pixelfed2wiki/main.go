package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
	s3Endpoint := os.Getenv("S3_ENDPOINT")
	if s3Endpoint == "" {
		s3Endpoint = "s3.nl-ams.scw.cloud"
	}
	secure := true
	if value := os.Getenv("S3_SECURE"); value != "" {
		secure, err = strconv.ParseBool(value)
		if err != nil {
			log.Fatalf("invalid S3_SECURE value %q: %v", value, err)
		}
	}
	options := &minio.Options{
		Creds:  credentials.NewStaticV4(os.Getenv("ACCESS_KEYID"), os.Getenv("SECRET_KEY"), ""),
		Secure: secure,
	}
	if os.Getenv("S3_ENDPOINT") != "" {
		options.BucketLookup = minio.BucketLookupPath
	}
	mc, err := minio.New(s3Endpoint, options)
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
