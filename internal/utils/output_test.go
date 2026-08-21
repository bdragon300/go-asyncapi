package utils

import "testing"

func TestTruncateANSIString(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		maxLength int
		want      string
		truncated bool
	}{
		{"No truncation", "\033[31mHello, World!\033[0m", 13, "\033[31mHello, World!\033[0m", false},
		{"Truncate with ANSI", "\033[31mHello\033[0m, World!\033[0m", 6, "\033[31mHello\033[0m,", true},
		{"Truncate with ANSI and bold", "\033[32mHello, \033[1mWorld!\033[0m", 10, "\033[32mHello, \033[1mWor", true},
		{"No ANSI codes", "Hello, World!", 5, "Hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, truncated := TruncateANSIString(tt.input, tt.maxLength)
			if got != tt.want || truncated != tt.truncated {
				t.Errorf("TruncateANSIString() = %v, %v; want %v, %v", got, truncated, tt.want, tt.truncated)
			}
		})
	}
}
