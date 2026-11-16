package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	maxWikiExportBytes   = 512 << 20
	defaultMaxImageBytes = 16 << 20
)

type tiddler map[string]json.RawMessage

type dimensions struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type additionalDimension struct {
	Host       string     `json:"host"`
	Dimensions dimensions `json:"dimensions"`
}

type tiddlerReport struct {
	Title       string                `json:"title"`
	Status      string                `json:"status"`
	PrimaryHost string                `json:"primaryHost,omitempty"`
	Primary     dimensions            `json:"primaryDimensions,omitempty"`
	Additional  []additionalDimension `json:"additionalDimensions,omitempty"`
	Changes     map[string]string     `json:"changes,omitempty"`
	Errors      []string              `json:"errors,omitempty"`
}

type report struct {
	GeneratedAt      time.Time       `json:"generatedAt"`
	WikiHost         string          `json:"wikiHost"`
	BackupFile       string          `json:"backupFile"`
	ExportedTiddlers int             `json:"exportedTiddlers"`
	EligibleTiddlers int             `json:"eligibleTiddlers"`
	AlreadyComplete  int             `json:"alreadyComplete"`
	WouldUpdate      int             `json:"wouldUpdate"`
	Failed           int             `json:"failed"`
	Applied          int             `json:"applied"`
	HostCounts       map[string]int  `json:"hostCounts,omitempty"`
	Tiddlers         []tiddlerReport `json:"tiddlers,omitempty"`
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }
func (s *stringList) Set(value string) error {
	for _, item := range strings.Split(value, ",") {
		item = strings.ToLower(strings.TrimSpace(item))
		if item != "" {
			*s = append(*s, item)
		}
	}
	return nil
}

type backfiller struct {
	wikiURL     *url.URL
	client      *http.Client
	username    string
	password    string
	allowedHost map[string]bool
	maxImage    int64
	cache       map[string]dimensions
}

func main() {
	var allowedHosts stringList
	flag.Var(&allowedHosts, "allowed-host", "Allowed thumbnail host (repeat or comma-separate); defaults to THUMBNAIL_BACKFILL_ALLOWED_HOSTS")
	wikiURLValue := flag.String("wiki-url", os.Getenv("WIKI_URL"), "TiddlyWiki base URL (or WIKI_URL)")
	username := os.Getenv("WIKI_USERNAME")
	password := os.Getenv("WIKI_PASSWORD")
	backupFile := flag.String("backup-file", "", "absolute path for a private pre-migration tiddler export (required)")
	reportFile := flag.String("report-file", "", "write the sanitized report to this path instead of stdout")
	listHosts := flag.Bool("list-hosts", false, "inventory thumbnail hosts without downloading images")
	apply := flag.Bool("apply", false, "write missing dimensions; default is read-only dry-run")
	confirm := flag.Bool("confirm-live-write", false, "explicitly approve production writes")
	limit := flag.Int("limit", 50, "maximum tiddlers to update in apply mode")
	maxImageBytes := flag.Int64("max-image-bytes", defaultMaxImageBytes, "maximum bytes read per thumbnail")
	flag.Parse()

	if len(allowedHosts) == 0 {
		_ = allowedHosts.Set(os.Getenv("THUMBNAIL_BACKFILL_ALLOWED_HOSTS"))
	}

	if err := run(*wikiURLValue, username, password, allowedHosts, *backupFile, *reportFile, *listHosts, *apply, *confirm, *limit, *maxImageBytes); err != nil {
		fmt.Fprintln(os.Stderr, "thumbnail backfill:", err)
		os.Exit(1)
	}
}

func run(wikiURL, username, password string, allowedHosts []string, backupFile, reportFile string, listHosts, apply, confirm bool, limit int, maxImageBytes int64) error {
	if wikiURL == "" || username == "" || password == "" {
		return errors.New("set WIKI_URL, WIKI_USERNAME, and WIKI_PASSWORD")
	}
	backupPath, err := validateBackupPath(backupFile)
	if err != nil {
		return err
	}
	if limit < 1 || maxImageBytes < 1 {
		return errors.New("--limit and --max-image-bytes must be positive")
	}
	if apply && !confirm {
		return errors.New("apply mode requires --confirm-live-write")
	}

	baseURL, err := url.Parse(wikiURL)
	if err != nil || (baseURL.Scheme != "http" && baseURL.Scheme != "https") || baseURL.Host == "" || baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return fmt.Errorf("invalid WIKI_URL %q", wikiURL)
	}
	baseURL.Path = strings.TrimRight(baseURL.Path, "/")
	allowed := make(map[string]bool, len(allowedHosts))
	for _, host := range allowedHosts {
		allowed[strings.ToLower(strings.TrimSpace(host))] = true
	}
	wikiHost := strings.ToLower(baseURL.Hostname())
	client := &http.Client{
		Timeout: 45 * time.Second,
		CheckRedirect: func(request *http.Request, _ []*http.Request) error {
			host := strings.ToLower(request.URL.Hostname())
			if host != wikiHost && !allowed[host] {
				return errors.New("refusing redirect to a host outside the wiki and image allowlists")
			}
			return nil
		},
	}

	bf := &backfiller{
		wikiURL:     baseURL,
		client:      client,
		username:    username,
		password:    password,
		allowedHost: allowed,
		maxImage:    maxImageBytes,
		cache:       make(map[string]dimensions),
	}

	tiddlers, rawExport, err := bf.getAllTiddlers()
	if err != nil {
		return err
	}
	if err := writePrivateFile(backupPath, rawExport); err != nil {
		return fmt.Errorf("write pre-migration tiddler export: %w", err)
	}

	result := report{
		GeneratedAt:      time.Now().UTC(),
		WikiHost:         baseURL.Host,
		BackupFile:       backupPath,
		ExportedTiddlers: len(tiddlers),
		HostCounts:       make(map[string]int),
	}
	eligible := make([]tiddler, 0)
	for _, item := range tiddlers {
		if getString(item, "pixelfed") == "" || getString(item, "published") == "" {
			continue
		}
		eligible = append(eligible, item)
		for _, imageURL := range imageURLs(item) {
			if parsed, parseErr := url.Parse(imageURL); parseErr == nil {
				result.HostCounts[strings.ToLower(parsed.Hostname())]++
			}
		}
	}
	result.EligibleTiddlers = len(eligible)
	if listHosts {
		return writeReport(reportFile, result)
	}
	if len(allowed) == 0 {
		return errors.New("provide --allowed-host after reviewing --list-hosts output")
	}

	rows := make([]tiddlerReport, 0, len(eligible))
	changesByTitle := make(map[string]map[string]string)
	for _, item := range eligible {
		row, changes := bf.inspectTiddler(item)
		rows = append(rows, row)
		if len(row.Errors) > 0 {
			result.Failed++
		} else if len(changes) > 0 {
			result.WouldUpdate++
			changesByTitle[row.Title] = changes
		} else {
			result.AlreadyComplete++
		}
	}
	result.Tiddlers = rows

	if apply {
		if result.Failed > 0 {
			return writeReportAndError(reportFile, result, fmt.Errorf("refusing to apply: %d tiddlers have validation errors", result.Failed))
		}
		applied, err := bf.applyChanges(eligible, changesByTitle, limit)
		result.Applied = applied
		if err != nil {
			return writeReportAndError(reportFile, result, err)
		}
	}
	return writeReport(reportFile, result)
}

func validateBackupPath(path string) (string, error) {
	if path == "" || !filepath.IsAbs(path) {
		return "", errors.New("--backup-file must be an absolute path outside the repository")
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	workingDirectory, err := os.Getwd()
	if err != nil {
		return "", err
	}
	repositoryRoot := findRepositoryRoot(workingDirectory)
	relative, err := filepath.Rel(repositoryRoot, absolute)
	if err != nil {
		return "", err
	}
	if relative != ".." && !strings.HasPrefix(relative, ".."+string(os.PathSeparator)) {
		return "", errors.New("backup file must be outside the current repository")
	}
	if _, err := os.Stat(filepath.Dir(absolute)); err != nil {
		return "", fmt.Errorf("backup directory is not available: %w", err)
	}
	return absolute, nil
}

func findRepositoryRoot(start string) string {
	current := start
	for {
		if _, err := os.Stat(filepath.Join(current, ".git")); err == nil {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return start
		}
		current = parent
	}
}

func (bf *backfiller) getAllTiddlers() ([]tiddler, []byte, error) {
	endpoint := *bf.wikiURL
	endpoint.Path = strings.TrimRight(endpoint.Path, "/") + "/recipes/default/tiddlers.json"
	query := endpoint.Query()
	query.Set("exclude", "__thumbnail_backfill_no_such_field__")
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequest(http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return nil, nil, fmt.Errorf("create tiddler-export request: %w", err)
	}
	request.SetBasicAuth(bf.username, bf.password)
	response, err := bf.client.Do(request)
	if err != nil {
		return nil, nil, fmt.Errorf("request tiddler export: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("tiddler export returned %s", response.Status)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, maxWikiExportBytes+1))
	if err != nil {
		return nil, nil, fmt.Errorf("read tiddler export: %w", err)
	}
	if int64(len(data)) > maxWikiExportBytes {
		return nil, nil, fmt.Errorf("tiddler export exceeds %d bytes", maxWikiExportBytes)
	}
	var tiddlers []tiddler
	if err := json.Unmarshal(data, &tiddlers); err != nil {
		return nil, nil, fmt.Errorf("decode tiddler export: %w", err)
	}
	if len(tiddlers) > 0 {
		includesText := false
		for _, item := range tiddlers {
			if _, ok := item["text"]; ok {
				includesText = true
				break
			}
		}
		if !includesText {
			return nil, nil, errors.New("TiddlyWiki export omitted text fields; refusing to call it a restorable backup")
		}
	}
	return tiddlers, data, nil
}

func (bf *backfiller) inspectTiddler(item tiddler) (tiddlerReport, map[string]string) {
	title := getString(item, "title")
	row := tiddlerReport{Title: title, Status: "unchanged"}
	changes := make(map[string]string)
	urls := imageURLs(item)
	if len(urls) == 0 {
		row.Errors = []string{"no thumbnail or image URL"}
		row.Status = "error"
		return row, nil
	}

	primary, host, err := bf.measure(urls[0])
	row.PrimaryHost = host
	if err != nil {
		row.Errors = append(row.Errors, "primary: "+err.Error())
	} else {
		row.Primary = primary
		compareDimensionField(item, "thumbnail-width", primary.Width, changes, &row.Errors)
		compareDimensionField(item, "thumbnail-height", primary.Height, changes, &row.Errors)
	}

	additionalSizes := make([]dimensions, 0, len(urls)-1)
	for index, imageURL := range urls[1:] {
		size, host, measureErr := bf.measure(imageURL)
		if measureErr != nil {
			row.Errors = append(row.Errors, fmt.Sprintf("additional thumbnail %d: %s", index+1, measureErr))
			continue
		}
		additionalSizes = append(additionalSizes, size)
		row.Additional = append(row.Additional, additionalDimension{Host: host, Dimensions: size})
	}
	if len(additionalSizes) == len(urls)-1 && len(additionalSizes) > 0 {
		compareDimensionList(item, "additional-thumbnail-widths", additionalSizes, true, changes, &row.Errors)
		compareDimensionList(item, "additional-thumbnail-heights", additionalSizes, false, changes, &row.Errors)
	}

	if len(row.Errors) > 0 {
		row.Status = "error"
		return row, nil
	}
	if len(changes) > 0 {
		row.Status = "would-update"
		row.Changes = changes
	}
	return row, changes
}

func compareDimensionField(item tiddler, field string, actual int, changes map[string]string, errors *[]string) {
	value, exists, err := fieldString(item, field)
	if err != nil {
		*errors = append(*errors, fmt.Sprintf("invalid existing %s: %v", field, err))
		return
	}
	if !exists || value == "" {
		changes[field] = strconv.Itoa(actual)
		return
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		*errors = append(*errors, fmt.Sprintf("invalid existing %s=%q", field, value))
	} else if parsed != actual {
		*errors = append(*errors, fmt.Sprintf("existing %s=%d does not match image size %d", field, parsed, actual))
	}
}

func compareDimensionList(item tiddler, field string, sizes []dimensions, width bool, changes map[string]string, errors *[]string) {
	values := make([]string, len(sizes))
	for index, size := range sizes {
		value := size.Height
		if width {
			value = size.Width
		}
		values[index] = strconv.Itoa(value)
	}
	actual := strings.Join(values, " ")
	existing, exists, err := fieldString(item, field)
	if err != nil {
		*errors = append(*errors, fmt.Sprintf("invalid existing %s: %v", field, err))
		return
	}
	if !exists || existing == "" {
		changes[field] = actual
		return
	}
	if existing != actual {
		*errors = append(*errors, fmt.Sprintf("existing %s=%q does not match measured value %q", field, existing, actual))
	}
}

func (bf *backfiller) measure(imageURL string) (dimensions, string, error) {
	parsed, err := url.Parse(imageURL)
	if err != nil {
		return dimensions{}, "", fmt.Errorf("invalid image URL")
	}
	host := strings.ToLower(parsed.Hostname())
	if parsed.Scheme != "https" || !bf.allowedHost[host] {
		return dimensions{}, host, fmt.Errorf("host %q is not in the HTTPS allowlist", host)
	}
	if cached, ok := bf.cache[imageURL]; ok {
		return cached, host, nil
	}
	request, err := http.NewRequest(http.MethodGet, imageURL, nil)
	if err != nil {
		return dimensions{}, host, fmt.Errorf("create image request: %w", err)
	}
	response, err := bf.client.Do(request)
	if err != nil {
		return dimensions{}, host, fmt.Errorf("download image: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return dimensions{}, host, fmt.Errorf("download returned %s", response.Status)
	}
	if response.ContentLength > bf.maxImage {
		return dimensions{}, host, fmt.Errorf("image exceeds %d bytes", bf.maxImage)
	}
	config, _, err := image.DecodeConfig(io.LimitReader(response.Body, bf.maxImage))
	if err != nil {
		return dimensions{}, host, fmt.Errorf("decode image dimensions: %w", err)
	}
	result := dimensions{Width: config.Width, Height: config.Height}
	if result.Width < 1 || result.Height < 1 {
		return dimensions{}, host, errors.New("decoded non-positive image dimensions")
	}
	bf.cache[imageURL] = result
	return result, host, nil
}

func (bf *backfiller) applyChanges(items []tiddler, pending map[string]map[string]string, limit int) (int, error) {
	latest, _, err := bf.getAllTiddlers()
	if err != nil {
		return 0, fmt.Errorf("refresh tiddlers before apply: %w", err)
	}
	latestByTitle := make(map[string]tiddler, len(latest))
	for _, item := range latest {
		latestByTitle[getString(item, "title")] = item
	}

	type update struct {
		title string
		item  tiddler
	}
	updates := make([]update, 0, limit)
	for _, original := range items {
		title := getString(original, "title")
		changes := pending[title]
		if len(changes) == 0 {
			continue
		}
		current, ok := latestByTitle[title]
		if !ok {
			return 0, fmt.Errorf("tiddler %q disappeared after inventory; rerun the dry run", title)
		}
		for _, field := range []string{"pixelfed", "published", "thumbnail", "image", "additional-thumbnails"} {
			if getString(current, field) != getString(original, field) {
				return 0, fmt.Errorf("tiddler %q field %s changed after inventory; rerun the dry run", title, field)
			}
		}
		for field, value := range changes {
			existing, exists, err := fieldString(current, field)
			if err != nil {
				return 0, fmt.Errorf("tiddler %q field %s is not a string; review manually", title, field)
			}
			if exists && existing != "" {
				if existing != value {
					return 0, fmt.Errorf("tiddler %q field %s changed after inventory; rerun the dry run", title, field)
				}
				continue
			}
			current[field], _ = json.Marshal(value)
		}
		updates = append(updates, update{title: title, item: current})
		if len(updates) >= limit {
			break
		}
	}

	applied := 0
	for _, update := range updates {
		body, err := json.Marshal(update.item)
		if err != nil {
			return applied, fmt.Errorf("marshal tiddler %q: %w", update.title, err)
		}
		endpoint := strings.TrimRight(bf.wikiURL.String(), "/") + "/recipes/default/tiddlers/" + url.PathEscape(update.title)
		request, err := http.NewRequest(http.MethodPut, endpoint, bytes.NewReader(body))
		if err != nil {
			return applied, fmt.Errorf("create update request for %q: %w", update.title, err)
		}
		request.Header.Set("x-requested-with", "TiddlyWiki")
		request.Header.Set("Content-Type", "application/json")
		request.SetBasicAuth(bf.username, bf.password)
		response, err := bf.client.Do(request)
		if err != nil {
			return applied, fmt.Errorf("update tiddler %q: %w", update.title, err)
		}
		response.Body.Close()
		if response.StatusCode != http.StatusNoContent {
			return applied, fmt.Errorf("update tiddler %q returned %s", update.title, response.Status)
		}
		applied++
	}
	return applied, nil
}

func imageURLs(item tiddler) []string {
	primary := getString(item, "thumbnail")
	if primary == "" {
		primary = getString(item, "image")
	}
	if primary == "" {
		return nil
	}
	urls := []string{primary}
	urls = append(urls, strings.Fields(getString(item, "additional-thumbnails"))...)
	return urls
}

func getString(item tiddler, field string) string {
	value, _, _ := fieldString(item, field)
	return value
}

func fieldString(item tiddler, field string) (string, bool, error) {
	raw, exists := item[field]
	if !exists {
		return "", false, nil
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", true, err
	}
	return value, true, nil
}

func writePrivateFile(path string, data []byte) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	file, err := os.OpenFile(absolute, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		return fmt.Errorf("create backup %q: %w", absolute, err)
	}
	if _, err := file.Write(data); err != nil {
		file.Close()
		return fmt.Errorf("write backup: %w", err)
	}
	if err := file.Sync(); err != nil {
		file.Close()
		return fmt.Errorf("sync backup: %w", err)
	}
	if err := file.Close(); err != nil {
		return fmt.Errorf("close backup: %w", err)
	}
	return nil
}

func writeReport(path string, value report) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if path == "" || path == "-" {
		_, err = os.Stdout.Write(data)
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func writeReportAndError(path string, value report, runErr error) error {
	if err := writeReport(path, value); err != nil {
		return fmt.Errorf("%v (also failed to write report: %w)", runErr, err)
	}
	return runErr
}
