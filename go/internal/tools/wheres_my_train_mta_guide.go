package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
)

// TextContent is defined in registry.go
func HandleTrainStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	trainLine, _ := getString(args, "train_line")
	if trainLine == "" {
		return err("missing required parameter: train_line")
	}

	apiURL := "https://api.mta.info/train-status"
	client := http.DefaultClient

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
	}

	query := url.Values{}
	query.Set("line", trainLine)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Authorization", "Bearer "+os.Getenv("MTA_API_KEY"))

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %d", resp.StatusCode))
	}

	var result map[string]interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)
	if decodeErr != nil {
		return err(decodeErr.Error())
	}

	status, found := result["status"].(string)
	if !found {
		return err("unexpected response format: missing 'status'")
	}

	return ok(status)
}

func HandleTrainSchedule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	station, _ := getString(args, "station")
	if station == "" {
		return err("missing required parameter: station")
	}

	apiURL := "https://api.mta.info/train-schedule"
	client := http.DefaultClient

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
	}

	query := url.Values{}
	query.Set("station", station)
	req.URL.RawQuery = query.Encode()
	req.Header.Set("Authorization", "Bearer "+os.Getenv("MTA_API_KEY"))

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %d", resp.StatusCode))
	}

	var result map[string]interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)
	if decodeErr != nil {
		return err(decodeErr.Error())
	}

	schedule, found := result["schedule"].(string)
	if !found {
		return err("unexpected response format: missing 'schedule'")
	}

	return ok(schedule)
}

func HandleTrainDelays(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	apiURL := "https://api.mta.info/train-delays"
	client := http.DefaultClient

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
	}

	req.Header.Set("Authorization", "Bearer "+os.Getenv("MTA_API_KEY"))

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %d", resp.StatusCode))
	}

	var result map[string]interface{}
	decodeErr := json.NewDecoder(resp.Body).Decode(&result)
	if decodeErr != nil {
		return err(decodeErr.Error())
	}

	delays, found := result["delays"].(string)
	if !found {
		return err("unexpected response format: missing 'delays'")
	}

	return ok(delays)
}
