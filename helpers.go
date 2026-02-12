package yfinance

import (
	"encoding/json"
	"io"
	"time"
)

// Helper functions for parsing JSON responses

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func getFloat64(m map[string]interface{}, key string) float64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return val
		case float32:
			return float64(val)
		case int:
			return float64(val)
		case int64:
			return float64(val)
		case json.Number:
			f, _ := val.Float64()
			return f
		}
	}
	return 0
}

func getInt(m map[string]interface{}, key string) int {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val)
		case float32:
			return int(val)
		case int:
			return val
		case int64:
			return int(val)
		case json.Number:
			f, _ := val.Float64()
			return int(f)
		}
	}
	return 0
}

func getInt64(m map[string]interface{}, key string) int64 {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return int64(val)
		case float32:
			return int64(val)
		case int:
			return int64(val)
		case int64:
			return val
		case json.Number:
			f, _ := val.Float64()
			return int64(f)
		}
	}
	return 0
}

func getTime(m map[string]interface{}, key string) time.Time {
	if v, ok := m[key]; ok {
		switch val := v.(type) {
		case float64:
			return time.Unix(int64(val), 0)
		case int64:
			return time.Unix(val, 0)
		case string:
			t, _ := time.Parse(time.RFC3339, val)
			return t
		}
	}
	return time.Time{}
}

func parseJSONResponse(body io.ReadCloser, v interface{}) error {
	data, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func joinModules(modules []string) string {
	result := ""
	for i, m := range modules {
		if i > 0 {
			result += ","
		}
		result += m
	}
	return result
}
