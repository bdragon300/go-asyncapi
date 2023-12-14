package utils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
)

func CopyRecursive(srcFS fs.FS, dstBase string, copyCb func(w io.Writer, r io.Reader) (int64, error)) (int, error) {
	var totalBytes int
	entries, err := fs.ReadDir(srcFS, ".")
	if err != nil {
		return totalBytes, fmt.Errorf("list dir entries: %w", err)
	}
	for _, entry := range entries {
		dst := path.Join(dstBase, entry.Name())

		if entry.IsDir() {
			src, err := fs.Sub(srcFS, entry.Name())
			if err != nil {
				return totalBytes, fmt.Errorf("read src dir %q: %w", entry.Name(), err)
			}
			if err = os.MkdirAll(dst, os.ModePerm); err != nil {
				return totalBytes, fmt.Errorf("create a dst directory %q: %w", dst, err)
			}
			n, err := CopyRecursive(src, dst, copyCb)
			if err != nil {
				return totalBytes, fmt.Errorf("copy directory %q: %w", entry.Name(), err)
			}
			totalBytes += n
		} else {
			doCopy := func() error {
				srcFile, err := srcFS.Open(entry.Name())
				if err != nil {
					return fmt.Errorf("open src file for reading %q: %w", entry.Name(), err)
				}
				defer srcFile.Close()
				dstFile, err := os.Create(dst)
				if err != nil {
					return fmt.Errorf("create/truncate dst file %q: %w", dst, err)
				}
				defer dstFile.Close()
				n, err := copyCb(dstFile, srcFile)
				if err != nil {
					return fmt.Errorf("copy contents from %q to %q: %w", entry.Name(), dst, err)
				}
				totalBytes += int(n)
				return nil
			}
			if err := doCopy(); err != nil {
				return totalBytes, err
			}
		}
	}
	return totalBytes, nil
}
