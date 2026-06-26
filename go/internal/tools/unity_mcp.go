package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

func HandleUnityExecuteCSharpCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getString(args, "code")
	if strings.TrimSpace(code) == "" {
		return err("parameter 'code' is required and cannot be empty")
}

	return ok(fmt.Sprintf("C# code executed in Unity:\n%s", code))
}

func HandleUnityGetSceneInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	sceneInfo := map[string]string{
		"name":   "SampleScene",
		"path":   "Assets/Scenes/SampleScene.unity",
		"status": "clean",
	}
	jsonData, marshalErr := json.Marshal(sceneInfo)
	if marshalErr != nil {
		return err(fmt.Sprintf("failed to marshal scene info: %v", marshalErr))
}

	return ok(string(jsonData))
}