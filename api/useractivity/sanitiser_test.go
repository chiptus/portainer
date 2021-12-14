package useractivity

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Sanitise(t *testing.T) {
	type RawMap = map[string]interface{}
	tests := map[string]struct {
		input  RawMap
		output RawMap
	}{
		"nothing to sanitise": {
			input:  RawMap{"name": "a"},
			output: RawMap{"name": "a"},
		},
		"redact single password field": {
			input:  RawMap{"PaSswOrd": "a"},
			output: RawMap{"PaSswOrd": RedactedValue},
		},
		"redact multiple fields on the same level": {
			input:  RawMap{"PaSswOrd": "a", "stringData": "b"},
			output: RawMap{"PaSswOrd": RedactedValue, "stringData": RedactedValue},
		},
		"redact fields with a map value": {
			input:  RawMap{"PaSswOrd": RawMap{"one": 1, "two": 2}},
			output: RawMap{"PaSswOrd": RawMap{"one": RedactedValue, "two": RedactedValue}},
		},
		"complex structure with unredactred and redacted values": {
			input: RawMap{
				"LeAve":      "Alone",
				"StringData": RawMap{"one": 1, "two": 2},
				"This":       2,
				"Password":   "qwerty",
				"pwd":        "seCret",
				"nest":       RawMap{"password": "hide", "empty": 0},
			},
			output: RawMap{
				"LeAve":      "Alone",
				"StringData": RawMap{"one": RedactedValue, "two": RedactedValue},
				"This":       2,
				"Password":   RedactedValue,
				"pwd":        "seCret",
				"nest":       RawMap{"password": RedactedValue, "empty": 0},
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			Sanitise(tt.input)
			assert.Equal(t, tt.output, tt.input)
		})
	}
}
