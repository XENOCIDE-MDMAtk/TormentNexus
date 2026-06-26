package tools

import (
	"context"
	"fmt"
	"os/ex")

[Switched to Model: z-ai/glm-5.1 | Provider: nvidia via Global Random Retry]

package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func runAdb(ctx context.Context, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "adb", args...)
	output, e := cmd.CombinedOutput()
	if e != nil {
		return "", fmt.Errorf("adb failed: %w, output: %s", e, string(output))
}

	return strings.TrimSpace(string(output)), nil
}

func HandleClick(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x, _ :=getInt(args, "pointX")
	y, _ :=getInt(args, "pointY")

	_, e := runAdb(ctx, "shell", "input", "tap", strconv.Itoa(x), strconv.Itoa(y))
	if e != nil {
		return err(e.Error())
}

	return ok("true")
}

func HandleSwipe(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	x1, _ :=getInt(args, "fromX")
	y1, _ :=getInt(args, "fromY")
	x2, _ :=getInt(args, "toX")
	y2, _ :=getInt(args, "toY")
	step, _ :=getInt(args, "step")
	if step == 0 {
		step = 300
	}

	_, e := runAdb(ctx, "shell", "input", "swipe", strconv.Itoa(x1), strconv.Itoa(y1), strconv.Itoa(x2), strconv.Itoa(y2), strconv.Itoa(step))
	if e != nil {
		return err(e.Error())
}

	return ok("true")
}

func HandleDumpWindowHierarchy(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	_ = getBool(args, "compressed")

	_, dumpErr := runAdb(ctx, "shell", "uiautomator", "dump", "/sdcard/window_dump.xml")
	if dumpErr != nil {
		return err(dumpErr.Error())
}

	xml, catErr := runAdb(ctx, "shell", "cat", "/sdcard/window_dump.xml")
	if catErr != nil {
		return err(catErr.Error())
}

	return ok(xml)
}

func HandleGetProp(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	name, _ :=getString(args, "name")

	val, e := runAdb(ctx, "shell", "getprop", name)
	if e != nil {
		return err(e.Error())
}

	return ok(val)
}

func HandlePressKeyCode(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	code, _ :=getInt(args, "key_code")

	_, e := runAdb(ctx, "shell", "input", "keyevent", strconv.Itoa(code))
	if e != nil {
		return err(e.Error())
}

	return ok("true")
}

func HandleSetText(ctx context.Context, args map[string]interface{}) (ToolResponse, error) {
	text, _ :=getString(args, "text")

	_, e := runAdb(ctx, "shell", "input", "text", text)
	if e != nil {
		return err(e.Error())
}

	return ok("true")
}