package writer

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"github.com/bdragon300/asyncapi-codegen-go/implementations"
	"github.com/bdragon300/asyncapi-codegen-go/internal/utils"
)

func WriteImplementation(implDir, baseDir string) (int, error) {
	if err := os.MkdirAll(baseDir, 0o750); err != nil {
		return 0, fmt.Errorf("create directory %q: %w", baseDir, err)
	}

	subDir, err := fs.Sub(implementations.Implementations, implDir)
	if err != nil {
		return 0, err
	}

	insertGeneratedPreamble := func(w io.Writer, r io.Reader) (n int64, err error) {
		rd := bufio.NewReader(r)
		line1, err := rd.ReadString('\n')
		if err != nil {
			return 0, err
		}
		if c, err := io.WriteString(w, line1); err == nil {
			n += int64(c)
		} else {
			return n, err
		}
		// Write a preamble only for go source code files
		if strings.HasPrefix(line1, "package") {
			if c, err := io.WriteString(w, "\n// "+GeneratedCodePreamble+"\n"); err == nil {
				n += int64(c)
			} else {
				return n, err
			}
		}
		if c, err := io.Copy(w, rd); err == nil {
			n += c
		} else {
			return n, err
		}
		return n, nil
	}

	return utils.CopyRecursive(subDir, baseDir, insertGeneratedPreamble)
}
