package utils

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strings"
	"unicode/utf8"
)

const (
	defaultFiller = ' '
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

// ArrangeInColumns takes multiple readers and arranges their content in equal columns, separated by columnSeparator.
func ArrangeInColumns(contents []io.Reader, lineWidth int, columnSeparator string) (string, error) {
	var columns []*bufio.Scanner
	var widths []int

	ellipsisBytes := make([]byte, utf8.RuneLen('…'))
	utf8.EncodeRune(ellipsisBytes, '…')

	contentWidth := lineWidth - utf8.RuneCountInString(columnSeparator)*(len(contents)-1)
	w := contentWidth
	for i := range contents {
		s := bufio.NewScanner(contents[i])
		s.Split(bufio.ScanRunes)
		columns = append(columns, s)

		widths = append(widths, contentWidth/len(contents))
		w -= widths[i]
	}
	if w > 0 {
		widths[len(widths)-1] += w
	}

	var b strings.Builder
	for {
		eof := make([]bool, len(columns))

		for col := 0; col < len(columns); col++ {
			var hasContent bool
			spaceLeft := widths[col]
			for spaceLeft > 0 {
				if hasContent = columns[col].Scan(); !hasContent {
					if err := columns[col].Err(); err != nil {
						return "", err
					}
					break // Page has no more content
				}

				char := columns[col].Bytes() // Next rune bytes
				if char[0] == '\r' {         //nolint:gocritic
					continue // Skip symbol
				} else if char[0] == '\n' {
					break // Line has no more content
				} else if spaceLeft == 1 {
					char = ellipsisBytes
				}

				b.Write(char)
				spaceLeft--
			}

			// Fast-forward scanner to the end of the line
			for hasContent && columns[col].Bytes()[0] != '\n' {
				hasContent = columns[col].Scan()
			}

			// Fill the remaining space with filler
			for ; spaceLeft > 0; spaceLeft-- {
				b.WriteRune(defaultFiller)
			}

			// columns separator
			if col < len(columns)-1 {
				for _, r := range columnSeparator {
					b.WriteRune(r)
				}
			}
			eof[col] = !hasContent
		}

		if !slices.Contains(eof, false) {
			return b.String(), nil // End of output
		}
		b.WriteRune('\n')
	}
}
