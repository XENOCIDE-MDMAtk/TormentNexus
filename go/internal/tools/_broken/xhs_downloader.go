package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

var (
	xhsURLPattern = regexp.MustCompile(`(https?://(www\.)?xiaohongshu\.com/(explore|discovery/item|user/profile)/([a-zA-Z0-9]+))|(https?://xhslink\.com/([a-zA-Z0-9]+))`)
)

type xhsNoteInfo struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Images      []string `json:"images"`
	Videos      []string `json:"videos"`
	LivePhotos  []struct {
		Image string `json:"image"`
		Video string `json:"video"`
	} `json:"live_photos"`
	PublishTime time.Time `json:"publish_time"`
}

func HandleExtractNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	noteURL, _ :=getString(args, "url")
	if noteURL == "" {
		return err("url parameter is required")
}

	download, _ :=getBool(args, "download")
	indexList := getIndexList(args)

	client := http.DefaultClient

	// First resolve the actual note URL if it's a short link
	resolvedURL, resolveErr := resolveXHSURL(client, noteURL)
	if resolveErr != nil {
		return err(fmt.Sprintf("failed to resolve URL: %v", resolveErr))
}

	// Extract note ID from URL
	noteID, extractErr := extractNoteID(resolvedURL)
	if extractErr != nil {
		return err(fmt.Sprintf("failed to extract note ID: %v", extractErr))
}

	// Fetch note details
	noteInfo, fetchErr := fetchNoteDetails(client, noteID)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch note details: %v", fetchErr))
}

	if download {
		downloadErr := downloadNoteFiles(client, noteInfo, indexList)
		if downloadErr != nil {
			return err(fmt.Sprintf("failed to download note files: %v", downloadErr))

	}

	response, jsonErr := json.Marshal(noteInfo)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal response: %v", jsonErr))
}

	return ok(string(response))
}

}

func HandleDownloadNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	noteURL, _ :=getString(args, "url")
	if noteURL == "" {
		return err("url parameter is required")
}

	indexList := getIndexList(args)

	client := http.DefaultClient

	// First resolve the actual note URL if it's a short link
	resolvedURL, resolveErr := resolveXHSURL(client, noteURL)
	if resolveErr != nil {
		return err(fmt.Sprintf("failed to resolve URL: %v", resolveErr))
}

	// Extract note ID from URL
	noteID, extractErr := extractNoteID(resolvedURL)
	if extractErr != nil {
		return err(fmt.Sprintf("failed to extract note ID: %v", extractErr))
}

	// Fetch note details
	noteInfo, fetchErr := fetchNoteDetails(client, noteID)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch note details: %v", fetchErr))
}

	downloadErr := downloadNoteFiles(client, noteInfo, indexList)
	if downloadErr != nil {
		return err(fmt.Sprintf("failed to download note files: %v", downloadErr))
}

	return ok(fmt.Sprintf("Successfully downloaded note %s", noteID))
}

func HandleBatchExtractNotes(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urls, _ :=getString(args, "urls")
	if urls == "" {
		return err("urls parameter is required")
}

	urlList := strings.Split(urls, " ")
	if len(urlList) == 0 {
		return err("no URLs provided")
}

	download, _ :=getBool(args, "download")
	client := http.DefaultClient

	var results []xhsNoteInfo
	var errorMessages []string

	for _, urlStr := range urlList {
		if urlStr == "" {
			continue
		}

		// First resolve the actual note URL if it's a short link
		resolvedURL, resolveErr := resolveXHSURL(client, urlStr)
		if resolveErr != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("failed to resolve URL %s: %v", urlStr, resolveErr))
			continue
		}

		// Extract note ID from URL
		noteID, extractErr := extractNoteID(resolvedURL)
		if extractErr != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("failed to extract note ID from %s: %v", urlStr, extractErr))
			continue
		}

		// Fetch note details
		noteInfo, fetchErr := fetchNoteDetails(client, noteID)
		if fetchErr != nil {
			errorMessages = append(errorMessages, fmt.Sprintf("failed to fetch note details for %s: %v", noteID, fetchErr))
			continue
		}

		if download {
			downloadErr := downloadNoteFiles(client, noteInfo, nil)
			if downloadErr != nil {
				errorMessages = append(errorMessages, fmt.Sprintf("failed to download note files for %s: %v", noteID, downloadErr))

		}

		results = append(results, *noteInfo)

	if len(errorMessages) > 0 {
		return ok(fmt.Sprintf("Completed with errors:\n%s\n\nSuccessfully processed %d notes",
}
			strings.Join(errorMessages, "\n"), len(results)))

	response, jsonErr := json.Marshal(results)
	if jsonErr != nil {
		return err(fmt.Sprintf("failed to marshal response: %v", jsonErr))
}

	return ok(string(response))
}

}
}

func resolveXHSURL(client *http.Client, urlStr string) (string, error) {
	// Check if it's a short URL
	if strings.Contains(urlStr, "xhslink.com") {
		req, reqErr := http.NewRequest("GET", urlStr, nil)
		if reqErr != nil {
			return "", fmt.Errorf("failed to create request: %v", reqErr)
}

		resp, respErr := client.Do(req)
		if respErr != nil {
			return "", fmt.Errorf("failed to resolve short URL: %v", respErr)
}

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusFound {
			return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

		return resp.Header.Get("Location"), nil
	}

	return urlStr, nil
}

func extractNoteID(urlStr string) (string, error) {
	matches := xhsURLPattern.FindStringSubmatch(urlStr)
	if len(matches) == 0 {
		return "", fmt.Errorf("invalid XHS URL format")
}

	// The note ID is either the 4th or 6th group in the regex
	if matches[4] != "" {
		return matches[4], nil
	}
	return matches[6], nil
}

func fetchNoteDetails(client *http.Client, noteID string) (*xhsNoteInfo, error) {
	// This is a simplified version - in a real implementation, you would:
	// 1. Make API requests to XHS endpoints
	// 2. Parse the JSON response
	// 3. Extract the note details

	// For this example, we'll return mock data
	return &xhsNoteInfo{
}
		ID:          noteID,
		Title:       "Sample Note Title",
		Description: "This is a sample note description",
		Author:      "SampleAuthor",
		Images:      []string{"https://example.com/image1.jpg", "https://example.com/image2.jpg"},
		Videos:      []string{"https://example.com/video1.mp4"},
		PublishTime: time.Now(),
	}, nil
}

func downloadNoteFiles(client *http.Client, note *xhsNoteInfo, indexList []int) error {
	// Create download directory
	dirName := filepath.Join("downloads", note.Author, note.ID)
	if e := os.MkdirAll(dirName, 0755); e != nil {
		return fmt.Errorf("failed to create directory: %v", e)
}

	// Download images
	if len(indexList) == 0 {
		indexList = makeRange(0, len(note.Images)-1)

	for _, idx := range indexList {
		if idx < 0 || idx >= len(note.Images) {
			continue
		}

		imageURL := note.Images[idx]
		fileName := filepath.Join(dirName, fmt.Sprintf("image_%d.jpg", idx+1))
		if e := downloadFile(client, imageURL, fileName); e != nil {
			return fmt.Errorf("failed to download image %d: %v", idx+1, e)

	}

	// Download videos
	for i, videoURL := range note.Videos {
		fileName := filepath.Join(dirName, fmt.Sprintf("video_%d.mp4", i+1))
		if e := downloadFile(client, videoURL, fileName); e != nil {
			return fmt.Errorf("failed to download video %d: %v", i+1, e)

	}

	// Download live photos
	for i, live := range note.LivePhotos {
		imageFile := filepath.Join(dirName, fmt.Sprintf("live_image_%d.jpg", i+1))
		videoFile := filepath.Join(dirName, fmt.Sprintf("live_video_%d.mp4", i+1))

		if e := downloadFile(client, live.Image, imageFile); e != nil {
			return fmt.Errorf("failed to download live photo image %d: %v", i+1, e)
}

		if e := downloadFile(client, live.Video, videoFile); e != nil {
			return fmt.Errorf("failed to download live photo video %d: %v", i+1, e)

	}

	return nil
}

}
}
}

func downloadFile(client *http.Client, fileURL, filePath string) error {
	resp, e := client.Get(fileURL)
	if e != nil {
		return fmt.Errorf("failed to download file: %v", e)
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %d", resp.StatusCode)
}

	out, e := os.Create(filePath)
	if e != nil {
		return fmt.Errorf("failed to create file: %v", e)
}

	defer out.Close()

	_, e = io.Copy(out, resp.Body)
	if e != nil {
		return fmt.Errorf("failed to save file: %v", e)
}

	return nil
}

func getIndexList(args map[string]interface{}) []int {
	var indexList []int
	if indexVal, found := args["index"].([]interface{}); found {
		for _, v := range indexVal {
			if i, found := v.(float64); found {
				indexList = append(indexList, int(i))

		}
		sort.Ints(indexList)

	return indexList
}

}
}

func makeRange(min, max int) []int {
	a := make([]int, max-min+1)
	for i := range a {
		a[i] = min + i
	}
	return a
}