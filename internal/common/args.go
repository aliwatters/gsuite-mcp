package common

// ParseStringArg extracts a string argument from the request.
// Returns defaultVal if the argument is missing or empty.
func ParseStringArg(args map[string]any, key string, defaultVal string) string {
	if val, ok := args[key].(string); ok && val != "" {
		return val
	}
	return defaultVal
}

// ParseBoolArg extracts a boolean argument.
// Returns defaultVal if the argument is missing or invalid.
func ParseBoolArg(args map[string]any, key string, defaultVal bool) bool {
	if val, ok := args[key].(bool); ok {
		return val
	}
	return defaultVal
}

// ParseMaxResults extracts 'max_results' argument and enforces limits.
func ParseMaxResults(args map[string]any, defaultVal, maxLimit int64) int64 {
	maxResults := defaultVal
	if max, ok := args["max_results"].(float64); ok && max > 0 {
		maxResults = int64(max)
		if maxResults > maxLimit {
			maxResults = maxLimit
		}
	}
	return maxResults
}
