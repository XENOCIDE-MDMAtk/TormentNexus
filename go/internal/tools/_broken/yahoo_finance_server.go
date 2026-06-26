package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// #############################################################################
// HandleGetQuote
// #############################################################################

func HandleGetQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	client := http.DefaultClient
	u := fmt.Sprintf("https://query1.finance.yahoo.com/v8/finance/chart/%s?interval=1d&range=1d", url.PathEscape(symbol))
	req, reqErr := http.NewRequestWithContext(ctx, "GET", u, nil)
	if reqErr != nil {
		return err(fmt.Sprintf("failed to create request: %v", reqErr))
}

	req.Header.Set("User-Agent", "Mozilla/5.0")
	req.Header.Set("Accept", "application/json")

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch quote: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(fmt.Sprintf("failed to read response: %v", readErr))
}

	if resp.StatusCode != 200 {
		return err(fmt.Sprintf("Yahoo Finance API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if parseErr := json.Unmarshal(body, &result); parseErr != nil {
		return err(fmt.Sprintf("failed to parse JSON: %v", parseErr))
}

	chart, _ := result["chart"].(map[string]interface{})
	if chart == nil {
		return err("invalid response: no chart data")
}

	resultArr, _ := chart["result"].([]interface{})
	if len(resultArr) == 0 {
		return err("no result in chart data")
}

	firstResult, _ := resultArr[0].(map[string]interface{})
	if firstResult == nil {
		return err("invalid result structure")
}

	meta, _ := firstResult["meta"].(map[string]interface{})
	if meta == nil {
		return err("no meta in result")
}

	symbolName, _ := meta["symbol"].(string)
	regularMarketPrice, _ := meta["regularMarketPrice"].(float64)
	previousClose, _ := meta["chartPreviousClose"].(float64)
	currency, _ := meta["currency"].(string)
	exchangeName, _ := meta["exchangeName"].(string)

	timestamps, _ := firstResult["timestamp"].([]interface{})
	indicators, _ := firstResult["indicators"].(map[string]interface{})
	var closePrices []interface{}
	if indicators != nil {
		quoteArr, _ := indicators["quote"].([]interface{})
		if len(quoteArr) > 0 {
		}
	}
}