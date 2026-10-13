package path

import "testing"

func TestWindowsConverter(t *testing.T) {
	converter := &windowsConverter{}

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "/mnt/c/Users/Admin/Documents/Projects",
			expected: "C:\\Users\\Admin\\Documents\\Projects",
		},
		{
			input:    "/mnt/d/repos/test",
			expected: "D:\\repos\\test",
		},
		{
			input:    "/home/user/project",
			expected: "/home/user/project",
		},
		{
			input:    "relative/path",
			expected: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := converter.Convert(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestLinuxConverter(t *testing.T) {
	converter := &linuxConverter{}

	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "/mnt/c/Users/Admin/Documents/Projects",
			expected: "/mnt/c/Users/Admin/Documents/Projects",
		},
		{
			input:    "/home/user/project",
			expected: "/home/user/project",
		},
		{
			input:    "relative/path",
			expected: "relative/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := converter.Convert(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestNewPathConverter(t *testing.T) {
	tests := []struct {
		format       string
		input        string
		expectedType string
		expected     string
	}{
		{
			format:       "windows",
			input:        "/mnt/c/path",
			expectedType: "windows",
			expected:     "C:\\path",
		},
		{
			format:       "WINDOWS",
			input:        "/mnt/c/path",
			expectedType: "windows",
			expected:     "C:\\path",
		},
		{
			format:       "linux",
			input:        "/mnt/c/path",
			expectedType: "linux",
			expected:     "/mnt/c/path",
		},
		{
			format:       "invalid",
			input:        "/mnt/c/path",
			expectedType: "linux",
			expected:     "/mnt/c/path",
		},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			converter := NewPathConverter(tt.format)
			result := converter.Convert(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}
