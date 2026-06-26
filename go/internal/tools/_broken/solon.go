package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"net/http"
)

// solonInfo holds the static information about the Solon framework
var solonInfo = map[string]interface{}{
	"name":        "Solon",
	"description": "Java enterprise application development framework for full scenario",
	"features": []string{
		"700% higher concurrency",
		"50% memory savings",
		"10x faster startup",
		"90% smaller packaging",
		"Supports Java 8 to Java 25",
		"Native runtime support (GraalVM)",
	},
	"website": "https://solon.noear.org",
	"license": "Apache 2.0",
}

// solonRepos holds the list of main code repositories
var solonRepos = []map[string]string{
	{"name": "solon", "url": "https://gitee.com/opensolon/solon", "desc": "Main code repository"},
	{"name": "solon-examples", "url": "https://gitee.com/opensolon/solon-examples", "desc": "Official website supporting sample code"},
	{"name": "solon-ai", "url": "https://gitee.com/opensolon/solon-ai", "desc": "Solon Ai code repository"},
	{"name": "solon-cloud", "url": "https://gitee.com/opensolon/solon-cloud", "desc": "Solon Cloud code repository"},
	{"name": "solon-admin", "url": "https://gitee.com/opensolon/solon-admin", "desc": "Solon Admin code repository"},
	{"name": "solon-maven-plugin", "url": "https://gitee.com/opensolon/solon-maven-plugin", "desc": "Solon Maven plugin code repository"},
}

// HandleSolonInfo returns general information about the Solon framework
func HandleSolonInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	data, e := json.MarshalIndent(solonInfo, "", "  ")
	if e != nil {
		return err(fmt.Sprintf("failed to marshal info: %v", e))
}

	return ok(string(data))
}

// HandleSolonRepos returns a list of main Solon code repositories
func HandleSolonRepos(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ :=getString(args, "category")

	var filteredRepos []map[string]string
	if category == "" {
		filteredRepos = solonRepos
	} else {
		for _, repo := range solonRepos {
			if strings.Contains(repo["name"], category) {
				filteredRepos = append(filteredRepos, repo)

		}
	}

	data, e := json.MarshalIndent(filteredRepos, "", "  ")
	if e != nil {
		return err(fmt.Sprintf("failed to marshal repos: %v", e))
}

	return ok(string(data))
}

}

// HandleSolonSearch searches for information in Solon documentation or repositories
func HandleSolonSearch(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	query, _ :=getString(args, "query")
	if query == "" {
		return err("query parameter is required")
}

	// Simulate search by checking against known keywords
	knownKeywords := []string{"feature", "performance", "startup", "memory", "concurrency", "java", "cloud", "ai", "plugin"}

	var matches []string
	for _, keyword := range knownKeywords {
		if strings.Contains(strings.ToLower(query), keyword) {
			matches = append(matches, keyword)

	}

	if len(matches) == 0 {
		return ok(fmt.Sprintf("No specific matches found for '%s'. Try keywords like: %s", query, strings.Join(knownKeywords, ", ")))
}

	sort.Strings(matches)
	result := fmt.Sprintf("Found matches for '%s': %s", query, strings.Join(matches, ", "))
	return ok(result)
}

}

// HandleSolonContribute returns contribution guidelines for Solon
func HandleSolonContribute(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	guidelines := `
# Solon Contribution Guidelines

## 1. Copyright
Source code copyright belongs to noear open source organization.

## 2. Contribution Categories
- Code Contribution: Fix issues, optimize code, add new plugins, add unit tests.
- Cooperation Contribution: Add Solon adaptation to your own open source projects.
- Other Contributions: Submit issues, write blogs, record videos, recommend Solon in communities.

## 3. Code Contribution Steps
1. Submit an Issue and confirm with administrators.
2. Fork the repository.
3. Write code on the main branch and add unit tests.
4. Use 'solon-test' for batch testing.
5. PR to the main branch (link an Issue).
6. For distributed middleware, adapt to solon cloud specifications first.
7. Add more comments.

## 4. Branch Protection
- main branch: No direct pushes allowed. Only admins can merge PRs.

## 5. Test Directory Structure
- src/test/benchmark: Performance tests (optional)
- src/test/demo: Simple examples (required)
- src/test/features: Feature tests (required, included in batch tests)
- src/test/labs: Experimental tests (optional, not in batch tests)

## 6. Commit Message Prefixes
- 新增 (New): Add new module
- 添加 (Add): Add new capability to a module
- 优化 (Optimize): Optimize existing code
- 修复 (Fix): Fix existing issues
- 调整 (Adjust): Adjust existing code (may have compatibility risks)
- 移除 (Remove): Remove redundant classes
- 文档 (Doc): Improve documentation
- 测试 (Test): Improve tests
- 其它 (Other): Other changes
`
	return ok(guidelines)
}

// HandleSolonCheckHealth checks the availability of Solon's main website
func HandleSolonCheckHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	urlStr := "https://solon.noear.org"

	client := http.DefaultClient
	req, e := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if e != nil {
		return err(fmt.Sprintf("failed to create request: %v", e))
}

	resp, fetchErr := client.Do(req)
	if fetchErr != nil {
		return err(fmt.Sprintf("failed to fetch website: %v", fetchErr))
}

	defer resp.Body.Close()

	status := "healthy"
	if resp.StatusCode != 200 {
		status = fmt.Sprintf("unhealthy (status code: %d)", resp.StatusCode)

	result := fmt.Sprintf("Solon website (%s) is %s", urlStr, status)
	return ok(result)
}

}

// HandleSolonRegexPattern validates if a string matches Solon's commit message pattern
func HandleSolonRegexPattern(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	input, _ :=getString(args, "input")
	if input == "" {
		return err("input parameter is required")
}

	// Pattern based on the commit message prefixes described in CONTRIBUTING.md
	// Matches: 新增, 添加, 优化, 修复, 调整, 移除, 文档, 测试, 其它 followed by space and text
	pattern := `^(新增|添加|优化|修复|调整|移除|文档|测试|其它)\s+.+$`

	re := regexp.MustCompile(pattern)
	isMatch := re.MatchString(input)

	result := fmt.Sprintf("Input: '%s'\nMatches Solon commit pattern: %v", input, isMatch)
	if !isMatch {
		result += "\nExpected format: 'Prefix Text' (e.g., '新增 solon-xxx 模块')"
	}

	return ok(result)
}