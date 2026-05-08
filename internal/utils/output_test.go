package utils

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

const (
	text11 = `
lorem ipsum dolor sit amet, consectetur adipiscing elit, 
sed 

do eiusmod tempor
incididunt ut labore et dolore magna aliqua. 

`
	text12 = `Duis aute irure dolor
in reprehenderit in voluptate velit esse cillum dolore eufugiat nulla 
pariatur.`

	unicodeText1 = `Hello 👋 World é 日本語 テスト
Foo 🚀 Bar
`
	unicodeText2 = `हैलो विश्व
Baz ⭐ Qux ñoño açúcar

e̊ fo中文文本Ελληνικά🎉مرحبا 🌍עבריתCafé ☕ Привет, мир`
)

type ColumnsReaderTestSuite struct {
	suite.Suite
}

func (s *ColumnsReaderTestSuite) TestArrangeInColumns_OneColumn_ShouldFitTextInWidth() {
	tests := []struct {
		name      string
		text      string
		expected  string
		separator string
	}{
		{
			name: "Text11",
			text: text11,
			expected: "                                        \n" +
				"lorem ipsum dolor sit amet, consectetur…\n" +
				"sed                                     \n" +
				"                                        \n" +
				"do eiusmod tempor                       \n" +
				"incididunt ut labore et dolore magna al…\n" +
				"                                        \n" +
				"                                        ",
			separator: "|",
		},
		{
			name: "Unicode Text2",
			text: unicodeText2,
			expected: "हैलो विश्व                              \n" +
				"Baz ⭐ Qux ñoño açúcar                   \n" +
				"                                        \n" +
				"e̊ fo中文文本Ελληνικά🎉مرحبا 🌍עבריתCafé ☕ Пр…",
			separator: " → ",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			readers := []io.Reader{
				io.Reader(strings.NewReader(tt.text)),
			}
			r, err := ArrangeInColumns(readers, 40, tt.separator)
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.expected, r)
		})
	}
}

func (s *ColumnsReaderTestSuite) TestArrangeInColumns_TwoColumns() {
	tests := []struct {
		name      string
		text1     string
		text2     string
		expected  string
		separator string
	}{
		{
			name:  "Text11 and Text12 with single char separator",
			text1: text11,
			text2: text12,
			expected: "                   |Duis aute irure dol…\n" +
				"lorem ipsum dolor …|in reprehenderit in…\n" +
				"sed                |pariatur.           \n" +
				"                   |                    \n" +
				"do eiusmod tempor  |                    \n" +
				"incididunt ut labo…|                    \n" +
				"                   |                    \n" +
				"                   |                    ",
			separator: "|",
		},
		{
			name:  "Text11 and Text12 with multichar separator",
			text1: text11,
			text2: text12,
			expected: "                   | Duis aute irure do…\n" +
				"lorem ipsum dolor… | in reprehenderit i…\n" +
				"sed                | pariatur.          \n" +
				"                   |                    \n" +
				"do eiusmod tempor  |                    \n" +
				"incididunt ut lab… |                    \n" +
				"                   |                    \n" +
				"                   |                    ",
			separator: " | ",
		},
		{
			name:  "Unicode Text1 and Unicode Text2",
			text1: unicodeText1,
			text2: unicodeText2,
			expected: "Hello 👋 World é 日… → हैलो विश्व         \n" +
				"Foo 🚀 Bar          → Baz ⭐ Qux ñoño açú…\n" +
				"                   →                    \n" +
				"                   → e̊ fo中文文本Ελληνικά🎉…",
			separator: " → ",
		},
	}
	for _, tt := range tests {
		s.Run(tt.name, func() {
			readers := []io.Reader{
				io.Reader(strings.NewReader(tt.text1)),
				io.Reader(strings.NewReader(tt.text2)),
			}
			r, err := ArrangeInColumns(readers, 40, tt.separator)
			assert.NoError(s.T(), err)
			assert.Equal(s.T(), tt.expected, r)
		})
	}
}

func (s *ColumnsReaderTestSuite) TestArrangeInColumns_EmptyReaders_ShouldPrintEmptyLine() {
	readers := []io.Reader{
		io.Reader(strings.NewReader("")),
		io.Reader(strings.NewReader("")),
	}
	const expected = `                   |                    `
	r, err := ArrangeInColumns(readers, 40, "|")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), expected, r)
}

func TestColumnsReaderTestSuite(t *testing.T) {
	suite.Run(t, new(ColumnsReaderTestSuite))
}
