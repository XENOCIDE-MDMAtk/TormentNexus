package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const chromaBaseURL = "http://localhost:8000"

var http.DefaultClient = http.DefaultClient

func HandleCreateCollection(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("collection name is required")
}

	body := map[string]string{"name": name}
	jsonBody, jsonErr := json.Marshal(body)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	req, reqErr := http.NewRequestWithContext(ctx, "POST", chromaBaseURL+"/api/v1/collections", bytes.NewBuffer(jsonBody))
	if reqErr != nil {
		return err(reqErr.Error())
}

	req.Header.Set("Content-Type", "application/json")

	resp, apiErr := http.DefaultClient.Do(req)
	if apiErr != nil {
		return err(apiErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return err(fmt.Sprintf("chroma error %d: %s", resp.StatusCode, string(b)))
}

	return ok(fmt.Sprintf("Collection '%s' created successfully.", name))
}