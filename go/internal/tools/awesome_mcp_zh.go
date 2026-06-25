package tools

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var (
	categories = []string{
		"🔍 搜索",
		"📚 知识库",
		"🎨 图像",
		"🎬 视频",
		"🎵 音频",
		"📝 写作",
		"💻 开发",
		"☁️ 云服务",
		"🌐 浏览器",
		"📊 数据",
		"🔒 安全",
		"📱 移动",
		"🎮 游戏",
		"💰 金融",
		"🏥 医疗",
		"📚 教育",
		"🌍 国际化",
		"🔧 工具",
	}

	mcpResources = map[string][]map[string]string{
		"🔍 搜索": {
			{"name": "MCP Search", "desc": "通用搜索工具", "tags": "官方实现, TypeScript开发, 本地运行"},
		},
		"📚 知识库": {
			{"name": "Knowledge Base MCP", "desc": "知识库管理工具", "tags": "社区实现, Python开发, 本地运行"},
		},
		"💻 开发": {
			{"name": "File System MCP", "desc": "文件系统操作", "tags": "官方参考, TypeScript开发, 本地运行"},
			{"name": "Git MCP", "desc": "Git版本控制", "tags": "社区实现, TypeScript开发, 本地运行"},
		},
	}

	contributionGuidelines = `# 贡献指南

## 一、如何提交

1. Fork 本仓库，在 README.md 对应分类下新增一行
2. 保持表格格式一致，按字母/拼音顺序就近插入
3. 提交 PR，标题简洁说明新增内容

## 二、收录标准

### ✅ 我们欢迎
- 真实可验证的 MCP Server / Client / 资源
- 可安装 / 可调用
- 文档清晰
- 对中文用户有价值

### ❌ 通常不予收录
- 并非真实的 MCP Server
- 付费墙套壳
- 竞品目录 / 聚合导航站
- 纯推广 / 营销
- 不稳定 / 难验证
- 重复条目

## 三、格式规范

| 名称 | 中文介绍 | 备注 |
| :--- | :--- | :--- |
| [项目名](链接) | 一句话中文介绍 | 实现类型, 开发语言, 运行方式 |

## 四、备注标签

| 标签 | 含义 |
| :--- | :--- |
| 官方实现 🎖️ | 由项目方/厂商官方维护 |
| 社区实现 | 第三方/社区维护 |
| Python 开发 🐍 | 主要开发语言 |
| TypeScript 开发 📇 | 主要开发语言 |
| Go 开发 🏎️ | 主要开发语言 |
| 本地运行 🏠 | 运行方式 |
| 云服务 ☁️ | 运行方式 |`
)

// HandleListCategories lists categories
func HandleListCategories(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	var builder strings.Builder
	builder.WriteString("# Awesome MCP ZH 分类列表\n\n")
	builder.WriteString("共收录 **")
	builder.WriteString(strconv.Itoa(len(categories)))
	builder.WriteString("** 个分类：\n\n")

	for i, cat := range categories {
		builder.WriteString(fmt.Sprintf("%d. %s\n", i+1, cat))
	}

	builder.WriteString("\n---\n\n")
	builder.WriteString("**GitHub**: https://github.com/yzfly/Awesome-MCP-ZH\n")
	builder.WriteString("**描述**: 面向中文用户的 MCP（模型上下文协议）资源列表\n")

	return ok(builder.String())
}

// HandleSearchResources searches resources
func HandleSearchResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	keyword, _ := getString(args, "keyword")
	if keyword == "" {
		return err("keyword 参数不能为空")
	}

	keyword = strings.ToLower(keyword)
	var results []map[string]interface{}
	var resultCount int

	for category, resources := range mcpResources {
		for _, res := range resources {
			name := strings.ToLower(res["name"])
			desc := strings.ToLower(res["desc"])
			tags := strings.ToLower(res["tags"])

			if strings.Contains(name, keyword) ||
				strings.Contains(desc, keyword) ||
				strings.Contains(tags, keyword) {

				results = append(results, map[string]interface{}{
					"category": category,
					"name":     res["name"],
					"desc":     res["desc"],
					"tags":     res["tags"],
				})
				resultCount++
			}
		}
	}

	var builder strings.Builder
	builder.WriteString("# 搜索结果\n\n")
	builder.WriteString(fmt.Sprintf("关键词: **%s**\n", keyword))
	builder.WriteString(fmt.Sprintf("找到 **%d** 个匹配结果\n\n", resultCount))

	if resultCount == 0 {
		builder.WriteString("未找到匹配的资源，请尝试其他关键词。\n")
	} else {
		builder.WriteString("---\n\n")
		for _, res := range results {
			builder.WriteString(fmt.Sprintf("### %s\n", res["name"]))
			builder.WriteString(fmt.Sprintf("**分类**: %s\n", res["category"]))
			builder.WriteString(fmt.Sprintf("**介绍**: %s\n", res["desc"]))
			builder.WriteString(fmt.Sprintf("**标签**: %s\n\n", res["tags"]))
		}
	}

	return ok(builder.String())
}

// HandleGetGuidelines returns guidelines
func HandleGetGuidelines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	section, _ := getString(args, "section")

	var builder strings.Builder
	builder.WriteString("# Awesome MCP ZH 贡献指南\n\n")
	builder.WriteString("**GitHub**: https://github.com/yzfly/Awesome-MCP-ZH\n\n")
	builder.WriteString("---\n\n")

	if section == "" {
		builder.WriteString(contributionGuidelines)
	} else {
		section = strings.ToLower(section)
		switch section {
		case "提交", "如何提交", "submit":
			builder.WriteString("## 一、如何提交\n\n")
			builder.WriteString("1. Fork 本仓库，在 README.md 对应分类下新增一行\n")
			builder.WriteString("2. 保持表格格式一致，按字母/拼音顺序就近插入\n")
			builder.WriteString("3. 提交 PR，标题简洁说明新增内容\n")
			builder.WriteString("4. **优先提交 PR**\n\n")
			builder.WriteString("> 一个 PR 尽量只做一件事\n")
		case "标准", "收录标准", "criteria":
			builder.WriteString("## 二、收录标准\n\n")
			builder.WriteString("### ✅ 我们欢迎\n")
			builder.WriteString("- 真实可验证的 MCP Server / Client / 资源\n")
			builder.WriteString("- 可安装 / 可调用\n")
			builder.WriteString("- 文档清晰\n")
			builder.WriteString("- 对中文用户有价值\n\n")
			builder.WriteString("### ❌ 通常不予收录\n")
			builder.WriteString("- 并非真实的 MCP Server\n")
			builder.WriteString("- 付费墙套壳\n")
			builder.WriteString("- 竞品目录 / 聚合导航站\n")
			builder.WriteString("- 纯推广 / 营销\n")
			builder.WriteString("- 不稳定 / 难验证\n")
			builder.WriteString("- 重复条目\n")
		case "格式", "格式规范", "format":
			builder.WriteString("## 三、格式规范\n\n")
			builder.WriteString("```markdown\n")
			builder.WriteString("| 名称 | 中文介绍 | 备注 |\n")
			builder.WriteString("| :--- | :--- | :--- |\n")
			builder.WriteString("| [项目名](链接) | 一句话中文介绍 | 实现类型, 开发语言, 运行方式 |\n")
			builder.WriteString("```\n\n")
			builder.WriteString("### 常用标签\n")
			builder.WriteString("| 标签 | 含义 |\n")
			builder.WriteString("| :--- | :--- |\n")
			builder.WriteString("| 官方实现 🎖️ | 由项目方/厂商官方维护 |\n")
			builder.WriteString("| 社区实现 | 第三方/社区维护 |\n")
			builder.WriteString("| Python 开发 🐍 | 主要开发语言 |\n")
			builder.WriteString("| TypeScript 开发 📇 | 主要开发语言 |\n")
			builder.WriteString("| 本地运行 🏠 | 运行方式 |\n")
			builder.WriteString("| 云服务 ☁️ | 运行方式 |\n")
		default:
			builder.WriteString(contributionGuidelines)
		}
	}

	return ok(builder.String())
}

// HandleValidateFormat validates format
func HandleValidateFormat(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ := getString(args, "name")
	desc, _ := getString(args, "description")
	tags, _ := getString(args, "tags")

	var builder strings.Builder
	builder.WriteString("# 格式验证结果\n\n")

	var issues []string
	var warnings []string

	if name == "" {
		issues = append(issues, "缺少名称字段")
	}
	if desc == "" {
		issues = append(issues, "缺少中文介绍字段")
	}
	if tags == "" {
		warnings = append(warnings, "缺少备注标签字段")
	}

	if strings.Contains(name, "[") && !strings.Contains(name, "]") {
		issues = append(issues, "名称格式错误：缺少闭合括号 ]")
	}
	if strings.Contains(name, "]") && !strings.Contains(name, "[") {
		issues = append(issues, "名称格式错误：缺少开放括号 [")
	}

	linkPattern := regexp.MustCompile(`\]\(https?://`)
	if strings.Contains(name, "](") && !linkPattern.MatchString(name) {
		warnings = append(warnings, "链接格式可能不正确，应为 [文本](URL) 格式")
	}

	requiredTags := []string{"官方实现", "社区实现", "官方参考"}
	hasRequiredTag := false
	for _, tag := range requiredTags {
		if strings.Contains(tags, tag) {
			hasRequiredTag = true
			break
		}
	}
	if !hasRequiredTag && tags != "" {
		warnings = append(warnings, "建议添加实现类型标签（官方实现/社区实现/官方参考）")
	}

	langTags := []string{"Python", "TypeScript", "Go", "Rust", "JavaScript"}
	hasLangTag := false
	for _, tag := range langTags {
		if strings.Contains(tags, tag) {
			hasLangTag = true
			break
		}
	}
	if !hasLangTag && tags != "" {
		warnings = append(warnings, "建议添加开发语言标签（Python/TypeScript/Go/Rust）")
	}

	builder.WriteString(fmt.Sprintf("**名称**: %s\n", name))
	builder.WriteString(fmt.Sprintf("**介绍**: %s\n", desc))
	builder.WriteString(fmt.Sprintf("**标签**: %s\n\n", tags))

	if len(issues) > 0 {
		builder.WriteString("## ❌ 问题\n\n")
		for _, issue := range issues {
			builder.WriteString(fmt.Sprintf("- %s\n", issue))
		}
		builder.WriteString("\n")
	}

	if len(warnings) > 0 {
		builder.WriteString("## ⚠️ 警告\n\n")
		for _, warning := range warnings {
			builder.WriteString(fmt.Sprintf("- %s\n", warning))
		}
		builder.WriteString("\n")
	}

	if len(issues) == 0 && len(warnings) == 0 {
		builder.WriteString("## ✅ 验证通过\n\n")
		builder.WriteString("格式符合 Awesome MCP ZH 的收录标准！\n")
	} else if len(issues) == 0 {
		builder.WriteString("## ✅ 无严重问题\n\n")
		builder.WriteString("请根据警告内容进行适当修改。\n")
	}

	return ok(builder.String())
}

// HandleGetResourceInfo gets resource info
func HandleGetResourceInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	resourceName, _ := getString(args, "name")
	if resourceName == "" {
		return err("name 参数不能为空")
	}

	resourceName = strings.ToLower(resourceName)

	for category, resources := range mcpResources {
		for _, res := range resources {
			if strings.ToLower(res["name"]) == resourceName {
				var builder strings.Builder
				builder.WriteString(fmt.Sprintf("# %s\n\n", res["name"]))
				builder.WriteString(fmt.Sprintf("**分类**: %s\n", category))
				builder.WriteString(fmt.Sprintf("**介绍**: %s\n", res["desc"]))
				builder.WriteString(fmt.Sprintf("**标签**: %s\n", res["tags"]))
				builder.WriteString("\n---\n\n")
				builder.WriteString("**GitHub**: https://github.com/yzfly/Awesome-MCP-ZH\n")

				return ok(builder.String())
			}
		}
	}

	return err(fmt.Sprintf("未找到资源: %s", resourceName))
}

// HandleListResources lists resources
func HandleListResources(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	category, _ := getString(args, "category")
	limit, _ := getInt(args, "limit")
	if limit == 0 {
		limit = 20
	}

	var builder strings.Builder
	builder.WriteString("# Awesome MCP ZH 资源列表\n\n")

	if category != "" {
		builder.WriteString(fmt.Sprintf("分类: **%s**\n\n", category))

		resources, exists := mcpResources[category]
		if !exists {
			return err(fmt.Sprintf("未找到分类: %s", category))
		}

		count := 0
		for _, res := range resources {
			if count >= limit {
				break
			}
			builder.WriteString(fmt.Sprintf("### %s\n", res["name"]))
			builder.WriteString(fmt.Sprintf("**介绍**: %s\n", res["desc"]))
			builder.WriteString(fmt.Sprintf("**标签**: %s\n\n", res["tags"]))
			count++
		}

		builder.WriteString(fmt.Sprintf("共 %d 个资源\n", len(resources)))
	} else {
		builder.WriteString("所有分类资源概览：\n\n")

		totalCount := 0
		for cat, resources := range mcpResources {
			builder.WriteString(fmt.Sprintf("## %s (%d)\n", cat, len(resources)))
			for _, res := range resources {
				if totalCount >= limit {
					break
				}
				builder.WriteString(fmt.Sprintf("- **%s**: %s\n", res["name"], res["desc"]))
				totalCount++
			}
			builder.WriteString("\n")
		}
	}

	builder.WriteString("---\n")
	builder.WriteString("**GitHub**: https://github.com/yzfly/Awesome-MCP-ZH\n")

	return ok(builder.String())
}