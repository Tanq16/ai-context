package utils

type Dictionary map[string]any

// UnwindValue safely navigates and retrieves values from nested maps
func (d Dictionary) UnwindValue(keys ...string) any {
	current := d
	for _, key := range keys {
		if current == nil {
			return nil
		}
		if val, ok := current[key]; ok {
			switch v := val.(type) {
			case map[string]any:
				current = Dictionary(v)
			case Dictionary:
				current = v
			default:
				return val // Gracefully return the value if it's not a map
			}
		} else {
			return nil
		}
	}
	return current
}

func (d Dictionary) UnwindString(keys ...string) string {
	if val := d.UnwindValue(keys...); val != nil {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func (d Dictionary) UnwindBool(keys ...string) bool {
	if val := d.UnwindValue(keys...); val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}

func (d Dictionary) UnwindFloat64(keys ...string) float64 {
	if val := d.UnwindValue(keys...); val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0.0
}

func (d Dictionary) UnwindFloat32(keys ...string) float32 {
	if val := d.UnwindValue(keys...); val != nil {
		switch v := val.(type) {
		case float32:
			return v
		case float64:
			return float32(v)
		case int:
			return float32(v)
		case int32:
			return float32(v)
		case int64:
			return float32(v)
		}
	}
	return 0.0
}

func (d Dictionary) UnwindInt64(keys ...string) int64 {
	if val := d.UnwindValue(keys...); val != nil {
		switch v := val.(type) {
		case int64:
			return v
		case int32:
			return int64(v)
		case int:
			return int64(v)
		case float64:
			return int64(v)
		case float32:
			return int64(v)
		}
	}
	return 0
}

func (d Dictionary) UnwindInt32(keys ...string) int32 {
	if val := d.UnwindValue(keys...); val != nil {
		switch v := val.(type) {
		case int32:
			return v
		case int64:
			return int32(v)
		case int:
			return int32(v)
		case float64:
			return int32(v)
		case float32:
			return int32(v)
		}
	}
	return 0
}

func (d Dictionary) UnwindInt(keys ...string) int {
	if val := d.UnwindValue(keys...); val != nil {
		switch v := val.(type) {
		case int:
			return v
		case int32:
			return int(v)
		case int64:
			return int(v)
		case float64:
			return int(v)
		case float32:
			return int(v)
		}
	}
	return 0
}

func (d Dictionary) UnwindUint64(keys ...string) uint64 {
	if val := d.UnwindValue(keys...); val != nil {
		switch v := val.(type) {
		case uint64:
			return v
		case uint32:
			return uint64(v)
		case uint:
			return uint64(v)
		case float64:
			return uint64(v)
		case float32:
			return uint64(v)
		}
	}
	return 0
}

func (d Dictionary) UnwindUint(keys ...string) uint {
	if val := d.UnwindValue(keys...); val != nil {
		switch v := val.(type) {
		case uint:
			return v
		case uint32:
			return uint(v)
		case uint64:
			return uint(v)
		case float64:
			return uint(v)
		case float32:
			return uint(v)
		}
	}
	return 0
}

func (d Dictionary) UnwindSlice(keys ...string) []any {
	if val := d.UnwindValue(keys...); val != nil {
		if slice, ok := val.([]any); ok {
			return slice
		}
	}
	return nil
}

func (d Dictionary) UnwindMap(keys ...string) Dictionary {
	if val := d.UnwindValue(keys...); val != nil {
		if m, ok := val.(Dictionary); ok {
			return m
		}
		if m, ok := val.(map[string]any); ok {
			return Dictionary(m)
		}
	}
	return nil
}
