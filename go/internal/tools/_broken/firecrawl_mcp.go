
Wait, I removed `net/url` and `strconv` but kept `os`. Let me double check if `os` is used... Yes, for `os.Getenv`. Good.

Actually, I need to reconsider the imports. The reviewer flagged `net/url`, `os`, and `strconv` as unused, but I'm using `os` for the environment variable calls. Let me verify which ones are actually needed by checking the code for `url.Parse`, `strconv` conversions, and `os.Getenv` calls.

Since I'm calling `os.Getenv` for both `getAPIKey` and `getAPIURL`, `os` is necessary. The other two packages aren't referenced anywhere, so they should be removed. I also notice `defaultVersion` is declared but never used, which could be cleaned up, though the reviewer didn't specifically flag it. Let me finalize the import list and complete the function implementations.

package tools

import (
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "os"
    "strings"
    "time"
)

const (
    defaultAPIURL = "https://api.firecrawl.dev"
)

var http.DefaultClient = http.DefaultClient

func getAPIKey() string {
    return os.Getenv("FIRECRAWL_API_KEY")
}

func getAPIURL() string {
    if url := os.Getenv("FIRECRAWL_API_URL"); url != "" {
        return url
    }
    return defaultAPIURL
}

func HandleMap(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url parameter is required")
}

    apiKey := getAPIKey()
    if apiKey == "" {
        return err("FIRECRAWL_API_KEY is not set")
}

    baseURL := getAPIURL()
    apiURL := baseURL + "/v1/map"

    bodyMap := map[string]interface{}{
        "url": url,
    }
    if search := getString(args, "search"); search != "" {
        bodyMap["search"] = search
    }
    if limit := getInt(args, "limit"); limit > 0 {
        bodyMap["limit"] = limit
    }

    reqBody, marshalErr := json.Marshal(bodyMap)
    if marshalErr != nil {
        return err(fmt.Sprintf("marshal error: %v", marshalErr))
}

    req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(reqBody)))
    if reqErr != nil {
        return err(fmt.Sprintf("request creation error: %v", reqErr))
}

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+apiKey)

    resp, doErr := http.DefaultClient.Do(req)
    if doErr != nil {
        return err(fmt.Sprintf("API request failed: %v", doErr))
}

    defer resp.Body.Close()

    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return err(fmt.Sprintf("response read error: %v", readErr))
}

    if resp.StatusCode != 200 {
        return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)))
}

    var apiResp struct {
        Success bool     `json:"success"`
        Links   []string `json:"links"`
    }
    if parseErr := json.Unmarshal(body, &apiResp); parseErr != nil {
        return err(fmt.Sprintf("parse error: %v", parseErr))
}

    if !apiResp.Success {
        return err("API returned unsuccessful response")
}

    if len(apiResp.Links) == 0 {
        return ok("No links found.")
}

    return ok(strings.Join(apiResp.Links, "\n"))
}

func HandleCrawl(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    url, _ :=getString(args, "url")
    if url == "" {
        return err("url parameter is required")
}

    apiKey := getAPIKey()
    if apiKey == "" {
        return err("FIRECRAWL_API_KEY is not set")
}

    baseURL := getAPIURL()
    apiURL := baseURL + "/v2/crawl"

    bodyMap := map[string]interface{}{
        "url": url,
    }
    if maxDepth := getInt(args, "maxDepth"); maxDepth > 0 {
        bodyMap["maxDepth"] = maxDepth
    }
    if maxPages := getInt(args, "maxPages"); maxPages > 0 {
        bodyMap["maxPages"] = maxPages
    }

    reqBody, marshalErr := json.Marshal(bodyMap)
    if marshalErr != nil {
        return err(fmt.Sprintf("marshal error: %v", marshalErr))
}

    req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(reqBody)))
    if reqErr != nil {
        return err(fmt.Sprintf("request creation error: %v", reqErr))
}

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+apiKey)

    resp, doErr := http.DefaultClient.Do(req)
    if doErr != nil {
        return err(fmt.Sprintf("API request failed: %v", doErr))
}

    defer resp.Body.Close()

    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return err(fmt.Sprintf("response read error: %v", readErr))
}

    if resp.StatusCode != 200 {
        return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)))
}

    var apiResp struct {
        Success bool `json:"success"`
        Data    struct {
            ID string `json:"id"`
        } `json:"data"`
    }
    if parseErr := json.Unmarshal(body, &apiResp); parseErr != nil {
        return err(fmt.Sprintf("parse error: %v", parseErr))
}

    if !apiResp.Success {
        return err("API returned unsuccessful response")
}

    crawlID := apiResp.Data.ID
    statusURL := baseURL + "/v2/crawl/" + crawlID

    return ok(fmt.Sprintf("Crawl started. ID: %s\nCheck status at: %s", crawlID, statusURL))
}

func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    query, _ :=getString(args, "query")
    if query == "" {
        return err("query parameter is required")
}

    apiKey := getAPIKey()
    if apiKey == "" {
        return err("FIRECRAWL_API_KEY is not set")
}

    baseURL := getAPIURL()
    apiURL := baseURL + "/v2/search"

    bodyMap := map[string]interface{}{
        "query": query,
    }
    if limit := getInt(args, "limit"); limit > 0 {
        bodyMap["limit"] = limit
    }

    reqBody, marshalErr := json.Marshal(bodyMap)
    if marshalErr != nil {
        return err(fmt.Sprintf("marshal error: %v", marshalErr))
}

    req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(reqBody)))
    if reqErr != nil {
        return err(fmt.Sprintf("request creation error: %v", reqErr))
}

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+apiKey)

    resp, doErr := http.DefaultClient.Do(req)
    if doErr != nil {
        return err(fmt.Sprintf("API request failed: %v", doErr))
}

    defer resp.Body.Close()

    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return err(fmt.Sprintf("response read error: %v", readErr))
}

    if resp.StatusCode != 200 {
        return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)))
}

    var apiResp struct {
        Success bool `json:"success"`
        Data    []struct {
            URL     string `json:"url"`
            Title   string `json:"title"`
            Content string `json:"content"`
        } `json:"data"`
    }
    if parseErr := json.Unmarshal(body, &apiResp); parseErr != nil {
        return err(fmt.Sprintf("parse error: %v", parseErr))
}

    if !apiResp.Success {
        return err("API returned unsuccessful response")
}

    if len(apiResp.Data) == 0 {
        return ok("No search results found.")
}

    var sb strings.Builder
    for _, result := range apiResp.Data {
        sb.WriteString(fmt.Sprintf("URL: %s\nTitle: %s\nContent: %s\n\n", result.URL, result.Title, result.Content))

    return ok(sb.String())
}

}

func HandleScrape(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    urlToScrape, _ :=getString(args, "url")
    if urlToScrape == "" {
        return err("url parameter is required")
}

    apiKey := getAPIKey()
    if apiKey == "" {
        return err("FIRECRAWL_API_KEY is not set")
}

    baseURL := getAPIURL()
    apiURL := baseURL + "/v2/scrape"

    bodyMap := map[string]interface{}{
        "url": urlToScrape,
    }
    if formats := getString(args, "formats"); formats != "" {
        bodyMap["formats"] = strings.Split(formats, ",")

    if onlyMain := getBool(args, "onlyMainContent"); onlyMain {
        bodyMap["onlyMainContent"] = true
    }

    reqBody, marshalErr := json.Marshal(bodyMap)
    if marshalErr != nil {
        return err(fmt.Sprintf("failed to marshal request: %v", marshalErr))
}

    req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(reqBody)))
    if reqErr != nil {
        return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+apiKey)

    resp, doErr := http.DefaultClient.Do(req)
    if doErr != nil {
        return err(fmt.Sprintf("API request failed: %v", doErr))
}

    defer resp.Body.Close()

    body, readErr := io.ReadAll(resp.Body)
    if readErr != nil {
        return err(fmt.Sprintf("failed to read response: %v", readErr))
}

    if resp.StatusCode != 200 {
        return err(fmt.Sprintf("API error (status %d): %s", resp.StatusCode, string(body)))
}

    var apiResp struct {
        Success bool `json:"success"`
        Data    struct {
            Markdown string `json:"markdown"`
        } `json:"data"`
    }
    if parseErr := json.Unmarshal(body, &apiResp); parseErr != nil {
        return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

    if !apiResp.Success {
        return err("API returned unsuccessful response")
}

    if apiResp.Data.Markdown == "" {
        return ok("No content extracted.")
}

    return ok(apiResp.Data.Markdown)
}
}