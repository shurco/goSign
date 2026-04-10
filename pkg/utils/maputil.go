package utils

// GetStringFromMap safely extracts a string from map[string]any.
func GetStringFromMap(m map[string]any, key, def string) string {
	if m == nil {
		return def
	}
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return def
}

// GetBoolFromMap safely extracts a bool from map[string]any,
// supporting both native bool and string "true"/"false".
func GetBoolFromMap(m map[string]any, key string, def bool) bool {
	if m == nil {
		return def
	}
	if v, ok := m[key]; ok {
		if b, ok := v.(bool); ok {
			return b
		}
		if s, ok := v.(string); ok {
			return s == "true"
		}
	}
	return def
}
