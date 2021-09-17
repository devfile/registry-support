package server

import (
	"os"
	"testing"
)

func TestGetOptionalEnv(t *testing.T) {
	os.Setenv("SET_BOOL", "true")
	os.Setenv("SET_STRING", "test")

	tests := []struct {
		name         string
		key          string
		defaultValue interface{}
		want         interface{}
	}{
		{
			name:         "Test get SET_BOOL environment variable",
			key:          "SET_BOOL",
			defaultValue: false,
			want:         true,
		},
		{
			name:         "Test get unset bool environment variable",
			key:          "UNSET_BOOL",
			defaultValue: false,
			want:         false,
		},
		{
			name:         "Test get SET_STRING environment variable",
			key:          "SET_STRING",
			defaultValue: "anonymous",
			want:         "test",
		},
		{
			name:         "Test get unset string environment variable",
			key:          "UNSET_STRING",
			defaultValue: "anonymous",
			want:         "anonymous",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := getOptionalEnv(test.key, test.defaultValue)
			if value != test.want {
				t.Errorf("Got: %v, want: %v", value, test.want)
			}
		})
	}
}
