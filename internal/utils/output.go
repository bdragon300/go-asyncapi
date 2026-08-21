package utils

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// ReadAllWithLineNumbers prints the content of the reader with line numbers skipping skipLines lines that will be
// printed without line numbers. placeWidth is the space occupied by the line number including the separator.
func ReadAllWithLineNumbers(r io.Reader, skipLines int, placeWidth int) string {
	var b strings.Builder
	rd := bufio.NewReader(r)
	format := fmt.Sprintf("%%-%dd│ ", placeWidth-2)

	line := 1
	for i := 0; ; i++ {
		s, err := rd.ReadString('\n')
		if err != nil {
			break // Supposing that the only error can appear here is io.EOF
		}
		if i >= skipLines {
			b.WriteString(fmt.Sprintf(format, line))
			line++
		} else {
			b.WriteString(strings.Repeat(" ", placeWidth-2) + "│ ")
		}
		b.WriteString(s)
	}

	return b.String()
}

// TruncateANSIString truncates the string to the specified length, taking into account ANSI escape sequences.
// It returns the truncated string and a boolean indicating whether the string was truncated.
func TruncateANSIString(s string, maxLength int) (string, bool) {
	var b strings.Builder
	length := 0
	truncated := false

	for i := 0; i < len(s); {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			// Start of an ANSI escape sequence
			end := i + 2
			for end < len(s) && (s[end] < 'A' || s[end] > 'z') {
				end++
			}
			if end < len(s) {
				end++ // Include the final character of the escape sequence
			}
			b.WriteString(s[i:end])
			i = end
		} else {
			if length >= maxLength {
				truncated = true
				break
			}
			b.WriteByte(s[i])
			length++
			i++
		}
	}

	return b.String(), truncated
}
