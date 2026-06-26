package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

var (
	mindsdbBaseURL = func() string {
		if v := os.Getenv("MINDSDB_API_URL"); v != "" {
			return strings.TrimRight(v, "/")
}

		return "http://localhost:47334"
	}()
	http.DefaultClient = http.DefaultClient
)

// HandleListModels returns a list of models available in MindsDB.
func HandleListModels(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	// Optional project filter
	project, _ :=getString(args, "project")
	endpoint := fmt.Sprintf("%s/api/models", mindsdbBaseURL)

	if project != "" {
		values := url.Values{}
		values.Add("project", project)
		endpoint = fmt.Sprintf("%s?%s", endpoint, values.Encode())

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("unexpected status: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

}

// HandlePredict runs a prediction against a specified model.
func HandlePredict(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	if model == "" {
		return err("model name is required")
}

	inputJSON, _ :=getString(args, "input")
	if inputJSON == "" {
		return err("input JSON is required")
}

	endpoint := fmt.Sprintf("%s/api/predict", mindsdbBaseURL)

	payload := map[string]interface{}{
		"model": model,
		"input": json.RawMessage(inputJSON),
	}
	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payloadBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("prediction failed with status: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

// HandleSQLQuery executes a raw SQL query against MindsDB.
func HandleSQLQuery(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "sql")
	if query == "" {
		return err("sql query is required")
}

	project, _ :=getString(args, "project")

	endpoint := fmt.Sprintf("%s/api/sql", mindsdbBaseURL)

	payload := map[string]string{
		"query": query,
	}
	if project != "" {
		payload["project"] = project
	}
	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payloadBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("sql query failed with status: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}

// HandleTrainModel triggers a training job for a given model.
func HandleTrainModel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	model, _ :=getString(args, "model")
	if model == "" {
		return err("model name is required")
}

	query, _ :=getString(args, "query")
	if query == "" {
		return err("training query is required")
}

	project, _ :=getString(args, "project")

	endpoint := fmt.Sprintf("%s/api/train", mindsdbBaseURL)

	payload := map[string]string{
		"model": model,
		"query": query,
	}
	if project != "" {
		payload["project"] = project
	}
	payloadBytes, marshalErr := json.Marshal(payload)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(string(payloadBytes)))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, fetchErr := http.DefaultClient.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		return err(fmt.Sprintf("train request failed with status: %d", resp.StatusCode))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(body))
}