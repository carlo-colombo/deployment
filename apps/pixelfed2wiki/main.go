package main

import (
	"fmt"
	"log"
	"os"

	"github.com/carlo-colombo/pixelfed2wiki/tiddlywiki"

	"github.com/mmcdole/gofeed"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
	fp := gofeed.NewParser()

	feed, err := fp.ParseURL(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	minioClient, err := minio.New(
		"litapp-blog-images.s3.nl-ams.scw.cloud", &minio.Options{
			Creds: credentials.NewStaticV4(
				os.Getenv("ACCESS_KEYID"),
				os.Getenv("SECRET_KEY"), ""),
			Secure: true,
		})

	if err != nil {
		log.Fatal(err)
	}

	client := tiddlywiki.NewClient(os.Args[2],
		os.Getenv("WIKI_USERNAME"),
		os.Getenv("WIKI_PASSWORD"))

	for _, value := range feed.Items {
		tiddler := tiddlywiki.NewTiddler(*value)

		if tiddler.Err != nil {
			fmt.Printf("error while creating tiddler %s\n", tiddler.Err)
			continue
		}
		err := client.CreateIfNew(minioClient, tiddler)
		if err != nil {
			fmt.Printf("error while saving tiddler: %s\n", err)
			continue
		}
	}
}
