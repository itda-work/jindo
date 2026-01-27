package config

import (
	"errors"
	"strconv"
	"strings"
)

var (
	// ErrKeyNotFound is returned when a key doesn't exist
	ErrKeyNotFound = errors.New("key not found")
	// ErrInvalidKey is returned for invalid key format
	ErrInvalidKey = errors.New("invalid key format")
	// ErrNotAMap is returned when trying to traverse into a non-map value
	ErrNotAMap = errors.New("intermediate key is not a map")
)

// parseDotKey splits a dot notation key into parts
// "common.api_keys.tiingo" -> ["common", "api_keys", "tiingo"]
func parseDotKey(key string) ([]string, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, ErrInvalidKey
	}
	parts := strings.Split(key, ".")
	for _, part := range parts {
		if part == "" {
			return nil, ErrInvalidKey
		}
	}
	return parts, nil
}

// getNestedValue retrieves a value from nested maps using key parts
func getNestedValue(data map[string]any, keys []string) (any, error) {
	if len(keys) == 0 {
		return data, nil
	}

	current := any(data)
	for i, key := range keys {
		m, ok := current.(map[string]any)
		if !ok {
			return nil, ErrNotAMap
		}
		val, exists := m[key]
		if !exists {
			return nil, ErrKeyNotFound
		}
		if i == len(keys)-1 {
			return val, nil
		}
		current = val
	}
	return current, nil
}

// setNestedValue sets a value in nested maps, creating intermediates as needed
func setNestedValue(data map[string]any, keys []string, value any) error {
	if len(keys) == 0 {
		return ErrInvalidKey
	}

	current := data
	for i := 0; i < len(keys)-1; i++ {
		key := keys[i]
		val, exists := current[key]
		if !exists {
			// Create intermediate map
			newMap := make(map[string]any)
			current[key] = newMap
			current = newMap
		} else {
			// Check if it's a map
			m, ok := val.(map[string]any)
			if !ok {
				return ErrNotAMap
			}
			current = m
		}
	}

	// Set the final value
	current[keys[len(keys)-1]] = value
	return nil
}

// deleteNestedValue removes a key from nested maps
func deleteNestedValue(data map[string]any, keys []string) error {
	if len(keys) == 0 {
		return ErrInvalidKey
	}

	if len(keys) == 1 {
		delete(data, keys[0])
		return nil
	}

	// Navigate to parent
	current := data
	for i := 0; i < len(keys)-1; i++ {
		val, exists := current[keys[i]]
		if !exists {
			return ErrKeyNotFound
		}
		m, ok := val.(map[string]any)
		if !ok {
			return ErrNotAMap
		}
		current = m
	}

	// Delete the final key
	delete(current, keys[len(keys)-1])
	return nil
}

// ParseValue attempts to parse string values into appropriate types
// "true" -> bool(true), "123" -> int64(123), "3.14" -> float64(3.14), else string
func ParseValue(s string) any {
	// Try boolean
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Try integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		return i
	}

	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Default to string
	return s
}

// toEnvKey converts dot notation to environment variable format
// "common.api_keys.tiingo" -> "ITDA_COMMON_API_KEYS_TIINGO"
func toEnvKey(dotKey string) string {
	upper := strings.ToUpper(dotKey)
	underscored := strings.ReplaceAll(upper, ".", "_")
	return "ITDA_" + underscored
}
