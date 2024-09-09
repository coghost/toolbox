package pathlib

import (
	"io"
	"os"
	"path/filepath"
)

func (p *FsPath) Copy(newfile string) error {
	sourceFile, err := p.Reader()
	if err != nil {
		return err
	}

	destFile, err := os.Create(newfile)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	si, err := p.Stat()
	if err == nil {
		err = os.Chmod(newfile, si.Mode())
	}

	return err
}

// MkDirs quick create dir for given path.
func (p *FsPath) MkDirs() error {
	return p.fs.MkdirAll(p.absPath, DirMode755)
}

// MkParentDir creates the parent directory for the given path
func (p *FsPath) MkParentDir() error {
	return p.fs.MkdirAll(filepath.Dir(p.absPath), DirMode755)
}

func (p *FsPath) Move(newfile string) error {
	return p.Rename(newfile)
}

// Rename moves the file to a new location
func (p *FsPath) Rename(newfile string) error {
	return p.fs.Rename(p.absPath, newfile)
}
