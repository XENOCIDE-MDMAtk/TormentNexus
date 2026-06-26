package tools".")

Also note: The original code has a comment at the top: "PRE-DEFINED in parity.go (do NOT redeclare): ...". That comment is part of the original code? Actually it's a meta-instruction. The code provided is the original code. We need to output the fixed code. We should not include the meta-instruction, just the code.

Thus, output the code from "package tools" to the end, with the import fixed.

Let's produce the final answer.package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const exaBaseURL = "https://api.exa.ai"

func getExaAPIKey() string {
	key := os.Getenv("EXA_API_KEY")
	return key
}

func makeExaRequest(ctx context.Context, method, path string, body io.Reader) (*http.Response, error) {
	client := http.DefaultClient
	fullURL := exaBaseURL + path

	req, reqErr := http.NewRequestWithContext(ctx, method, fullURL, body)
	if reqErr != nil {
		return nil, reqErr
	}

	req.Header.Set("Authorization", "Bearer "+getExaAPIKey())
	req.Header.Set("Content-Type", "application/json")

	return client.Do(req)
}

// HandleExaSearch performs a web search using the Exa API
func HandleExaSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	numResults, _ :=getInt(args, "num_results")
	if numResults <= 0 {
		numResults = 10
	}

	includeDomains := []string{}
	if domainsArg, found := args["include_domains"]; found {
		if domainsSlice, found := domainsArg.([]interface{}); found {
			for _, d := range domainsSlice {
				if ds, found := d.(string); found {
					includeDomains = append(includeDomains, ds)

			}
		}
	}

	excludeDomains := []string{}
	if domainsArg, found := args["exclude_domains"]; found {
		if domainsSlice, found := domainsArg.([]interface{}); found {
			for _, d := range domainsSlice {
				if ds, found := d.(string); found {
					excludeDomains = append(excludeDomains, ds)

			}
		}
	}

	reqBody := map[string]interface{}{
		"query":           query,
		"numResults":      numResults,
		"includeDomains":  includeDomains,
		"excludeDomains":  excludeDomains,
	}

	if useAutoprompt, found := args["use_autoprompt"]; found {
		reqBody["useAutoprompt"] = getBool(useAutoprompt.(map[string]interface{}), "use_autoprompt")

	jsonBody, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	resp, apiErr := makeExaRequest(ctx, "POST", "/search", jsonBody)
	if apiErr != nil {
		return err(apiErr.Error())
}

	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Exa API error: %s - %s", resp.Status, string(bodyBytes)))
}

	return ok(string(bodyBytes))
}

}
}
}

// HandleExaFindSimilar finds similar pages to a given URL using the Exa API
func HandleExaFindSimilar(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr, _ :=getString(args, "url")
	if urlStr == "" {
		return err("url parameter is required")
}

	numResults, _ :=getInt(args, "num_results")
	if numResults <= 0 {
		numResults = 10
	}

	reqBody := map[string]interface{}{
		"url":        urlStr,
		"numResults": numResults,
	}

	if excludeSource, found := args["exclude_source"]; found {
		reqBody["excludeSource"] = getBool(args, "exclude_source")

	jsonBody, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	resp, apiErr := makeExaRequest(ctx, "POST", "/findSimilar", jsonBody)
	if apiErr != nil {
		return err(apiErr.Error())
}

	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Exa API error: %s - %s", resp.Status, string(bodyBytes)))
}

	return ok(string(bodyBytes))
}

}

// HandleExaGetContents retrieves the contents of specific URLs using the Exa API
func HandleExaGetContents(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	idsArg, found := args["ids"]
	if !found {
		return err("ids parameter is required")
}

	ids := []string{}
	if idsSlice, found := idsArg.([]interface{}); found {
		for _, id := range idsSlice {
			if idStr, found := id.(string); found {
				ids = append(ids, idStr)

		}
	}

	if len(ids) == 0 {
		return err("at least one id is required")
}

	reqBody := map[string]interface{}{
		"ids": ids,
	}

	if text, found := args["text"]; found {
		reqBody["text"] = getBool(text.(map[string]interface{}), "text")

	if highlights, found := args["highlights"]; found {
		reqBody["highlights"] = getBool(highlights.(map[string]interface{}), "highlights")

	jsonBody, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	resp, apiErr := makeExaRequest(ctx, "POST", "/contents", jsonBody)
	if apiErr != nil {
		return err(apiErr.Error())
}

	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Exa API error: %s - %s", resp.Status, string(bodyBytes)))
}

	return ok(string(bodyBytes))
}

}
}
}

// HandleExaAnswer generates an answer to a query using the Exa API
func HandleExaAnswer(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	reqBody := map[string]interface{}{
		"query": query,
	}

	if model, found := args["model"]; found {
		reqBody["model"] = model.(string)

	jsonBody, marshalErr := json.Marshal(reqBody)
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	resp, apiErr := makeExaRequest(ctx, "POST", "/answer", jsonBody)
	if apiErr != nil {
		return err(apiErr.Error())
}

	defer resp.Body.Close()

	bodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("Exa API error: %s - %s", resp.Status, string(bodyBytes)))
}

	return ok(string(bodyBytes))
}
}