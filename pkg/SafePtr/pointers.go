package SafePtr

func String(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func Int32(value *int32) int32 {
	if value == nil {
		return 0
	}
	return *value
}

func Int64(value *int64) int64 {
	if value == nil {
		return 0
	}
	return *value
}

func Bool(value *bool) bool {
	if value == nil {
		return false
	}
	return *value
}
