package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var viperBaseURL string

func init() {
	viperBaseURL = "http://localhost:8080"
}

type ViperStatus struct {
	Version      string `json:"version"`
	Status       string `json:"status"`
	SessionCount int    `json:"session_count"`
	HostCount    int    `json:"host_count"`
}

type ViperSession struct {
	ID          string `json:"id"`
	HostID      string `json:"host_id"`
	Type        string `json:"type"`
	InternalIP  string `json:"internal_ip"`
	ExternalIP  string `json:"external_ip"`
	User        string `json:"user"`
	PID         int    `json:"pid"`
	Online      bool   `json:"online"`
	CheckInTime string `json:"checkin_time"`
}

type ViperHost struct {
	ID           string `json:"id"`
	InternalIP   string `json:"internal_ip"`
	ExternalIP   string `json:"external_ip"`
	Hostname     string `json:"hostname"`
	OS           string `json:"os"`
	Arch         string `json:"arch"`
	SessionCount int    `json:"session_count"`
	FirstSeen    string `json:"first_seen"`
	LastSeen     string `json:"last_seen"`
}

type ViperModule struct {
	Name        string `json:"name"`
	ModuleType  string `json:"module_type"`
	Description string `json:"description"`
	Author      string `json:"author"`
	MITRE       string `json:"mitre"`
}

func HandleViperStatus(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	if serverURL == "" {
		serverURL = viperBaseURL
	}

	client := http.DefaultClient
	apiURL := fmt.Sprintf("%s/api/v1/status", serverURL)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Failed to connect to Viper server at %s: %v", serverURL, fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var status ViperStatus
		if decodeErr := json.NewDecoder(resp.Body).Decode(&status); decodeErr != nil {
			return err(decodeErr.Error())
}

		result := fmt.Sprintf("Viper Status: %s\nVersion: %s\nSessions: %d\nHosts: %d",
			status.Status, status.Version, status.SessionCount, status.HostCount)
		return ok(result)
}

	return err(fmt.Sprintf("Viper server returned status: %d", resp.StatusCode))
}

func HandleViperSessions(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	if serverURL == "" {
		serverURL = viperBaseURL
	}
	onlineOnly, _ :=getBool(args, "online_only")

	client := http.DefaultClient
	apiURL := fmt.Sprintf("%s/api/v1/sessions", serverURL)

	if onlineOnly {
		v := url.Values{}
		v.Set("online", "true")
		apiURL = apiURL + "?" + v.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Failed to fetch sessions: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var sessions []ViperSession
		if decodeErr := json.NewDecoder(resp.Body).Decode(&sessions); decodeErr != nil {
			return err(decodeErr.Error())
}

		if len(sessions) == 0 {
			return ok("No sessions found")
}

		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("Found %d session(s):\n\n", len(sessions)))
		for _, s := range sessions {
			status := "OFFLINE"
			if s.Online {
				status = "ONLINE"
			}
			builder.WriteString(fmt.Sprintf("ID: %s | Host: %s | Type: %s | IP: %s | User: %s | Status: %s\n",
				s.ID, s.HostID, s.Type, s.InternalIP, s.User, status))

		return ok(builder.String())
}

	return err(fmt.Sprintf("Failed to get sessions: HTTP %d", resp.StatusCode))
}

}
}

func HandleViperHosts(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	if serverURL == "" {
		serverURL = viperBaseURL
	}

	client := http.DefaultClient
	apiURL := fmt.Sprintf("%s/api/v1/hosts", serverURL)

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Failed to fetch hosts: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var hosts []ViperHost
		if decodeErr := json.NewDecoder(resp.Body).Decode(&hosts); decodeErr != nil {
			return err(decodeErr.Error())
}

		if len(hosts) == 0 {
			return ok("No hosts found")
}

		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("Found %d host(s):\n\n", len(hosts)))
		for _, h := range hosts {
			builder.WriteString(fmt.Sprintf("ID: %s | IP: %s | Hostname: %s | OS: %s | Sessions: %d\n",
				h.ID, h.InternalIP, h.Hostname, h.OS, h.SessionCount))

		return ok(builder.String())
}

	return err(fmt.Sprintf("Failed to get hosts: HTTP %d", resp.StatusCode))
}

}

func HandleViperModules(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	if serverURL == "" {
		serverURL = viperBaseURL
	}
	moduleType, _ :=getString(args, "type")
	mitre, _ :=getString(args, "mitre")

	client := http.DefaultClient
	apiURL := fmt.Sprintf("%s/api/v1/modules", serverURL)

	v := url.Values{}
	if moduleType != "" {
		v.Set("type", moduleType)

	if mitre != "" {
		v.Set("mitre", mitre)

	if len(v) > 0 {
		apiURL = apiURL + "?" + v.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Failed to fetch modules: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var modules []ViperModule
		if decodeErr := json.NewDecoder(resp.Body).Decode(&modules); decodeErr != nil {
			return err(decodeErr.Error())
}

		if len(modules) == 0 {
			return ok("No modules found")
}

		var builder strings.Builder
		builder.WriteString(fmt.Sprintf("Found %d module(s):\n\n", len(modules)))
		for _, m := range modules {
			builder.WriteString(fmt.Sprintf("[%s] %s\n  Type: %s | MITRE: %s\n  Author: %s\n  %s\n\n",
				m.Name, m.ModuleType, m.MITRE, m.Author, m.Description))

		return ok(builder.String())
}

	return err(fmt.Sprintf("Failed to get modules: HTTP %d", resp.StatusCode))
}

}
}
}
}

func HandleViperRunModule(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	if serverURL == "" {
		serverURL = viperBaseURL
	}
	sessionID, _ :=getString(args, "session_id")
	moduleName, _ :=getString(args, "module_name")
	options, _ :=getString(args, "options")

	if sessionID == "" {
		return err("session_id is required")
}

	if moduleName == "" {
		return err("module_name is required")
}

	client := http.DefaultClient
	apiURL := fmt.Sprintf("%s/api/v1/modules/execute", serverURL)

	payload := map[string]interface{}{
		"session_id":  sessionID,
		"module_name": moduleName,
	}
	if options != "" {
		payload["options"] = options
	}

	body, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(body)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Failed to execute module: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
		var result map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
			return ok("Module execution started successfully")
}

		if msg, found := result["message"].(string); found {
			return ok(msg)
}

		return ok("Module executed successfully")
}

	return err(fmt.Sprintf("Module execution failed: HTTP %d", resp.StatusCode))
}

func HandleViperCredentials(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	if serverURL == "" {
		serverURL = viperBaseURL
	}
	credType, _ :=getString(args, "type")

	client := http.DefaultClient
	apiURL := fmt.Sprintf("%s/api/v1/credentials", serverURL)

	v := url.Values{}
	if credType != "" {
		v.Set("type", credType)

	if len(v) > 0 {
		apiURL = apiURL + "?" + v.Encode()

	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Failed to fetch credentials: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var creds map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&creds); decodeErr != nil {
			return err(decodeErr.Error())
}

		credJSON, jsonErr := json.MarshalIndent(creds, "", "  ")
		if jsonErr != nil {
			return err(jsonErr.Error())
}

		return ok("Credentials:\n" + string(credJSON))
}

	return err(fmt.Sprintf("Failed to get credentials: HTTP %d", resp.StatusCode))
}

}
}

func HandleViperRoute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	serverURL, _ :=getString(args, "server_url")
	if serverURL == "" {
		serverURL = viperBaseURL
	}
	action, _ :=getString(args, "action")
	sessionID, _ :=getString(args, "session_id")
	targetHost, _ :=getString(args, "target_host")
	targetPort, _ :=getString(args, "target_port")

	client := http.DefaultClient
	apiURL := fmt.Sprintf("%s/api/v1/route", serverURL)

	var payload map[string]interface{}
	if action == "add" {
		if sessionID == "" || targetHost == "" || targetPort == "" {
			return err("session_id, target_host, and target_port are required for add action")
}

		port, portErr := strconv.Atoi(targetPort)
		if portErr != nil {
			return err("target_port must be a valid number")
}

		payload = map[string]interface{}{
			"action":      "add",
			"session_id":  sessionID,
			"target_host": targetHost,
			"target_port": port,
		}
	} else if action == "remove" || action == "list" {
		payload = map[string]interface{}{
			"action": action,
		}
		if sessionID != "" {
			payload["session_id"] = sessionID
		}
	} else {
		return err("action must be 'add', 'remove', or 'list'")
}

	body, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", apiURL, strings.NewReader(string(body)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("Failed to manage route: %v", fetchErr))
}

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var result map[string]interface{}
		if decodeErr := json.NewDecoder(resp.Body).Decode(&result); decodeErr != nil {
			return ok("Route operation completed")
}

		if msg, found := result["message"].(string); found {
			return ok(msg)
}

		resultJSON, jsonErr := json.MarshalIndent(result, "", "  ")
		if jsonErr != nil {
			return ok("Route operation completed")
}

		return ok(string(resultJSON))
}

	return err(fmt.Sprintf("Route operation failed: HTTP %d", resp.StatusCode))
}