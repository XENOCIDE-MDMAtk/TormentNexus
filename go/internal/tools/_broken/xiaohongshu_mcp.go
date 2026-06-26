package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func HandleSearchXiaohongshu(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ :=getString(args, "keyword")
	if keyword == "" {
		return err("keyword is required")
}

	page, _ :=getInt(args, "page")
	if page <= 0 {
		page = 1
	}

	pageSize, _ :=getInt(args, "page_size")
	if pageSize <= 0 {
		pageSize = 10
	}

	values := url.Values{}
	values.Set("keyword", keyword)
	values.Set("page", strconv.Itoa(page))
	values.Set("page_size", strconv.Itoa(pageSize))
	searchURL := "https://www.xiaohongshu.com/api/sns/v1/search/notes?" + values.Encode()

	client := http.DefaultClient
	req, reqErr := http.NewRequest("GET", searchURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.xiaohongshu.com/")
	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("search request failed with status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var searchResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Notes []struct {
				ID    string   `json:"id"`
				Title string   `json:"title"`
				Desc  string   `json:"desc"`
				User  struct {
					Nickname string `json:"nickname"`
				} `json:"user"`
				Likes  int      `json:"likes"`
				Images []string `json:"images"`
			} `json:"notes"`
			HasMore bool `json:"has_more"`
		} `json:"data"`
	}
	parseErr := json.Unmarshal(body, &searchResp)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if searchResp.Code != 0 {
		return err(searchResp.Msg)
}

	var result strings.Builder
	result.WriteString(fmt.Sprintf("Search results for keyword \"%s\" (page %d, total %d notes):\n", keyword, page, len(searchResp.Data.Notes)))
	for i, note := range searchResp.Data.Notes {
		result.WriteString(fmt.Sprintf("%d. Title: %s\n   Author: %s\n   Likes: %d\n   Link: https://www.xiaohongshu.com/explore?note_id=%s\n", i+1, note.Title, note.User.Nickname, note.Likes, note.ID))
		if len(note.Images) > 0 {
			result.WriteString(fmt.Sprintf("   Images: %s\n", strings.Join(note.Images, ", ")))

		result.WriteString("\n")

	if !searchResp.Data.HasMore {
		result.WriteString("No more results.\n")

	return ok(result.String())
}

}
}
}

func HandleGetXiaohongshuNote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	noteID, _ :=getString(args, "note_id")
	if noteID == "" {
		return err("note_id is required")
}

	noteURL := fmt.Sprintf("https://www.xiaohongshu.com/api/sns/v1/note/%s", noteID)

	client := http.DefaultClient
	req, reqErr := http.NewRequest("GET", noteURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Referer", "https://www.xiaohongshu.com/")
	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to get note, status code: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var noteResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Note struct {
				ID      string `json:"id"`
				Title   string `json:"title"`
				Desc    string `json:"desc"`
				User    struct {
					Nickname string `json:"nickname"`
					UserID   string `json:"user_id"`
				} `json:"user"`
				Likes    int      `json:"likes"`
				Collects int      `json:"collects"`
				Comments int      `json:"comments"`
				Images   []string `json:"images"`
				VideoURL string   `json:"video_url"`
			} `json:"note"`
		} `json:"data"`
	}
	parseErr := json.Unmarshal(body, &noteResp)
	if parseErr != nil {
		return err(parseErr.Error())
}

	if noteResp.Code != 0 {
		return err(noteResp.Msg)
}

	note := noteResp.Data.Note
	var result strings.Builder
	result.WriteString(fmt.Sprintf("Note Details:\nTitle: %s\nAuthor: %s\nLikes: %d, Collects: %d, Comments: %d\nContent: %s\n", note.Title, note.User.Nickname, note.Likes, note.Collects, note.Comments, note.Desc))
	if len(note.Images) > 0 {
		result.WriteString(fmt.Sprintf("Images: %s\n", strings.Join(note.Images, ", ")))

	if note.VideoURL != "" {
		result.WriteString(fmt.Sprintf("Video: %s\n", note.VideoURL))

	result.WriteString(fmt.Sprintf("Link: https://www.xiaohongshu.com/explore?note_id=%s\n", noteID))

	return ok(result.String())
}
}
}