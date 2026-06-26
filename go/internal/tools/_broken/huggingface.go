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
)

// HandleModel returns information about a Hugging Face model.
func HandleModel(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	modelName, _ :=getString(args, "model_name")
	if modelName == "" {
		return err("model_name is required")
}

	req, reqErr := http.NewRequest("GET", "https://api.huggingface.co/models/"+modelName, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to retrieve model information: %d", resp.StatusCode))
}

	var modelInfo map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&modelInfo)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(fmt.Sprintf("Model information: %+v", modelInfo))
}

// HandleDataset returns information about a Hugging Face dataset.
func HandleDataset(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	datasetName, _ :=getString(args, "dataset_name")
	if datasetName == "" {
		return err("dataset_name is required")
}

	req, reqErr := http.NewRequest("GET", "https://api.huggingface.co/datasets/"+datasetName, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("failed to retrieve dataset information: %d", resp.StatusCode))
}

	var datasetInfo map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&datasetInfo)
	if parseErr != nil {
		return err(parseErr.Error())
}

	return ok(fmt.Sprintf("Dataset information: %+v", datasetInfo))
}

===GO_FILE===
package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// HandleModel returns information about a Hugging Face model.
// Expected argument: "model_name" (string)
func HandleSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	limit, _ :=getInt(args, "limit")
	if limit <= 0 {
		limit = 5
	}

	// Build query parameters using url.Values
	values := url.Values{}
	values.Set("search", query)
	values.Set("limit", fmt.Sprintf("%d", limit))

	searchURL := fmt.Sprintf("https://api.huggingface.co/models?%s", values.Encode())
	req, reqErr := http.NewRequest("GET", searchURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	req = req.WithContext(ctx)

	client := http.DefaultClient
	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("search request failed: %d", resp.StatusCode))
}

	var results []map[string]interface{}
	parseErr := json.NewDecoder(resp.Body).Decode(&results)
	if parseErr != nil {
		return err(parseErr.Error())
}

	// Simple textual representation of results
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Top %d results for \"%s\":\n", limit, query))
	for i, item := range results {
		if i >= limit {
			break
		}
		name, _ := item["modelId"].(string)
		desc, _ := item["pipeline_tag"].(string)
		sb.WriteString(fmt.Sprintf("%d. %s – %s\n", i+1, name, desc))

	return ok(sb.String())
}
}