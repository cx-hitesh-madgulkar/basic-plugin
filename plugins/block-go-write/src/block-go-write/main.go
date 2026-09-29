// Command block-go-write is a PreToolUse hook that denies any file
// create/edit/write tool call targeting a .go source file.
package main

import (
"encoding/json"
"io"
"os"
"strings"
)

// preToolUseInput mirrors the JSON payload VS Code sends to a PreToolUse hook.
type preToolUseInput struct {
ToolName  string                 `json:"tool_name"`
ToolInput map[string]interface{} `json:"tool_input"`
}

// hookOutput mirrors the JSON payload VS Code expects back from a PreToolUse hook.
type hookOutput struct {
HookSpecificOutput hookSpecificOutput `json:"hookSpecificOutput"`
}

type hookSpecificOutput struct {
HookEventName            string `json:"hookEventName"`
PermissionDecision       string `json:"permissionDecision"`
PermissionDecisionReason string `json:"permissionDecisionReason"`
}

// writeToolNames are the known local tool names that create, edit, or write files.
var writeToolNames = map[string]bool{
"create_file":                  true,
"replace_string_in_file":       true,
"multi_replace_string_in_file": true,
"edit_notebook_file":           true,
"apply_patch":                  true,
"insert_edit_into_file":        true,
}

// filePathKeys are the tool_input keys that commonly hold a target file path.
var filePathKeys = []string{"filePath", "path", "file_path", "notebookPath"}

func main() {
raw, err := io.ReadAll(os.Stdin)
if err != nil {
// Can't read input: fail open, don't block the agent on a hook error.
os.Exit(0)
}

var input preToolUseInput
if err := json.Unmarshal(raw, &input); err != nil {
os.Exit(0)
}

if !isWriteTool(input.ToolName) {
os.Exit(0)
}

if !targetsGoFile(input.ToolInput) {
os.Exit(0)
}

out := hookOutput{
HookSpecificOutput: hookSpecificOutput{
HookEventName:            "PreToolUse",
PermissionDecision:       "deny",
PermissionDecisionReason: "Agent (u the copilot) will not write golang code",
},
}
json.NewEncoder(os.Stdout).Encode(out)
os.Exit(0)
}

// isWriteTool reports whether toolName is a known file create/edit/write tool.
func isWriteTool(toolName string) bool {
if writeToolNames[toolName] {
return true
}
lower := strings.ToLower(toolName)
for _, kw := range []string{"create", "edit", "write", "replace", "patch"} {
if strings.Contains(lower, kw) {
return true
}
}
return false
}

// targetsGoFile reports whether any file path referenced in toolInput ends in .go.
func targetsGoFile(toolInput map[string]interface{}) bool {
for _, key := range filePathKeys {
if v, ok := toolInput[key]; ok {
if isGoPath(v) {
return true
}
}
}
// multi_replace_string_in_file nests paths inside a "replacements" array.
if replacements, ok := toolInput["replacements"].([]interface{}); ok {
for _, r := range replacements {
entry, ok := r.(map[string]interface{})
if !ok {
continue
}
for _, key := range filePathKeys {
if v, ok := entry[key]; ok && isGoPath(v) {
return true
}
}
}
}
return false
}

// isGoPath reports whether v is a string ending in the .go extension.
func isGoPath(v interface{}) bool {
s, ok := v.(string)
if !ok {
return false
}
return strings.HasSuffix(strings.ToLower(s), ".go")
}
