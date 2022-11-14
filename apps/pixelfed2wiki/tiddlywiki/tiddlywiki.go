package tiddlywiki

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
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
	req.SetBasicAuth(tc.username, tc.password)
	return req
}

func (tc TiddlywikiClient) CreateIfNew(tiddler Tiddler) error {
	path := fmt.Sprintf("%s/recipes/default/tiddlers/%s",
		tc.dest,
		url.PathEscape(tiddler.Title))

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req = tc.decorateReq(req)

	resp, err := tc.client.Do(req)

	if err != nil || resp.StatusCode != 404 {
		fmt.Printf("(%d)skipping creation %s \n", resp.StatusCode, tiddler.Title)
		return nil
	}

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
