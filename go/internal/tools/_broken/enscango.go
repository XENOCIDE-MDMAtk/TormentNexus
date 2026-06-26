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

var (
	enscanAPIURL = "http://localhost:8080"
	enscanClient = http.DefaultClient
)

// HandleSearchCompany searches for companies by name
func HandleSearchCompany(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	sourceType, _ :=getString(args, "type")
	if sourceType == "" {
		sourceType = "aqc"
	}

	investStr, _ :=getString(args, "invest")
	var invest int
	if investStr != "" {
		var parseErr error
		invest, parseErr = strconv.Atoi(investStr)
		if parseErr != nil {
			return err("invalid invest value")

	}

	branch, _ :=getBool(args, "branch")
	deepStr, _ :=getString(args, "deep")
	var deep int
	if deepStr != "" {
		var parseErr error
		deep, parseErr = strconv.Atoi(deepStr)
		if parseErr != nil {
			return err("invalid deep value")

	}

	params := url.Values{}
	params.Set("name", name)
	params.Set("type", sourceType)
	if invest > 0 {
		params.Set("invest", strconv.Itoa(invest))

	if branch {
		params.Set("branch", "true")

	if deep > 0 {
		params.Set("depth", strconv.Itoa(deep))

	apiURL := fmt.Sprintf("%s/api/info?%s", enscanAPIURL, params.Encode())
	
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := enscanClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to enscan server: %v (make sure enscan is running with --mcp flag)", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
		return ok(string(body))
}

	output, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(output))
}

}
}
}
}
}

// HandleGetCompanyInfo retrieves company information by PID
func HandleGetCompanyInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	pid, _ :=getString(args, "pid")
	if pid == "" {
		return err("pid is required")
}

	sourceType, _ :=getString(args, "type")
	if sourceType == "" {
		sourceType = "aqc"
	}

	params := url.Values{}
	params.Set("pid", pid)
	params.Set("type", sourceType)

	apiURL := fmt.Sprintf("%s/api/pro/get_base_info?%s", enscanAPIURL, params.Encode())
	
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := enscanClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to enscan server: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

// HandleGetCompanyFields retrieves specific fields for a company
func HandleGetCompanyFields(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	field, _ :=getString(args, "field")
	if field == "" {
		field = "icp,app,wechat,weibo"
	}

	sourceType, _ :=getString(args, "type")
	if sourceType == "" {
		sourceType = "aqc"
	}

	params := url.Values{}
	params.Set("name", name)
	params.Set("type", sourceType)
	params.Set("field", field)

	apiURL := fmt.Sprintf("%s/api/info?%s", enscanAPIURL, params.Encode())
	
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := enscanClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to enscan server: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
		return ok(string(body))
}

	output, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(output))
}

// HandleDeepSearch performs recursive company relationship search
func HandleDeepSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	investStr, _ :=getString(args, "invest")
	invest := 51
	if investStr != "" {
		var parseErr error
		invest, parseErr = strconv.Atoi(investStr)
		if parseErr != nil {
			return err("invalid invest value")

	}

	deepStr, _ :=getString(args, "deep")
	deep := 1
	if deepStr != "" {
		var parseErr error
		deep, parseErr = strconv.Atoi(deepStr)
		if parseErr != nil {
			return err("invalid deep value")

	}

	sourceType, _ :=getString(args, "type")
	if sourceType == "" {
		sourceType = "aqc"
	}

	fields, _ :=getString(args, "field")
	if fields == "" {
		fields = "icp"
	}

	params := url.Values{}
	params.Set("name", name)
	params.Set("type", sourceType)
	params.Set("invest", strconv.Itoa(invest))
	params.Set("depth", strconv.Itoa(deep))
	params.Set("field", fields)

	apiURL := fmt.Sprintf("%s/api/info?%s", enscanAPIURL, params.Encode())
	
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := enscanClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to enscan server: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
		return ok(string(body))
}

	output, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(output))
}

}
}

// HandleListDataSources returns available data sources
func HandleListDataSources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sources := map[string]interface{}{
		"data_sources": []map[string]string{
			{"code": "aqc", "name": "爱企查", "description": "百度旗下企业信息查询"},
			{"code": "tyc", "name": "天眼查", "description": "天眼查企业信息查询"},
			{"code": "kc", "name": "快查", "description": "快查企业信息查询"},
			{"code": "rb", "name": "风鸟", "description": "风鸟企业信息查询"},
		},
		"plugins": []map[string]string{
			{"code": "miit", "name": "工信部ICP备案", "description": "工业和信息化部ICP备案查询"},
			{"code": "coolapk", "name": "酷安市场", "description": "酷安应用市场APP查询"},
			{"code": "qimai", "name": "七麦数据", "description": "七麦数据APP查询"},
		},
		"fields": []map[string]string{
			{"code": "icp", "name": "网站备案", "description": "ICP备案信息"},
			{"code": "weibo", "name": "微博", "description": "官方微博账号"},
			{"code": "wechat", "name": "微信公众号", "description": "微信公众号"},
			{"code": "app", "name": "APP", "description": "移动应用信息"},
			{"code": "job", "name": "招聘", "description": "公开招聘信息"},
			{"code": "wx_app", "name": "小程序", "description": "微信小程序"},
			{"code": "copyright", "name": "著作权", "description": "软件著作权"},
			{"code": "supplier", "name": "供应商", "description": "供应商信息"},
		},
	}

	output, marshalErr := json.MarshalIndent(sources, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(output))
}

// HandleGetFieldMappings returns field mappings for a data source
func HandleGetFieldMappings(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sourceType, _ :=getString(args, "type")
	if sourceType == "" {
		sourceType = "aqc"
	}

	validTypes := map[string]bool{"aqc": true, "tyc": true, "kc": true, "rb": true}
	if !validTypes[sourceType] {
		return err(fmt.Sprintf("invalid type: %s", sourceType))
}

	params := url.Values{}
	params.Set("type", sourceType)

	apiURL := fmt.Sprintf("%s/api/pro/get_ensd?%s", enscanAPIURL, params.Encode())
	
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := enscanClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to enscan server: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	return ok(string(body))
}

// HandleAdvanceFilter performs advanced company filtering
func HandleAdvanceFilter(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")
	if name == "" {
		return err("name is required")
}

	sourceType, _ :=getString(args, "type")
	if sourceType == "" {
		sourceType = "aqc"
	}

	params := url.Values{}
	params.Set("name", name)
	params.Set("type", sourceType)

	// Optional filters
	investStr, _ :=getString(args, "invest")
	if investStr != "" {
		params.Set("invest", investStr)

	holds, _ :=getBool(args, "holds")
	if holds {
		params.Set("holds", "true")

	supplier, _ :=getBool(args, "supplier")
	if supplier {
		params.Set("supplier", "true")

	branch, _ :=getBool(args, "branch")
	if branch {
		params.Set("branch", "true")

	apiURL := fmt.Sprintf("%s/api/pro/advance_filter?%s", enscanAPIURL, params.Encode())
	
	req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if reqErr != nil {
		return err(reqErr.Error())
}

	resp, fetchErr := enscanClient.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to connect to enscan server: %v", fetchErr))
}

	defer resp.Body.Close()

	body, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		return err(readErr.Error())
}

	if resp.StatusCode != http.StatusOK {
		return err(fmt.Sprintf("API returned status %d: %s", resp.StatusCode, string(body)))
}

	var result map[string]interface{}
	if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
		return ok(string(body))
}

	output, marshalErr := json.MarshalIndent(result, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(output))
}

}
}
}
}

// HandleBatchSearch searches multiple companies from a list
func HandleBatchSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	names, _ :=getString(args, "names")
	if names == "" {
		return err("names is required (comma-separated company names)")
}

	sourceType, _ :=getString(args, "type")
	if sourceType == "" {
		sourceType = "aqc"
	}

	field, _ :=getString(args, "field")
	if field == "" {
		field = "icp"
	}

	companyList := strings.Split(names, ",")
	results := make([]map[string]interface{}, 0, len(companyList))

	for _, name := range companyList {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}

		params := url.Values{}
		params.Set("name", name)
		params.Set("type", sourceType)
		params.Set("field", field)

		apiURL := fmt.Sprintf("%s/api/info?%s", enscanAPIURL, params.Encode())
		
		req, reqErr := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
		if reqErr != nil {
			continue
		}

		resp, fetchErr := enscanClient.Do(req)
		if fetchErr != nil {
			results = append(results, map[string]interface{}{
				"name":   name,
				"error":  fetchErr.Error(),
				"status": "failed",
			})
			continue
		}
		defer resp.Body.Close()

		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			results = append(results, map[string]interface{}{
				"name":   name,
				"error":  readErr.Error(),
				"status": "failed",
			})
			continue
		}

		var result map[string]interface{}
		if jsonErr := json.Unmarshal(body, &result); jsonErr != nil {
			results = append(results, map[string]interface{}{
				"name":    name,
				"data":    string(body),
				"status":  "success",
			})
		} else {
			results = append(results, map[string]interface{}{
				"name":   name,
				"data":   result,
				"status": "success",
			})

		// Rate limiting between requests
		time.Sleep(500 * time.Millisecond)

	output, marshalErr := json.MarshalIndent(results, "", "  ")
	if marshalErr != nil {
		return err(marshalErr.Error())
}

	return ok(string(output))
}
}
}