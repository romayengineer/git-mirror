package pathutil

import "testing"

func TestParseKeyValuePairs(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected map[string]string
	}{
		{name: "empty string", input: "", expected: map[string]string{}},
		{name: "single pair", input: "key1:value1", expected: map[string]string{"key1": "value1"}},
		{name: "multiple pairs", input: "key1:value1,key2:value2", expected: map[string]string{"key1": "value1", "key2": "value2"}},
		{name: "malformed pair skipped", input: "key1:value1,invalidpair,key2:value2", expected: map[string]string{"key1": "value1", "key2": "value2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseKeyValuePairs(tt.input)
			if len(result) != len(tt.expected) {
				t.Errorf("expected %d pairs, got %d", len(tt.expected), len(result))
			}
			for k, v := range tt.expected {
				if result[k] != v {
					t.Errorf("expected [%s]=%s, got [%s]=%s", k, v, k, result[k])
				}
			}
		})
	}
}

func TestExpandPath(t *testing.T) {
	got, err := ExpandPath("/abs/path")
	if err != nil || got != "/abs/path" {
		t.Errorf("expected passthrough, got %q err %v", got, err)
	}
}
