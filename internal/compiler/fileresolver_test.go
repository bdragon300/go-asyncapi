package compiler

import (
	"reflect"
	"testing"
)

func TestParseCommandLine(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			"normal",
			"hello world",
			[]string{"hello", "world"},
		},
		{
			"quote double",
			"hello \"world hello\"",
			[]string{"hello", "world hello"},
		},
		{
			"quote single",
			"hello 'world hello'",
			[]string{"hello", "world hello"},
		},
		{
			"nested quote double",
			"hello 'world \"hello\" foo'",
			[]string{"hello", "world \"hello\" foo"},
		},
		{
			"nested quote single",
			"hello \"world 'hello' foo\"",
			[]string{"hello", "world 'hello' foo"},
		},
		{
			"utf-8",
			"hello 世界",
			[]string{"hello", "世界"},
		},
		{
			"space",
			"hello\\ world",
			[]string{"hello world"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseCommandLine(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expect %v, got %v", tt.want, got)
			}
		})
	}
}
