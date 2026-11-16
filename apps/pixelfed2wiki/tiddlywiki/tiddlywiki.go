package tiddlywiki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/carlo-colombo/pixelfed2wiki/uploader"
	"github.com/minio/minio-go/v7"
)

type TiddlywikiClient struct {
	client   http.Client
	dest     string
	username string
	password string
}

func NewClient(dest string, username string, password string) TiddlywikiClient {
	return TiddlywikiClient{
		dest:     dest,
		client:   http.Client{},
		username: username,
		password: password,
	}
}

func (tc TiddlywikiClient) decorateReq(req *http.Request) *http.Request {
	req.Header.Add("x-requested-with", "TiddlyWiki")
	if tc.username != "" && tc.password != "" {
		req.SetBasicAuth(tc.username, tc.password)
	}
	return req
}

func (tc TiddlywikiClient) CreateIfNew(mc *minio.Client, tiddler Tiddler) error {
	path := fmt.Sprintf("%s/recipes/default/tiddlers/%s",
		tc.dest,
		url.PathEscape(tiddler.Title))

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req = tc.decorateReq(req)

	resp, err := tc.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to check whether tiddler exists: %w", err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("failed to check whether tiddler exists: %s", resp.Status)
		}
		fmt.Printf("(%d)skipping creation %s \n", resp.StatusCode, tiddler.Title)
		return nil
	}

	additionalImages := make([]string, 0, len(tiddler.media))
	additionalThumbnails := make([]string, 0, len(tiddler.media))
	additionalThumbnailWidths := make([]string, 0, len(tiddler.media))
	additionalThumbnailHeights := make([]string, 0, len(tiddler.media))
	for i, photo := range tiddler.media {
		uploader := uploader.NewUploader(os.Getenv("UPLOAD_BUCKET"), mc, photo.data)
		if uploader.Err != nil {
			return fmt.Errorf("failed to prepare photo %d (%s) for post %q: %w", i+1, photo.source, tiddler.Title, uploader.Err)
		}

		imageURL, err := uploader.UploadImage()
		if err != nil {
			return fmt.Errorf("failed to upload photo %d (%s) for post %q: %w", i+1, photo.source, tiddler.Title, err)
		}
		thumbnailURL, thumbnailWidth, thumbnailHeight, err := uploader.UploadThumbnail()
		if err != nil {
			return fmt.Errorf("failed to upload thumbnail for photo %d (%s) in post %q: %w", i+1, photo.source, tiddler.Title, err)
		}

		if i == 0 {
			tiddler.Image = imageURL
			tiddler.Thumbnail = thumbnailURL
			tiddler.ThumbnailWidth = strconv.Itoa(thumbnailWidth)
			tiddler.ThumbnailHeight = strconv.Itoa(thumbnailHeight)
		} else {
			additionalImages = append(additionalImages, imageURL)
			additionalThumbnails = append(additionalThumbnails, thumbnailURL)
			additionalThumbnailWidths = append(additionalThumbnailWidths, strconv.Itoa(thumbnailWidth))
			additionalThumbnailHeights = append(additionalThumbnailHeights, strconv.Itoa(thumbnailHeight))
		}
	}
	tiddler.AdditionalImages = strings.Join(additionalImages, " ")
	tiddler.AdditionalThumbnails = strings.Join(additionalThumbnails, " ")
	tiddler.AdditionalThumbnailWidths = strings.Join(additionalThumbnailWidths, " ")
	tiddler.AdditionalThumbnailHeights = strings.Join(additionalThumbnailHeights, " ")

	tjson, err := json.Marshal(tiddler)

	if err != nil {
		return fmt.Errorf("failed to marshal tiddler: %w", err)
	}

	req, err = http.NewRequest("PUT", path, bytes.NewBuffer(tjson))

	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req = tc.decorateReq(req)

	resp, err = tc.client.Do(req)

	if err != nil {
		return fmt.Errorf("failed to create tiddler: %w", err)
	}

	if resp.StatusCode != 204 {
		return fmt.Errorf("failed to create tiddler: %v", resp.Status)
	}

	fmt.Printf("created %s\n", tiddler.Title)
	return nil
}
