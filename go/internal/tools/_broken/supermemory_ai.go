package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

func HandleX(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	key, _ :=getString(args, "key")
	if key == "" {
		return err("キーが見つかりません")
}

	resp, fetchErr := fetchData(key)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	processedData, processErr := processData(resp)
	if processErr != nil {
		return err(processErr.Error())
}

	return ok(processedData)
}

func fetchData(key string) (string, error) {
	client := http.DefaultClient
	req, reqErr := http.NewRequest("GET", "https://example.com/data", nil)
	if reqErr != nil {
		return "", reqErr
	}

	q := req.URL.Query()
	q.Add("key", key)
	req.URL.RawQuery = q.Encode()

	resp, sendErr := client.Do(req)
	if sendErr != nil {
		return "", sendErr
	}
	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return "", readErr
	}

	return string(body), nil
}

func processData(data string) (string, error) {
	var parsedData map[string]interface{}
	parseErr := json.Unmarshal([]byte(data), &parsedData)
	if parseErr != nil {
		return "", parseErr
	}

	processed := fmt.Sprintf("Processed: %s", parsedData["value"].(string))
	return processed, nil
}

func HandleY(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("HandleY response")
}

func HandleZ(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("HandleZ response")
}

func HandleA(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("HandleA response")
}

func HandleB(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	return ok("HandleB response")
}