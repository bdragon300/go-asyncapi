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
