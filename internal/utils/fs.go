package utils

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
)

func CopyRecursive(srcFS fs.FS, dstBase string, copyCb func(w io.Writer, r io.Reader) (int64, error)) error {
	entries, err := fs.ReadDir(srcFS, ".")
	if err != nil {
		return fmt.Errorf("cannot get dir entries: %w", err)
	}
	for _, entry := range entries {
		dst := path.Join(dstBase, entry.Name())

		if entry.IsDir() {
			src, err := fs.Sub(srcFS, entry.Name())
			if err != nil {
				return fmt.Errorf("cannot read src dir %q: %w", entry.Name(), err)
			}
			if err := os.MkdirAll(dst, os.ModePerm); err != nil {
				return fmt.Errorf("cannot create a dst directory %q: %w", dst, err)
			}
			if err := CopyRecursive(src, dst, copyCb); err != nil {
				return fmt.Errorf("error while copy dir %q: %w", entry.Name(), err)
			}
		} else {
			doCopy := func() error {
				srcFile, err := srcFS.Open(entry.Name())
				if err != nil {
					return fmt.Errorf("cannot open src file for reading %q: %w", entry.Name(), err)
				}
				defer srcFile.Close()
				dstFile, err := os.Create(dst)
				if err != nil {
					return fmt.Errorf("cannot craate/truncate dst file %q: %w", dst, err)
				}
				defer dstFile.Close()
				if _, err = copyCb(dstFile, srcFile); err != nil {
					return fmt.Errorf("cannot copy contents from %q to %q: %w", entry.Name(), dst, err)
				}
				return nil
			}
			if err := doCopy(); err != nil {
				return err
			}
		}
	}
	return nil
}
