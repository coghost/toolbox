package pathlib

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

const (
	_mode500 = 0o500
	_mode555 = 0o555
	_mode600 = 0o600
	_mode644 = 0o644 // Read and write for owner, read for group and others
	_mode666 = 0o666
	_mode755 = 0o755 // Read, write, and execute for owner, read and execute for group and others
	_mode777 = 0o777
)

var (
	FileMode644 = os.FileMode(_mode644)
	DirMode755  = os.FileMode(_mode755)
)

var ErrCannotCreateSiblingDir = errors.New("cannot create sibling directory to parent: current path is at root or one level below")

// FsPath represents a file system entity with various properties
type FsPath struct {
	// Stem represents the base name of the file or directory without any suffix (file extension).
	// For files, it's the file name without the extension. For directories, it's the directory name.
	//
	// Examples:
	//   - For a file "/tmp/folder/file.txt", Stem would be "file"
	//   - For a directory "/tmp/folder/", Stem would be "folder"
	//   - For a file "document.tar.gz", Stem would be "document.tar"
	//
	// This field is useful when you need to work with the core name of a file or directory
	// without considering its extension or full path.
	Stem string

	// Name represents the base name of the file or directory, including any suffix (file extension).
	// For files, it's the file name with the extension. For directories, it's the same as the Name field.
	//
	// Examples:
	//   - For a file "/tmp/folder/file.txt", Name would be "file.txt"
	//   - For a directory "/tmp/folder/", Name would be "folder"
	//   - For a file "document.tar.gz", Name would be "document.tar.gz"
	//
	// This field is useful when you need to work with the full name of a file including its extension,
	// or when you need to distinguish between files with the same base name but different extensions.
	Name string

	// Suffix represents the file extension, including the leading dot.
	// For files without an extension or for directories, it will be an empty string.
	//
	// Examples:
	//   - For a file "document.txt", Suffix would be ".txt"
	//   - For a file "archive.tar.gz", Suffix would be ".gz"
	//   - For a file "README" or a directory, Suffix would be ""
	//
	// This field is useful for identifying file types, filtering files by extension,
	// or when you need to work with or modify file extensions.
	Suffix string

	// absPath represents the absolute path of the file or directory.
	// It provides the full, unambiguous location of the item in the file system.
	//
	// Examples:
	//   - For a file "/home/user/documents/report.pdf", absPath would be "/home/user/documents/report.pdf"
	//   - For a directory "/var/log/", absPath would be "/var/log"
	//   - For a relative input path "../../file.txt", absPath would resolve to something like "/absolute/path/to/file.txt"
	//
	// This field is crucial for operations that require the complete file path,
	// especially when working across different directories or when absolute paths are necessary.
	// It resolves any relative paths to their absolute form.
	absPath string

	RawPath string

	fs afero.Fs // The underlying file system
}

// Path creates and returns a new Entity from the given file path
func Path(filePath string) *FsPath {
	fs, err := PathE(filePath)
	if err != nil {
		panic(err)
	}

	return fs
}

func PathE(filePath string) (*FsPath, error) {
	fs := afero.NewOsFs()

	absPath, err := ResolveAbsPath(filePath)
	if err != nil {
		return nil, err
	}

	stem, name, suffix := parseFileName(filepath.Base(filePath))

	pth := &FsPath{
		absPath: absPath,
		Stem:    stem,
		Name:    name,
		Suffix:  suffix,
		RawPath: filePath,
		// Use OS file system by default
		fs: fs,
	}

	return pth, nil
}

func (p *FsPath) String() string {
	return p.absPath
}

func (p *FsPath) AbsPath() string {
	return p.absPath
}

func (p *FsPath) NoSuffix(str string) string {
	name := strings.TrimSuffix(p.Name, str)

	return p.Dir().Join(name).absPath
}

func (p *FsPath) Fs() afero.Fs {
	return p.fs
}

// Exists check file exists or not.
func (p *FsPath) Exists() bool {
	_, err := p.Stat()
	return err == nil || os.IsExist(err)
}

func (p *FsPath) Stat() (fs.FileInfo, error) {
	return p.fs.Stat(p.absPath)
}

// IsDir checks if the entity is a directory
func (p *FsPath) IsDir() bool {
	isDir := strings.HasSuffix(p.RawPath, "/")

	if !isDir {
		isDir, _ = afero.IsDir(p.fs, p.absPath)
	}

	return isDir
}

// Suffixes returns a list of the path's file extensions.
func (p *FsPath) Suffixes() []string {
	name := filepath.Base(p.absPath)
	if name == "." || name == "/" {
		return []string{}
	}

	// Special handling for hidden files
	if strings.HasPrefix(name, ".") {
		nameParts := strings.SplitN(name, ".", 2)
		if len(nameParts) == 1 {
			return []string{} // Hidden file without extension
		}

		name = nameParts[1] // Remove the leading dot
	}

	parts := strings.Split(name, ".")
	if len(parts) <= 1 {
		return []string{}
	}

	suffixes := make([]string, len(parts)-1)
	for i := 1; i < len(parts); i++ {
		suffixes[i-1] = "." + parts[i]
	}

	return suffixes
}

// e checks the last argument for an error and panics if one is found.
// It returns all input arguments unchanged.
//
// This method is used internally to implement "Must" variants of other methods,
// where an error should cause a panic rather than be returned.
//
// Example:
//
//	p.e(someFunction())  // panics if someFunction returns an error
func (p *FsPath) e(args ...interface{}) {
	err, ok := args[len(args)-1].(error)
	if ok {
		panic(err)
	}
}

func parseFileName(name string) (stem, fullName, suffix string) {
	fullName = name
	suffix = filepath.Ext(name)
	stem = strings.TrimSuffix(name, suffix)

	const hiddenDotLen = 3

	// Special handling for hidden files
	if strings.HasPrefix(name, ".") {
		if len(name) > 1 && strings.Contains(name[1:], ".") {
			// Hidden file with extension
			parts := strings.SplitN(name, ".", hiddenDotLen)
			if len(parts) == hiddenDotLen {
				stem = "." + parts[1]
				suffix = "." + parts[2]
			}
		} else {
			// Hidden file without extension or just "."
			stem = name
			suffix = ""
		}
	}

	return stem, fullName, suffix
}
