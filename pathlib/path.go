package pathlib

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gookit/goutil/fsutil"
	"github.com/ungerik/go-dry"
)

// FSPath represents a file system entity with various properties
type FSPath struct {
	filepath string

	// Name the name of a file, no suffix include.
	//  e.g. filepath is: `/tmp/a/b.json`
	//  the Name is `b`
	Name           string
	NameWithSuffix string
	WorkingDir     string
	BaseDir        string // New field

	// Suffix eg: path/to/main.go => ".go"
	Suffix  string
	AbsPath string
	isDir   bool
}

// Path creates and returns a new Entity from the given file path
func Path(filePath string) *FSPath {
	nameWithSuffix := fsutil.Name(filePath)
	suffix := fsutil.Suffix(filePath)
	name := strings.TrimSuffix(nameWithSuffix, suffix)
	absPath := fsutil.Expand(filePath)
	isDir := fsutil.IsDir(absPath)

	working := absPath
	if !isDir {
		working = filepath.Dir(working)
	}

	pth := &FSPath{
		filepath:   filePath,
		AbsPath:    absPath,
		Name:       name,
		Suffix:     suffix,
		isDir:      isDir,
		WorkingDir: strings.TrimSuffix(working, "/"),
		// Set the BaseDir
		BaseDir: filepath.Base(working),

		NameWithSuffix: nameWithSuffix,
	}

	return pth
}

// Exists check file exists or not.
func (p *FSPath) Exists() bool {
	if p.IsDir() {
		return fsutil.DirExist(p.AbsPath)
	}
	return fsutil.FileExists(p.AbsPath)
}

// IsDir checks if the entity is a directory
func (p *FSPath) IsDir() bool {
	return p.isDir
}

// OriginalName returns the original name passed in
func (p *FSPath) OriginalName() string {
	return p.filepath
}

// Parents returns the parent directory path up to the specified number of levels
func (p *FSPath) Parents(num int) string {
	return parents(p.AbsPath, num)
}

// parents is a helper function to get parent directory paths
func parents(abspath string, num int) string {
	dir := strings.TrimSuffix(abspath, "/")
	arr := strings.Split(dir, "/")

	if num >= len(arr) {
		return ""
	}

	return strings.Join(arr[:len(arr)-num], "/")
}

// MkParentDir creates the parent directory for the given path
func (p *FSPath) MkParentDir() error {
	return fsutil.MkParentDir(p.AbsPath)
}

// MkDirs quick create dir for given path.
func (p *FSPath) MkDirs() error {
	return os.MkdirAll(p.AbsPath, 0o755) //nolint:mnd
}

// MustSetString sets the file content as a string, panics on error
func (p *FSPath) MustSetString(data string) {
	dry.PanicIfErr(p.SetString(data))
}

// SetString sets the file content as a string
func (p *FSPath) SetString(data string) error {
	if err := p.MkParentDir(); err != nil {
		return err
	}

	return p.SetBytes([]byte(data))
}

func (p *FSPath) SetBytes(data []byte) error {
	return dry.FileSetBytes(p.AbsPath, data)
}

func (p *FSPath) GetString() (string, error) {
	return dry.FileGetString(p.AbsPath)
}

// Reader returns an io.Reader for the file
func (p *FSPath) Reader() (io.Reader, error) {
	return dry.FileBufferedReader(p.AbsPath)
}

func (p *FSPath) MustGetBytes() []byte {
	b, err := dry.FileGetBytes(p.AbsPath)
	dry.PanicIfErr(err)

	return b
}

func (p *FSPath) GetBytes() ([]byte, error) {
	return dry.FileGetBytes(p.AbsPath)
}

// SplitPath splits path immediately following the final Separator, separating it into a directory and file name component
func (p *FSPath) SplitPath(pathStr string) (dir, name string) {
	return filepath.Split(pathStr)
}

// MustNewFileInWD creates a new file in the working directory, panics on error
func (p *FSPath) MustNewFileInWD(name string) *FSPath {
	e, err := p.NewFileInWD(name)
	dry.PanicIfErr(err)

	return e
}

// NewFileInWD creates a new file with name in same folder.
func (p *FSPath) NewFileInWD(name string) (*FSPath, error) {
	dir := p.WorkingDir
	dstName := path.Join(dir, fmt.Sprintf("%s%s", name, p.Suffix))

	err := os.Rename(p.AbsPath, dstName)
	if err != nil {
		return nil, err
	}

	return Path(dstName), nil
}

// GenRelativeFile generates a new file path relative to the working directory.
//
// Parameters:
//   - name: A string representing the relative path and filename.
//
// Returns:
//
//	A string representing the new absolute file path.
//
// The method works as follows:
//  1. It joins the current working directory with the provided relative path.
//  2. It resolves any relative path components (like ".." or ".") in the result.
//
// Note:
//   - This method only generates the new path string. It does not actually create the file or directory.
//   - The provided name can include subdirectories (e.g., "subdir/file.txt").
//   - Using ".." in the name allows referencing parent directories.
//
// Examples:
//
//  1. Current working directory: "/tmp/a/b/"
//     name: "c.json"
//     Result: "/tmp/a/b/c.json"
//
//  2. Current working directory: "/tmp/a/b/"
//     name: "../c.json"
//     Result: "/tmp/a/c.json"
//
//  3. Current working directory: "/tmp/a/b/"
//     name: "subdir/c.json"
//     Result: "/tmp/a/b/subdir/c.json"
//
// This method is useful for generating new file paths relative to the current working directory,
// allowing for flexible file placement in the directory structure.
func (p *FSPath) GenRelativeFile(name string) string {
	return path.Join(p.WorkingDir, name)
}

// GenPathInSiblingDir generates a new Entity for the current file in a new sibling directory.
//
// Parameters:
//   - newDirName: The name of the new sibling directory to be used in the path.
//
// Returns:
//
//	A new *Entity representing the file path in the sibling directory.
//
// The function works as follows:
//  1. It determines the parent directory of the current working directory.
//  2. It constructs a path for a new directory at the same level as the current directory.
//  3. It generates a new file path by combining the new directory path with the current filename.
//  4. It creates and returns a new Entity based on this new path.
//
// Note:
//   - This function only generates a new Entity. It does not create any directories or move any files.
//   - The new directory will be at the same level as the current file's directory, not above it.
//   - If the current path is at the root directory, the new directory will be created at the root level.
//
// Examples:
//
//  1. Current file: "/tmp/a/b/file.txt", newDirName: "c"
//     Result: Entity representing "/tmp/a/c/file.txt"
//
//  2. Current file: "/tmp/a/b/", newDirName: "c"
//     Result: Entity representing "/tmp/a/c/b"
//
//  3. Current file: "/root.txt", newDirName: "newdir"
//     Result: Entity representing "/newdir/root.txt"
//
// This function is useful for generating Entities that represent files in parallel directory structures,
// allowing for easy manipulation and access to file properties in the new location.
func (p *FSPath) GenPathInSiblingDir(newDirName string) *FSPath {
	parentDir := filepath.Dir(p.WorkingDir)
	newFolder := filepath.Join(parentDir, newDirName)

	return Path(filepath.Join(newFolder, p.NameWithSuffix))
}

// GenFilePathWithNewSuffix generates a new file path with the given suffix, optionally creating a new folder.
//
// Parameters:
//   - newSuffix: The new file suffix (extension) to use, without the leading dot.
//   - createTypeFolder: If true, creates a new subfolder named after the new suffix.
//
// Returns:
//   - A new *FSPath representing the generated file path.
//
// This method generates a new file path based on the current file, changing its suffix
// and optionally placing it in a new subfolder. It works as follows:
//  1. It changes the file's suffix to the provided newSuffix.
//  2. If createTypeFolder is true, it creates a new subfolder named after the new suffix.
//
// The behavior changes slightly depending on the current file's location:
//   - For files not in the root directory:
//   - Without type folder: /path/to/file.txt -> /path/to/file.newSuffix
//   - With type folder:    /path/to/file.txt -> /path/to_newSuffix/file.newSuffix
//   - For files in the root directory:
//   - Without type folder: /file.txt -> /file.newSuffix
//   - With type folder:    /file.txt -> /_newSuffix/file.newSuffix
//
// Note: This method only generates a new file path. It does not actually create any files or folders.
//
// Examples:
//
//	file := Path("/tmp/a/b/c.txt")
//	newPath := file.GenFilePathWithNewSuffix("json", false)  // Results in "/tmp/a/b/c.json"
//	newPath := file.GenFilePathWithNewSuffix("json", true)   // Results in "/tmp/a/b_json/c.json"
//
//	rootFile := Path("/tmp/c.txt")
//	newPath := rootFile.GenFilePathWithNewSuffix("yaml", true)  // Results in "/tmp/_yaml/c.yaml"
func (p *FSPath) GenFilePathWithNewSuffix(suffix string, createTypeFolder bool) *FSPath {
	name := p.Name + "." + suffix
	pwd := p.WorkingDir

	if createTypeFolder {
		dir, last := fsutil.SplitPath(p.WorkingDir)

		if dir != "/" {
			name = fmt.Sprintf("%s_%s/%s", last, suffix, name)
			pwd = dir
		} else {
			name = fmt.Sprintf("_%s/%s", suffix, name)
		}
	}

	return Path(path.Join(pwd, name))
}

// Copy copies the file to a new location
func (p *FSPath) Copy(name string) error {
	return dry.FileCopy(p.AbsPath, name)
}

// Move moves the file to a new location
func (p *FSPath) Move(fullpath string) error {
	return os.Rename(p.AbsPath, fullpath)
}

// MustCSVGetSlices reads the CSV file and returns its content as slices, panics on error
func (p *FSPath) MustCSVGetSlices() [][]string {
	arr, err := p.CSVGetSlices()
	dry.PanicIfErr(err)

	return arr
}

// CSVGetSlices reads the CSV file and returns its content as slices
func (p *FSPath) CSVGetSlices() ([][]string, error) {
	csvReader, err := p.Reader()
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(csvReader)
	r.Comma = ','
	r.Comment = '#'

	return r.ReadAll()
}

// ListFilesWithGlob lists files in the working directory matching the given pattern
func (p *FSPath) ListFilesWithGlob(pattern string) ([]string, error) {
	return ListFilesWithGlob(p.WorkingDir, pattern)
}

// ListFilesWithGlob lists files in the given root directory matching the given pattern
func ListFilesWithGlob(rootDir, pattern string) ([]string, error) {
	if pattern == "" {
		pattern = "*"
	}

	return filepath.Glob(fsutil.Expand(rootDir) + "/" + pattern)
}
