package hook

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// EventType represents the type of hook event
type EventType string

const (
	PreToolUse    EventType = "PreToolUse"
	PostToolUse   EventType = "PostToolUse"
	Notification  EventType = "Notification"
	Stop          EventType = "Stop"
	SubagentStop  EventType = "SubagentStop"
)

// AllEventTypes returns all valid event types
func AllEventTypes() []EventType {
	return []EventType{PreToolUse, PostToolUse, Notification, Stop, SubagentStop}
}

// EventTypeNames returns all valid event type names as strings (for CLI completion)
func EventTypeNames() []string {
	types := AllEventTypes()
	names := make([]string, len(types))
	for i, t := range types {
		names[i] = string(t)
	}
	return names
}

// ParseEventType parses a string to EventType with alias support
// Accepts: full name (PreToolUse), lowercase (pretooluse), or alias (pre)
func ParseEventType(s string) (EventType, error) {
	// Aliases mapping
	aliases := map[string]EventType{
		// Full names (case-insensitive)
		"pretooluse":   PreToolUse,
		"posttooluse":  PostToolUse,
		"notification": Notification,
		"stop":         Stop,
		"subagentstop": SubagentStop,
		// Short aliases
		"pre":      PreToolUse,
		"post":     PostToolUse,
		"notify":   Notification,
		"notif":    Notification,
		"subagent": SubagentStop,
		"sub":      SubagentStop,
	}

	lower := strings.ToLower(s)
	if et, ok := aliases[lower]; ok {
		return et, nil
	}

	return "", fmt.Errorf("invalid event type: %s\nValid types: PreToolUse(pre), PostToolUse(post), Notification(notify), Stop, SubagentStop(sub)", s)
}

// HookCommand represents a single hook command
// Example: {"type": "command", "command": "echo Done"}
type HookCommand struct {
	Type    string `json:"type"`
	Command string `json:"command"`
}

// HookRule represents a single hook rule with matcher and commands
// matcher is a string pattern: "Bash", "Edit|Write", "*"
type HookRule struct {
	Matcher string        `json:"matcher"`
	Hooks   []HookCommand `json:"hooks"`
}

// Hook represents a named hook configuration for display/management
type Hook struct {
	Name      string    `json:"name"`
	EventType EventType `json:"event_type"`
	Matcher   string    `json:"matcher"`  // pattern: "Bash", "Edit|Write", "*"
	Commands  []string  `json:"commands"` // from hooks[].command
}

// Settings represents the Claude Code settings.json structure
type Settings struct {
	Hooks map[EventType][]HookRule `json:"hooks,omitempty"`
	// Other settings fields can be added here
	Other map[string]interface{} `json:"-"`
}

// Store manages hooks in settings.json
type Store struct {
	settingsPath string
}

// NewStore creates a new hook store
func NewStore(settingsPath string) *Store {
	return &Store{settingsPath: settingsPath}
}

// expandPath expands ~ to home directory
func (s *Store) expandPath() (string, error) {
	path := s.settingsPath
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[2:])
	}
	return path, nil
}

// readSettings reads and parses settings.json
func (s *Store) readSettings() (*Settings, map[string]interface{}, error) {
	path, err := s.expandPath()
	if err != nil {
		return nil, nil, err
	}

	content, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Settings{Hooks: make(map[EventType][]HookRule)}, make(map[string]interface{}), nil
		}
		return nil, nil, err
	}

	// Parse into generic map to preserve unknown fields
	var raw map[string]interface{}
	if err := json.Unmarshal(content, &raw); err != nil {
		return nil, nil, fmt.Errorf("failed to parse settings.json: %w", err)
	}

	// Parse hooks
	settings := &Settings{Hooks: make(map[EventType][]HookRule)}

	if hooksRaw, ok := raw["hooks"].(map[string]interface{}); ok {
		for eventType, rules := range hooksRaw {
			rulesArr, ok := rules.([]interface{})
			if !ok {
				continue
			}

			var hookRules []HookRule
			for _, r := range rulesArr {
				ruleMap, ok := r.(map[string]interface{})
				if !ok {
					continue
				}

				rule := HookRule{}

				// Parse matcher string: "Bash", "Edit|Write", "*"
				if matcher, ok := ruleMap["matcher"].(string); ok {
					rule.Matcher = matcher
				}

				// Parse hooks array: [{"type": "command", "command": "..."}]
				if hooksArr, ok := ruleMap["hooks"].([]interface{}); ok {
					for _, h := range hooksArr {
						if hookMap, ok := h.(map[string]interface{}); ok {
							hookCmd := HookCommand{}
							if t, ok := hookMap["type"].(string); ok {
								hookCmd.Type = t
							}
							if cmd, ok := hookMap["command"].(string); ok {
								hookCmd.Command = cmd
							}
							if hookCmd.Command != "" {
								rule.Hooks = append(rule.Hooks, hookCmd)
							}
						}
					}
				}

				hookRules = append(hookRules, rule)
			}
			settings.Hooks[EventType(eventType)] = hookRules
		}
	}

	return settings, raw, nil
}

// writeSettings writes settings back to settings.json
func (s *Store) writeSettings(settings *Settings, raw map[string]interface{}) error {
	path, err := s.expandPath()
	if err != nil {
		return err
	}

	// Convert hooks to JSON format
	hooksMap := make(map[string][]map[string]interface{})
	for eventType, rules := range settings.Hooks {
		if len(rules) == 0 {
			continue
		}

		var rulesOutput []map[string]interface{}
		for _, rule := range rules {
			ruleMap := map[string]interface{}{
				"matcher": rule.Matcher, // string: "Bash", "Edit|Write", "*"
				"hooks": func() []map[string]interface{} {
					var hooks []map[string]interface{}
					for _, h := range rule.Hooks {
						hooks = append(hooks, map[string]interface{}{
							"type":    h.Type,
							"command": h.Command,
						})
					}
					return hooks
				}(),
			}
			rulesOutput = append(rulesOutput, ruleMap)
		}
		hooksMap[string(eventType)] = rulesOutput
	}

	if len(hooksMap) > 0 {
		raw["hooks"] = hooksMap
	} else {
		delete(raw, "hooks")
	}

	content, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return err
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, content, 0644)
}

// List returns all hooks as a flat list
func (s *Store) List() ([]*Hook, error) {
	settings, _, err := s.readSettings()
	if err != nil {
		return nil, err
	}

	var hooks []*Hook
	for eventType, rules := range settings.Hooks {
		for i, rule := range rules {
			name := generateHookName(eventType, rule.Matcher, i)

			var commands []string
			for _, h := range rule.Hooks {
				commands = append(commands, h.Command)
			}

			hooks = append(hooks, &Hook{
				Name:      name,
				EventType: eventType,
				Matcher:   rule.Matcher,
				Commands:  commands,
			})
		}
	}

	return hooks, nil
}

// Get retrieves a specific hook by name
func (s *Store) Get(name string) (*Hook, error) {
	hooks, err := s.List()
	if err != nil {
		return nil, err
	}

	for _, h := range hooks {
		if h.Name == name {
			return h, nil
		}
	}

	return nil, os.ErrNotExist
}

// Add adds a new hook rule
func (s *Store) Add(eventType EventType, matcher string, commands []string) (*Hook, error) {
	settings, raw, err := s.readSettings()
	if err != nil {
		return nil, err
	}

	// Build hook commands
	var hookCmds []HookCommand
	for _, cmd := range commands {
		hookCmds = append(hookCmds, HookCommand{
			Type:    "command",
			Command: cmd,
		})
	}

	rule := HookRule{
		Matcher: matcher,
		Hooks:   hookCmds,
	}

	settings.Hooks[eventType] = append(settings.Hooks[eventType], rule)

	if err := s.writeSettings(settings, raw); err != nil {
		return nil, err
	}

	idx := len(settings.Hooks[eventType]) - 1
	return &Hook{
		Name:      generateHookName(eventType, matcher, idx),
		EventType: eventType,
		Matcher:   matcher,
		Commands:  commands,
	}, nil
}

// Update updates an existing hook
func (s *Store) Update(name string, matcher string, commands []string) (*Hook, error) {
	settings, raw, err := s.readSettings()
	if err != nil {
		return nil, err
	}

	eventType, idx, err := parseHookName(name)
	if err != nil {
		return nil, err
	}

	rules, ok := settings.Hooks[eventType]
	if !ok || idx >= len(rules) {
		return nil, os.ErrNotExist
	}

	// Build hook commands
	var hookCmds []HookCommand
	for _, cmd := range commands {
		hookCmds = append(hookCmds, HookCommand{
			Type:    "command",
			Command: cmd,
		})
	}

	rules[idx].Matcher = matcher
	rules[idx].Hooks = hookCmds
	settings.Hooks[eventType] = rules

	if err := s.writeSettings(settings, raw); err != nil {
		return nil, err
	}

	return &Hook{
		Name:      generateHookName(eventType, matcher, idx),
		EventType: eventType,
		Matcher:   matcher,
		Commands:  commands,
	}, nil
}

// Delete removes a hook by name
func (s *Store) Delete(name string) error {
	settings, raw, err := s.readSettings()
	if err != nil {
		return err
	}

	eventType, idx, err := parseHookName(name)
	if err != nil {
		return err
	}

	rules, ok := settings.Hooks[eventType]
	if !ok || idx >= len(rules) {
		return os.ErrNotExist
	}

	// Remove the rule at index
	settings.Hooks[eventType] = append(rules[:idx], rules[idx+1:]...)

	return s.writeSettings(settings, raw)
}

// generateHookName creates a unique name for a hook
func generateHookName(eventType EventType, matcher string, index int) string {
	// Sanitize matcher for use in name
	sanitized := matcher
	if sanitized == "" || sanitized == "*" {
		sanitized = "all"
	}
	sanitized = strings.ReplaceAll(sanitized, "|", "-")
	sanitized = strings.ReplaceAll(sanitized, " ", "_")

	return fmt.Sprintf("%s-%s-%d", eventType, sanitized, index)
}

// parseHookName extracts event type and index from hook name
func parseHookName(name string) (EventType, int, error) {
	for _, et := range AllEventTypes() {
		prefix := string(et) + "-"
		if strings.HasPrefix(name, prefix) {
			rest := strings.TrimPrefix(name, prefix)
			// Find last dash to get index
			lastDash := strings.LastIndex(rest, "-")
			if lastDash == -1 {
				continue
			}
			var idx int
			if _, err := fmt.Sscanf(rest[lastDash+1:], "%d", &idx); err == nil {
				return et, idx, nil
			}
		}
	}
	return "", 0, fmt.Errorf("invalid hook name: %s", name)
}

// GetHooksDir returns the hooks script directory path
func GetHooksDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude", "hooks"), nil
}

// EnsureHooksDir creates the hooks directory if it doesn't exist
func EnsureHooksDir() (string, error) {
	dir, err := GetHooksDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}

// CreateScript creates a hook script file
func CreateScript(name, content string) (string, error) {
	dir, err := EnsureHooksDir()
	if err != nil {
		return "", err
	}

	scriptPath := filepath.Join(dir, name)
	if err := os.WriteFile(scriptPath, []byte(content), 0755); err != nil {
		return "", err
	}

	return scriptPath, nil
}

// ListScripts returns all script files in the hooks directory
func ListScripts() ([]string, error) {
	dir, err := GetHooksDir()
	if err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var scripts []string
	for _, entry := range entries {
		if !entry.IsDir() {
			scripts = append(scripts, entry.Name())
		}
	}

	return scripts, nil
}
