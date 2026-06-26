func getFloat(args map[string]interface{}, key string, defaultVal float64) float64 {
	if val, found := args[key]; found {
		switch v := val.(type) {
		case float64:
			return v
}
		case float32:
			return float64(v)
}
		case int:
			return float64(v)
}
		case int64:
			return float64(v)
		case int32:
			return float64(v)
		default:
			return defaultVal
		}
	}
	return defaultVal
}
if data, e := os.ReadFile(settingsPath); e == nil {
    _ = json.Unmarshal(data, &settings)
} else {
    settings = make(map[string]interface{})

if val, found := args["model"]; found {
    settings["model"] = val
}
if e := json.Unmarshal(data, &config); e == nil {
    result["config"] = config
}
if data, e := os.ReadFile(configPath); e == nil {
    var config map[string]interface{}
    if e := json.Unmarshal(data, &config); e == nil {
        result["config"] = config
    }
}
return ok(fmt.Sprintf("Osaurus chat request prepared for model '%s'. Message: %s (max_tokens: %d, temperature: %.2f)", model, message, maxTokens, temperature))
cmd := exec.Command("which", "osaurus")
if e := cmd.Run(); e != nil {
    return err("Osaurus CLI not found in PATH")

if e := json.Unmarshal(data, &config); e == nil {
    result["config"] = config
}
// Try to read model config if exists
configPath := filepath.Join(modelPath, "config.json")
if data, e := os.ReadFile(configPath); e == nil {
    var config map[string]interface{}
    if e := json.Unmarshal(data, &config); e == nil {
        result["config"] = config
    }
}