package utils

import "testing"

func TestWords(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"aaa5aa5Bb5bCCDd", []string{"aaa5aa5", "Bb5b", "CC", "Dd"}},
		{"_aaaB_54", []string{"aaa", "B", "54"}},
		{"", nil},
	}

	for _, test := range tests {
		result := Words(test.input)
		if len(result) != len(test.expected) {
			t.Errorf("Words(%q) = %v; expected %v", test.input, result, test.expected)
			continue
		}
		for i := range result {
			if result[i] != test.expected[i] {
				t.Errorf("Words(%q) = %v; expected %v", test.input, result, test.expected)
				break
			}
		}
	}
}