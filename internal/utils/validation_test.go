package utils

import "testing"

func TestIsValidName(t *testing.T) {
	tests := []struct {
		name  string
		valid bool
	}{
		{"root", true},
		{"my_user", true},
		{"my-user", true},
		{"user123", true},
		{"_hidden", true},
		{"123user", false}, // Must start with letter check? Regex says ^[a-z_]
		{"User", false},    // Uppercase usually discouraged/invalid in strict mode (regex is lowercase only)
		{"user@name", false},
		{"user name", false},
		{"", false},
	}

	for _, tt := range tests {
		if got := IsValidName(tt.name); got != tt.valid {
			t.Errorf("IsValidName(%q) = %v, want %v", tt.name, got, tt.valid)
		}
	}
}

func TestIsOneOf(t *testing.T) {
	allowed := []string{"present", "absent"}

	if !IsOneOf("present", allowed...) {
		t.Error("IsOneOf('present') should be true")
	}
	if !IsOneOf("absent", allowed...) {
		t.Error("IsOneOf('absent') should be true")
	}
	if IsOneOf("invalid", allowed...) {
		t.Error("IsOneOf('invalid') should be false")
	}
	if IsOneOf("", allowed...) {
		t.Error("IsOneOf('') should be false")
	}
}
