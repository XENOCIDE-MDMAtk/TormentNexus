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

const yfinanceBaseURL = "https://query1.finance.yahoo.com/v8/finance/chart/"

func HandleYahooFinanceQuote(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	symbol, _ :=getString(args, "symbol")
	if symbol == "" {
		return err("symbol is required")
}

	period, _ :=getString(args, "period")
	if period == "" {
		period = "1d"
	}

	interval, _ :=getString(args, "interval")
	if interval == "" {
		interval = "1m"
	}

	apiURL := fmt.Sprintf("%s%s?period=%s&interval=%s", yfinanceBaseURL, symbol, period, interval)

	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Get(apiURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var result map[string]interface{}
	parseErr := json.Unmarshal(body, &result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	chart, found := result["chart"].(map[string]interface{})
	if !found {
		return err("invalid API response format")
}

	indicators, found := chart["indicators"].(map[string]interface{})
	if !found {
		return err("invalid indicators format")
}

	quote, found := indicators["quote"].([]interface{})
	if !ok || len(quote) == 0 {
		return err("no quote data available")
}

	quoteData, found := quote[0].(map[string]interface{})
	if !found {
		return err("invalid quote data format")
}

	open, found := quoteData["open"].([]interface{})
	if !ok || len(open) == 0 {
		return err("no open price data available")
}

	close, found := quoteData["close"].([]interface{})
	if !ok || len(close) == 0 {
		return err("no close price data available")
}

	high, found := quoteData["high"].([]interface{})
	if !ok || len(high) == 0 {
		return err("no high price data available")
}

	low, found := quoteData["low"].([]interface{})
	if !ok || len(low) == 0 {
		return err("no low price data available")
}

	volume, found := quoteData["volume"].([]interface{})
	if !ok || len(volume) == 0 {
		return err("no volume data available")
}

	response := fmt.Sprintf("Symbol: %s\n", symbol)
	response += fmt.Sprintf("Period: %s\n", period)
	response += fmt.Sprintf("Interval: %s\n", interval)
	response += fmt.Sprintf("Open: %v\n", open[0])
	response += fmt.Sprintf("Close: %v\n", close[0])
	response += fmt.Sprintf("High: %v\n", high[0])
	response += fmt.Sprintf("Low: %v\n", low[0])
	response += fmt.Sprintf("Volume: %v\n", volume[0])

	return ok(response)
}

func HandleYahooFinanceSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query is required")
}

	apiURL := fmt.Sprintf("https://query1.finance.yahoo.com/v1/finance/search?q=%s", url.QueryEscape(query))

	client := http.Client{Timeout: 30 * time.Second}
	resp, fetchErr := client.Get(apiURL)
	if fetchErr != nil {
		return err(fetchErr.Error())
}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API request failed with status: %s", resp.Status))
}

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	var result map[string]interface{}
	parseErr := json.Unmarshal(body, &result)
	if parseErr != nil {
		return err(parseErr.Error())
}

	quotes, found := result["quotes"].([]interface{})
	if !found {
		return err("invalid API response format")
}

	if len(quotes) == 0 {
		return ok("No results found")
}

	var response strings.Builder
	response.WriteString("Search results for: " + query + "\n\n")

	for _, quote := range quotes {
		q, found := quote.(map[string]interface{})
		if !found {
			continue
		}

		symbol, found := q["symbol"].(string)
		if !found {
			continue
		}

		name, found := q["name"].(string)
		if !found {
			name = ""
		}

		exchange, found := q["exchange"].(string)
		if !found {
			exchange = ""
		}

		response.WriteString(fmt.Sprintf("Symbol: %s\n", symbol))
		response.WriteString(fmt.Sprintf("Name: %s\n", name))
		response.WriteString(fmt.Sprintf("Exchange: %s\n", exchange))
		response.WriteString("--------------------\n")

	return ok(response.String())
}
}