package tools

import (
    "context"
    "encoding/json"
    "os"
)

func HandleContributeGuidelines(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    content, readErr := os.ReadFile("CONTRIBUTING.md")
    if readErr != nil {
        return err("failed to read CONTRIBUTING.md: " + readErr.Error())
}

    return ok(string(content))
}

func HandleListTags(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    tags := []string{
        "官方实现 🎖️",
        "社区实现",
        "官方参考",
        "Python 开发 🐍",
        "TypeScript 开发 📇",
        "JavaScript 开发 📇",
        "Rust 开发 🦀",
        "Go 开发 🏎️",
        "本地运行 🏠",
        "云服务 ☁️",
        "云端/本地 🏠☁️",
        "跨平台 🍎🪟🐧",
    }
    data, marshalErr := json.Marshal(tags)
    if marshalErr != nil {
        return err("failed to marshal tags: " + marshalErr.Error())
}

    return ok(string(data))
}

func HandleHealth(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
    return ok("OK")
}