package util

import "strings"

// Get a key from a nested map. Not support array for the moment
func Get(mapObj map[string]interface{}, path string, key string) interface{} {
	if path == "" {
		return mapObj[key]
	}
	paths := strings.Split(path, ".")
	v := mapObj
	for _, p := range paths {
		if p == "" {
			continue
		}
		value, ok := v[p].(map[string]interface{})
		if ok {
			v = value
		} else {
			return ""
		}
	}
	return v[key]
}
