package tools

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func HandleReadFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	content, readErr := os.ReadFile(path)
	if readErr != nil {
		return err(readErr.Error())
}

	return ok(string(content))
}

func HandleWriteFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	content, _ :=getString(args, "content")

	if path == "" {
		return err("path is required")
}

	dir := filepath.Dir(path)
	if dir != "" && dir != "." {
		if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
			return err(mkErr.Error())

	}

	writeErr := os.WriteFile(path, []byte(content), 0644)
	if writeErr != nil {
		return err(writeErr.Error())
}

	return ok("File written successfully: " + path)
}

}

func HandleListDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		path = "."
	}

	entries, readErr := os.ReadDir(path)
	if readErr != nil {
		return err(readErr.Error())
}

	var result []map[string]interface{}
	for _, entry := range entries {
		info, infoErr := entry.Info()
		item := map[string]interface{}{
			"name":  entry.Name(),
			"is_dir": entry.IsDir(),
		}
		if infoErr == nil {
			item["size"] = info.Size()
			item["mod_time"] = info.ModTime().Format("2006-01-02 15:04:05")

		result = append(result, item)

	jsonBytes, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}

}
}

func HandleCreateDirectory(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	mkErr := os.MkdirAll(path, 0755)
	if mkErr != nil {
		return err(mkErr.Error())
}

	return ok("Directory created successfully: " + path)
}

func HandleDeletePath(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	info, statErr := os.Stat(path)
	if statErr != nil {
		return err(statErr.Error())
}

	var delErr error
	if info.IsDir() {
		delErr = os.RemoveAll(path)
	} else {
		delErr = os.Remove(path)

	if delErr != nil {
		return err(delErr.Error())
}

	return ok("Deleted successfully: " + path)
}

}

func HandleMovePath(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	dest, _ :=getString(args, "destination")

	if source == "" || dest == "" {
		return err("source and destination are required")
}

	moveErr := os.Rename(source, dest)
	if moveErr != nil {
		return err(moveErr.Error())
}

	return ok("Moved successfully from " + source + " to " + dest)
}

func HandleCopyFile(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	source, _ :=getString(args, "source")
	dest, _ :=getString(args, "destination")

	if source == "" || dest == "" {
		return err("source and destination are required")
}

	srcFile, openErr := os.Open(source)
	if openErr != nil {
		return err(openErr.Error())
}

	defer srcFile.Close()

	dir := filepath.Dir(dest)
	if dir != "" && dir != "." {
		if mkErr := os.MkdirAll(dir, 0755); mkErr != nil {
			return err(mkErr.Error())

	}

	dstFile, createErr := os.Create(dest)
	if createErr != nil {
		return err(createErr.Error())
}

	defer dstFile.Close()

	_, copyErr := io.Copy(dstFile, srcFile)
	if copyErr != nil {
		return err(copyErr.Error())
}

	srcInfo, statErr := os.Stat(source)
	if statErr == nil {
		os.Chmod(dest, srcInfo.Mode())

	return ok("Copied successfully from " + source + " to " + dest)
}

}
}

func HandlePathExists(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	_, statErr := os.Stat(path)
	if statErr != nil {
		if os.IsNotExist(statErr) {
			return ok("false")
		}
		return err(statErr.Error())
}

	return ok("true")
}

func HandleGetFileInfo(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	path, _ :=getString(args, "path")
	if path == "" {
		return err("path is required")
}

	info, statErr := os.Stat(path)
	if statErr != nil {
		return err(statErr.Error())
}

	result := map[string]interface{}{
		"name":    filepath.Base(path),
		"size":    info.Size(),
		"is_dir":  info.IsDir(),
		"mode":    info.Mode().String(),
		"mod_time": info.ModTime().Format("2006-01-02 15:04:05"),
	}

	if !info.IsDir() {
		ext := filepath.Ext(path)
		result["extension"] = ext
		result["name_without_ext"] = strings.TrimSuffix(filepath.Base(path), ext)

	jsonBytes, jsonErr := json.Marshal(result)
	if jsonErr != nil {
		return err(jsonErr.Error())
}

	return ok(string(jsonBytes))
}
}