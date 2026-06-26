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

const baseURL = "https://api.example.com"

var client = http.DefaultClient

func HandleExecuteCSharp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if !found {
		return err("missing 'code' parameter")
}

	reqBody, marshalErr := json.Marshal(map[string]string{"code": code})
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", marshalErr))
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", baseURL+"/execute", strings.NewReader(string(reqBody)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to execute request: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: %s", resp.Status))
}

	var unityResp struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if parseErr := json.Unmarshal(body, &unityResp); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	if unityResp.Error != "" {
		return err(unityResp.Error)
}

	return ok(unityResp.Result)
}

func HandleGetScene(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	req, reqErr := http.NewRequestWithContext(ctx, "GET", baseURL+"/scene", nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to execute request: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: %s", resp.Status))
}

	var unityResp struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if parseErr := json.Unmarshal(body, &unityResp); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	if unityResp.Error != "" {
		return err(unityResp.Error)
}

	return ok(unityResp.Result)
}

func HandleGetGameObject(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var reqURL string
	var reqErr error

	if name := getString(args, "name"); found {
		reqURL = fmt.Sprintf("%s/gameobject?name=%s", baseURL, url.QueryEscape(name))
	} else if id := getInt(args, "id"); found {
		reqURL = fmt.Sprintf("%s/gameobject?id=%d", baseURL, id)
	} else {
		return err("missing 'name' or 'id' parameter")
}

	req, reqErr := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to execute request: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: %s", resp.Status))
}

	var unityResp struct {
		Result json.RawMessage `json:"result"`
		Error  string          `json:"error"`
	}
	if parseErr := json.Unmarshal(body, &unityResp); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	if unityResp.Error != "" {
		return err(unityResp.Error)
}

	return ok(string(unityResp.Result))
}

func HandleSetGameObjectProperty(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	id, _ :=getInt(args, "gameobject_id")
	if !found {
		return err("missing 'gameobject_id' parameter")
}

	property, _ :=getString(args, "property")
	if !found {
		return err("missing 'property' parameter")
}

	value, found := args["value"]
	if !found {
		return err("missing 'value' parameter")
}

	reqBody, marshalErr := json.Marshal(map[string]interface{}{
		"property": property,
		"value":    value,
	})
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal request: %v", marshalErr))
}

	reqURL := fmt.Sprintf("%s/gameobject/%d/property", baseURL, id)
	req, reqErr := http.NewRequestWithContext(ctx, "POST", reqURL, strings.NewReader(string(reqBody)))
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to execute request: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("HTTP error: %s", resp.Status))
}

	var unityResp struct {
		Result string `json:"result"`
		Error  string `json:"error"`
	}
	if parseErr := json.Unmarshal(body, &unityResp); parseErr != nil {
		return err(fmt.Sprintf("failed to parse response: %v", parseErr))
}

	if unityResp.Error != "" {
		return err(unityResp.Error)
}

	return ok(unityResp.Result)
}