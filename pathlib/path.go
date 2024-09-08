package pathlib

import (
	"encoding/csv"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/afero"
)

const (
	_mode644 = 0o644 // Read and write for owner, read for group and others
	_mode755 = 0o755 // Read, write, and execute for owner, read and execute for group and others
)

var (
	FileMode644 = os.FileMode(_mode644)
	DirMode755  = os.FileMode(_mode755)
)

var ErrCannotCreateSiblingDir = errors.New("cannot create sibling directory to parent: current path is at root or one level below")

// FSPath represents a file system entity with various properties
type FSPath struct {
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
	//   - For a file "archive.tar.gz", Suffix would be ".tar.gz"
	//   - For a file "README" or a directory, Suffix would be ""
	//
	// This field is useful for identifying file types, filtering files by extension,
	// or when you need to work with or modify file extensions.
	Suffix string

	// AbsPath represents the absolute path of the file or directory.
	// It provides the full, unambiguous location of the item in the file system.
	//
	// Examples:
	//   - For a file "/home/user/documents/report.pdf", AbsPath would be "/home/user/documents/report.pdf"
	//   - For a directory "/var/log/", AbsPath would be "/var/log"
	//   - For a relative input path "../../file.txt", AbsPath would resolve to something like "/absolute/path/to/file.txt"
	//
	// This field is crucial for operations that require the complete file path,
	// especially when working across different directories or when absolute paths are necessary.
	// It resolves any relative paths to their absolute form.
	AbsPath string

	RawPath string

	fs afero.Fs // The underlying file system
}

// Path creates and returns a new Entity from the given file path
func Path(filePath string) *FSPath {
	fs := afero.NewOsFs()

	absPath := resolveAbsPath(filePath)
	name := filepath.Base(filePath)
	suffix := filepath.Ext(filePath)
	stem := strings.TrimSuffix(name, suffix)

	pth := &FSPath{
		AbsPath: absPath,
		Stem:    stem,
		Name:    name,
		Suffix:  suffix,
		RawPath: filePath,
		// Use OS file system by default
		fs: fs,
	}

	return pth
}

func (p *FSPath) String() string {
	return p.AbsPath
}

// Exists check file exists or not.
func (p *FSPath) Exists() bool {
	_, err := p.Stat()
	return err == nil
}

func (p *FSPath) Stat() (fs.FileInfo, error) {
	return p.fs.Stat(p.AbsPath)
}

// IsDir checks if the entity is a directory
func (p *FSPath) IsDir() bool {
	isDir := strings.HasSuffix(p.RawPath, "/")

	if !isDir {
		isDir, _ = afero.IsDir(p.fs, p.AbsPath)
	}

	return isDir
}

func (p *FSPath) Dir() *FSPath {
	if p.IsDir() {
		return p // If it's already a directory, return itself
	}

	// For files, return the parent directory
	return Path(filepath.Dir(p.AbsPath))
}

// BaseDir returns the name of the directory containing the file or directory represented by this FSPath.
//
// For a file path, it returns the name of the directory containing the file.
// For a directory path, it returns the name of the directory itself.
// For the root directory, it returns "/".
//
// This method is useful when you need to know the name of the immediate parent directory
// without getting the full path to that directory.
//
// Returns:
//   - string: The name of the base directory.
//
// Examples:
//
//  1. For a file path "/home/user/documents/file.txt":
//     BaseDir() returns "documents"
//
//  2. For a directory path "/home/user/documents/":
//     BaseDir() returns "documents"
//
//  3. For a file in the root directory "/file.txt":
//     BaseDir() returns "/"
//
//  4. For the root directory "/":
//     BaseDir() returns "/"
//
// Note:
//   - This method does not check if the directory actually exists in the file system.
//   - It works with the absolute path of the FSPath, regardless of how the FSPath was originally created.
//
// Usage:
//
//	file := Path("/home/user/documents/file.txt")
//	baseDirName := file.BaseDir()
//	// baseDirName is "documents"
func (p *FSPath) BaseDir() string {
	return filepath.Base(p.Dir().AbsPath)
}

// Parts splits the path into its components.
// It takes the absolute path of the FSPath and returns a slice of strings,
// where each string is a component of the path.
//
// The function preserves the leading "/" if present in the original path.
//
// Examples:
//
//	For a path "/usr/bin/golang":
//	parts := path.Parts()
//	// parts will be []string{"/", "usr", "bin", "golang"}
//
//	For a path "usr/bin/golang" (without leading slash):
//	parts := path.Parts()
//	// parts will be []string{"usr", "bin", "golang"}
//
//	For the root directory "/":
//	parts := path.Parts()
//	// parts will be []string{"/"}
//
// Note:
//   - Trailing slashes are ignored.
//   - Empty components (resulting from consecutive slashes) are omitted.
func (p *FSPath) Parts() []string {
	// If the path is empty, return an empty slice
	if p.AbsPath == "" {
		return []string{}
	}

	// Trim trailing slashes
	trimmed := strings.TrimRight(p.AbsPath, "/")

	// If the path was just "/", return a slice with only "/"
	if trimmed == "" {
		return []string{"/"}
	}

	// Split the path
	parts := strings.Split(trimmed, "/")

	// If the original path started with a slash, add it as the first element
	if strings.HasPrefix(p.AbsPath, "/") {
		parts = append([]string{"/"}, parts...)
	}

	// Filter out empty parts
	var result []string

	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// Parents returns the parent directory path up to the specified number of levels.
//
// Parameters:
//   - num: The number of directory levels to go up.
//
// Returns:
//   - A new FSPath instance representing the parent directory path.
//
// This method traverses up the directory tree by the specified number of levels.
// It works with both absolute and relative paths.
//
// Examples:
//
//  1. For an absolute path "/home/user/documents/file.txt":
//     - Parents(1) returns FSPath("/home/user/documents")
//     - Parents(2) returns FSPath("/home/user")
//     - Parents(3) returns FSPath("/home")
//     - Parents(4) or higher returns FSPath("/")
//
// Note:
//   - If num is 0, it returns the current path.
//   - If num is greater than or equal to the number of directories in the path,
//     it returns the root directory for absolute paths or "." for relative paths.
func (p *FSPath) Parents(num int) *FSPath {
	if num <= 0 {
		return p
	}

	current := p
	for i := 0; i < num; i++ {
		parent := current.Parent()
		if parent.AbsPath == current.AbsPath {
			// We've reached the root directory or "."
			return parent
		}

		current = parent
	}

	return current
}

// Parent returns the immediate parent directory path of the current path.
//
// Returns:
//
//	A string representing the parent directory path.
//
// This method is a convenience wrapper around Parents(1), providing quick access
// to the immediate parent directory.
//
// Behavior:
//   - For a file or directory not at the root, it returns the containing directory.
//   - For the root directory, it returns an empty string.
//
// Examples:
//
//  1. For a file "/home/user/documents/file.txt":
//     parent := path.Parent()
//     // parent is "/home/user/documents"
//
//  2. For a directory "/var/log/":
//     parent := path.Parent()
//     // parent is "/var"
//
//  3. For the root directory "/":
//     parent := path.Parent()
//     // parent is ""
//
// Note:
//   - This method does not check if the parent directory actually exists in the file system.
//   - It works with the absolute path of the FSPath, regardless of how the FSPath was originally created.
//
// Use cases:
//   - Quickly accessing the parent directory without specifying the number of levels to go up.
//   - Simplifying code when only the immediate parent is needed.
func (p *FSPath) Parent() *FSPath {
	if p.AbsPath == "/" {
		return p // Root directory is its own parent
	}

	parentPath := filepath.Dir(p.AbsPath)

	return Path(parentPath)
}

// MkParentDir creates the parent directory for the given path
func (p *FSPath) MkParentDir() error {
	return p.fs.MkdirAll(filepath.Dir(p.AbsPath), DirMode755)
}

// MkDirs quick create dir for given path.
func (p *FSPath) MkDirs() error {
	return p.fs.MkdirAll(p.AbsPath, DirMode755)
}

// SplitPath splits the given path into its directory and file name components.
//
// Returns:
//   - dir: The directory portion of the path.
//   - name: The file name portion of the path.
//
// Examples:
//
//  1. For a file path "/home/user/file.txt":
//     - dir would be "/home/user/"
//     - name would be "file.txt"
//
//  2. For a directory path "/home/user/docs/":
//     - dir would be "/home/user/docs/"
//     - name would be ""
//
//  3. For a root-level file "/file.txt":
//     - dir would be "/"
//     - name would be "file.txt"
//
// Note: This method uses filepath.Split internally and works with both file and directory paths.
func (p *FSPath) Split(pathStr string) (dir, name string) {
	return filepath.Split(pathStr)
}

// JoinPath joins one or more path components to the current path.
//
// If the current FSPath represents a file, the method joins the components
// to the parent directory of the file. If it's a directory, it joins directly
// to the current path.
//
// Parameters:
//   - others: One or more path components to join to the current path.
//
// Returns:
//   - A new FSPath instance representing the joined path.
//
// Examples:
//
//  1. For a directory "/home/user":
//     path.JoinPath("documents", "file.txt") returns a new FSPath for "/home/user/documents/file.txt"
//
//  2. For a file "/home/user/file.txt":
//     path.JoinPath("documents", "newfile.txt") returns a new FSPath for "/home/user/documents/newfile.txt"
//
//  3. For a path "/":
//     path.JoinPath("etc", "config") returns a new FSPath for "/etc/config"
//
// Note: This method does not modify the original FSPath instance or create any directories.
// It only returns a new FSPath instance representing the joined path.
func (p *FSPath) JoinPath(others ...string) *FSPath {
	if len(others) > 0 && filepath.IsAbs(others[0]) {
		// If the first component is an absolute path, use it as the base
		return Path(filepath.Join(others...))
	}

	components := append([]string{p.AbsPath}, others...)

	return Path(filepath.Join(components...))
}

// WithName returns a new FSPath with the name changed.
//
// This method creates a new FSPath instance with the same parent directory as the original,
// but with a different name. It's similar to Python's pathlib.Path.with_name() method.
//
// Parameters:
//   - name: The new name for the file or directory.
//
// Returns:
//   - A new FSPath instance with the updated name.
//
// Examples:
//
//  1. For a file path "/home/user/file.txt":
//     path.WithName("newfile.txt") would return a new FSPath for "/home/user/newfile.txt"
//
//  2. For a directory path "/home/user/docs/":
//     path.WithName("newdocs") would return a new FSPath for "/home/user/newdocs"
//
//  3. For a root-level file "/file.txt":
//     path.WithName("newfile.txt") would return a new FSPath for "/newfile.txt"
//
// Note: This method does not actually rename the file or directory on the file system.
// It only creates a new FSPath instance with the updated name.
func (p *FSPath) WithName(name string) *FSPath {
	return p.Parent().JoinPath(name)
}

func (p *FSPath) WithStem(stem string) *FSPath {
	newName := stem + p.Suffix

	return p.Parent().JoinPath(newName)
}

func (p *FSPath) WithSuffix(suffix string) *FSPath {
	if suffix != "" && !strings.HasPrefix(suffix, ".") {
		suffix = "." + suffix
	}

	return Path(strings.TrimSuffix(p.AbsPath, p.Suffix) + suffix)
}

// WithRenamedParentDir creates a new FSPath with the parent directory renamed.
//
// This method generates a new FSPath that represents the current file or directory
// placed within a renamed parent directory. The new parent directory name replaces
// the current parent directory name at the same level in the path hierarchy.
//
// Parameters:
//   - newParentName: The new name for the parent directory.
//
// Returns:
//   - A new *FSPath instance representing the path with the renamed parent directory.
//
// Behavior:
//   - For files: It creates a new path with the file in the renamed parent directory.
//   - For directories: It creates a new path with the current directory as a subdirectory of the renamed parent.
//   - For the root directory: It returns the original FSPath without changes.
//
// Examples:
//
//  1. For a file "/tmp/a/b/file.txt" with newParentName "c":
//     Result: FSPath representing "/tmp/a/c/file.txt"
//
//  2. For a directory "/tmp/a/b/" with newParentName "c":
//     Result: FSPath representing "/tmp/a/c/b"
//
//  3. For a file "/file.txt" with newParentName "newdir":
//     Result: FSPath representing "/newdir/file.txt"
//
//  4. For the root directory "/" with any newParentName:
//     Result: FSPath representing "/" (unchanged)
//
// Note:
//   - This method only generates a new FSPath and does not actually rename directories or move files on the filesystem.
//   - The method preserves the original file name or the last directory name in the new path.
//   - If the current path is the root directory, the method returns the original path unchanged.
//
// Usage:
//
//	file := Path("/tmp/a/b/file.txt")
//	newPath := file.WithRenamedParentDir("c")
//	// newPath now represents "/tmp/a/c/file.txt"
func (p *FSPath) WithRenamedParentDir(newParentName string) *FSPath {
	// If the current path is the root directory, return the original FSPath
	if p.AbsPath == "/" {
		return p
	}

	// Get the parent directory
	parentDir := p.Parent()

	// Create a new FSPath for the new directory within the parent
	newDir := parentDir.WithName(newParentName)

	// Join the new directory with the current file name
	return newDir.JoinPath(p.Name)
}

// WithSuffixAndSuffixedParentDir generates a new file path with a changed suffix,
// and places it in a new directory named with the same suffix appended to the original parent directory name.
//
// Parameters:
//   - newSuffix: The new file suffix (extension) to use, with or without the leading dot.
//
// Returns:
//   - A new *FSPath representing the generated file path.
//
// Behavior:
//  1. Changes the file's suffix to the provided newSuffix.
//  2. Appends the new suffix (without the dot) to the current parent directory name.
//  3. For files in the root directory: Creates a new directory prefixed with an underscore and the new suffix (without the dot).
//
// Examples:
//
//  1. File in subdirectory:
//     "/path/to/file.txt" -> "/path/to_json/file.json"
//
//  2. File in root directory:
//     "/file.txt" -> "/_json/file.json"
//
// Notes:
//   - This method only generates a new FSPath and does not actually create any files or directories.
//   - If newSuffix is empty, the resulting file will have no extension, but the parent directory will still be renamed.
//
// Usage:
//
//	file := Path("/tmp/docs/report.txt")
//	newPath := file.WithSuffixAndSuffixedParentDir(".pdf")
//	// newPath now represents "/tmp/docs_pdf/report.pdf"
func (p *FSPath) WithSuffixAndSuffixedParentDir(newSuffix string) *FSPath {
	// Ensure the new suffix starts with a dot
	if newSuffix != "" && !strings.HasPrefix(newSuffix, ".") {
		newSuffix = "." + newSuffix
	}

	// Remove the dot from the suffix for the directory name
	dirSuffix := strings.TrimPrefix(newSuffix, ".")

	// Change the file suffix
	newPath := p.WithSuffix(newSuffix)

	// Handle root directory case
	if p.Parent().AbsPath == "/" {
		return p.Parent().JoinPath("_"+dirSuffix, newPath.Name)
	}

	// Create new directory name and rename the parent directory
	newDirName := p.Parent().Name + "_" + dirSuffix

	// Use WithRenamedParentDir to create the new path
	return newPath.WithRenamedParentDir(newDirName)
}

func (p *FSPath) Copy(newfile string) error {
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

func (p *FSPath) Move(newfile string) error {
	return p.Rename(newfile)
}

// Rename moves the file to a new location
func (p *FSPath) Rename(newfile string) error {
	return p.fs.Rename(p.AbsPath, newfile)
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
func (p *FSPath) e(args ...interface{}) {
	err, ok := args[len(args)-1].(error)
	if ok {
		panic(err)
	}
}

// MustSetString sets the file content as a string, panics on error
func (p *FSPath) MustSetString(data string) {
	p.e(p.SetString(data))
}

// SetString sets the file content as a string
func (p *FSPath) SetString(data string) error {
	return p.SetBytes([]byte(data))
}

func (p *FSPath) SetBytes(data []byte) error {
	if err := p.MkParentDir(); err != nil {
		return err
	}

	return afero.WriteFile(p.fs, p.AbsPath, data, FileMode644)
}

func (p *FSPath) GetString() (string, error) {
	data, err := p.GetBytes()
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Reader returns an io.Reader for the file
func (p *FSPath) Reader() (io.Reader, error) {
	return p.fs.Open(p.AbsPath)
}

func (p *FSPath) MustGetBytes() []byte {
	b, err := p.GetBytes()
	p.e(err)

	return b
}

func (p *FSPath) GetBytes() ([]byte, error) {
	return afero.ReadFile(p.fs, p.AbsPath)
}

// CSVGetSlices reads the CSV file and returns its content as slices.
//
// This method reads the file at the FSPath's location as a CSV (Comma-Separated Values) file
// and returns its content as a slice of string slices. Each inner slice represents a row
// in the CSV file, with each string in the slice representing a field in that row.
//
// Returns:
//   - [][]string: A slice of slices containing the CSV data.
//   - error: An error if the file cannot be read or parsed.
//
// The method uses a comma (',') as the field separator and '#' as the comment character.
// Empty lines and lines starting with '#' (after trimming spaces) are skipped.
//
// Example usage:
//
//	data, err := path.CSVGetSlices()
//	if err != nil {
//		// handle error
//	}
//	for _, row := range data {
//		// process each row
//	}
//
// Note: This method reads the entire file into memory. For very large files,
// consider using a streaming approach instead.
func (p *FSPath) CSVGetSlices() ([][]string, error) {
	return p.readDelimitedFile(',')
}

// TSVGetSlices reads the TSV file and returns its content as slices.
//
// This method reads the file at the FSPath's location as a TSV (Tab-Separated Values) file
// and returns its content as a slice of string slices. Each inner slice represents a row
// in the TSV file, with each string in the slice representing a field in that row.
//
// Returns:
//   - [][]string: A slice of slices containing the TSV data.
//   - error: An error if the file cannot be read or parsed.
//
// The method uses a tab character ('\t') as the field separator and '#' as the comment character.
// Empty lines and lines starting with '#' (after trimming spaces) are skipped.
//
// Example usage:
//
//	data, err := path.TSVGetSlices()
//	if err != nil {
//		// handle error
//	}
//	for _, row := range data {
//		// process each row
//	}
//
// Note: This method reads the entire file into memory. For very large files,
// consider using a streaming approach instead.
func (p *FSPath) TSVGetSlices() ([][]string, error) {
	return p.readDelimitedFile('\t')
}

// MustCSVGetSlices reads the CSV file and returns its content as slices, panics on error.
//
// This method is similar to CSVGetSlices, but it panics if an error occurs instead of
// returning the error. It's useful in situations where you're certain the file exists
// and is readable, or where you want to halt execution if the file cannot be processed.
//
// Returns:
//   - [][]string: A slice of slices containing the CSV data.
//
// Example usage:
//
//	data := path.MustCSVGetSlices()
//	for _, row := range data {
//		// process each row
//	}
//
// Note: Use this method with caution, as it will cause your program to panic
// if there's any issue reading or parsing the file.
func (p *FSPath) MustCSVGetSlices() [][]string {
	arr, err := p.CSVGetSlices()
	p.e(err)

	return arr
}

// MustTSVGetSlices reads the TSV file and returns its content as slices, panics on error.
//
// This method is similar to TSVGetSlices, but it panics if an error occurs instead of
// returning the error. It's useful in situations where you're certain the file exists
// and is readable, or where you want to halt execution if the file cannot be processed.
//
// Returns:
//   - [][]string: A slice of slices containing the TSV data.
//
// Example usage:
//
//	data := path.MustTSVGetSlices()
//	for _, row := range data {
//		// process each row
//	}
//
// Note: Use this method with caution, as it will cause your program to panic
// if there's any issue reading or parsing the file.
func (p *FSPath) MustTSVGetSlices() [][]string {
	arr, err := p.TSVGetSlices()
	p.e(err)

	return arr
}

func (p *FSPath) readDelimitedFile(separator rune) ([][]string, error) {
	reader, err := p.Reader()
	if err != nil {
		return nil, err
	}

	r := csv.NewReader(reader)
	r.Comma = separator
	r.Comment = '#'

	return r.ReadAll()
}

// ListFilesWithGlob lists files in the working directory matching the given pattern
func (p *FSPath) ListFilesWithGlob(pattern string) ([]string, error) {
	if pattern == "" {
		pattern = "*"
	}

	return afero.Glob(p.fs, filepath.Join(p.Dir().AbsPath, pattern))
}

// ListFilesWithGlob lists files in the given root directory matching the given pattern
func ListFilesWithGlob(rootDir, pattern string) ([]string, error) {
	if pattern == "" {
		pattern = "*"
	}

	return filepath.Glob(filepath.Join(Expand(rootDir), pattern))
}

func Expand(path string) string {
	// Return immediately if path is empty
	if path == "" {
		return path
	}

	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return path // Return original path if there's an error
		}

		if path == "~" {
			return home
		}

		return filepath.Join(home, path[2:])
	}

	return os.ExpandEnv(path)
}

// resolveAbsPath takes a file path and returns its absolute path.
//
// This function first expands any environment variables or user home directory
// references in the given path. Then, it converts the expanded path to an
// absolute path without resolving symbolic links.
//
// Parameters:
//   - filePath: A string representing the file path to be converted.
//
// Returns:
//   - string: The absolute path of the input file path.
//
// If the function encounters an error while converting to an absolute path,
// it returns the expanded path instead.
//
// Example:
//
//	absPath := resolveAbsPath("~/documents/file.txt")
//	// On a Unix system, this might return "/home/user/documents/file.txt"
//
// Note: This function does not check if the path actually exists in the file system.
func resolveAbsPath(filePath string) string {
	// Expand the path first
	expandedPath := Expand(filePath)

	// Convert to absolute path without resolving symlinks
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		// If we can't get the absolute path, use the expanded path
		return expandedPath
	}

	return absPath
}
