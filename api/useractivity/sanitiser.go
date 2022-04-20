package useractivity

import "strings"

const (
	// RedactedValue is used for cleared fields
	RedactedValue = "[REDACTED]"
)

var blackList = map[string]struct{}{
	"azureauthenticationkey": {},
	"binarydata":             {},
	"clientsecret":           {},
	"data":                   {},
	"newpassword":            {},
	"password":               {},
	"repositorypassword":     {},
	"stringdata":             {},
	"tlscacertfile":          {},
	"tlscertfile":            {},
	"tlskeyfile":             {},
	"civoapikey":             {},
	"digitaloceantoken":      {},
	"linodetoken":            {},
}

// Sanitise removes possibly sensitive content from the map.
// The map will be updated in-place.
func Sanitise(val map[string]interface{}) map[string]interface{} {
	// Values of all black-listed fields will be replaced.
	// If a sensitive key holds a complex object, such objects will be obfuscated:
	// - top level keys will be kept
	// - values will be redacted
	for k, v := range val {
		if _, ok := blackList[strings.ToLower(k)]; ok {
			switch v := v.(type) {
			case map[string]interface{}:
				for kk := range v {
					v[kk] = RedactedValue
				}
				val[k] = v
			default:
				val[k] = RedactedValue
			}
			continue
		}

		if m, ok := v.(map[string]interface{}); ok {
			val[k] = Sanitise(m)
		}
	}

	return val
}
