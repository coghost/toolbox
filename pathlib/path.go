package pathlib

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
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

	// WorkingDir represents the absolute path of the working directory associated with this FSPath.
	// For files, it's the directory containing the file. For directories, it's the directory path itself.
	// It does not include a trailing slash.
	//
	// This field is useful for operations that need to work relative to the file's location,
	// such as creating new files in the same directory or performing glob operations.
	//
	// Examples:
	//   - For a file "/tmp/folder/file.txt", WorkingDir would be "/tmp/folder"
	//   - For a directory "/tmp/folder/", WorkingDir would be "/tmp/folder"
	WorkingDir string

	// BaseDir represents the name of the immediate parent directory of the file or directory.
	// For files or directories not at the root, it's the name of the containing folder.
	// For items at the root level, it will be the name of the root directory.
	//
	// Examples:
	//   - For a file "/tmp/folder/subfolder/file.txt", BaseDir would be "subfolder"
	//   - For a directory "/tmp/folder/subfolder/", BaseDir would be "folder"
	//   - For a file "/tmp/file.txt", BaseDir would be "tmp"
	//   - For the root directory "/", BaseDir would be ""
	//
	// This field is useful for operations that need to know or work with the immediate
	// parent directory name without dealing with the full path.
	BaseDir string

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

	filepath string
	isDir    bool

	fs afero.Fs // The underlying file system
}

// Path creates and returns a new Entity from the given file path
func Path(filePath string) *FSPath {
	fs := afero.NewOsFs()
	isDir := strings.HasSuffix(filePath, "/")

	absPath := getAbsPath(filePath)
	nameWithSuffix := filepath.Base(filePath)
	suffix := filepath.Ext(filePath)
	name := strings.TrimSuffix(nameWithSuffix, suffix)

	if !isDir {
		isDir, _ = afero.IsDir(fs, absPath)
	}

	working := determineWorkingDir(absPath, isDir)

	pth := &FSPath{
		AbsPath:    absPath,
		BaseDir:    filepath.Base(working),
		WorkingDir: working,

		Stem:   name,
		Name:   nameWithSuffix,
		Suffix: suffix,

		filepath: filePath,
		isDir:    isDir,

		fs: fs, // Use OS file system by default
	}

	return pth
}

// PathWithFs creates a new FSPath with a custom afero.Fs
func PathWithFs(filePath string, fs afero.Fs) *FSPath {
	pth := Path(filePath)
	pth.fs = fs

	return pth
}

// New function to handle absolute path logic
func getAbsPath(filePath string) string {
	// Expand the path first
	expandedPath := Expand(filePath)

	// Convert to absolute path without resolving symlinks
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		// If we can't get the absolute path, use the expanded path
		return expandedPath
	}

	// Preserve /var/folders/ prefix if it exists
	if strings.HasPrefix(absPath, "/private/var/folders/") {
		return strings.TrimPrefix(absPath, "/private")
	}

	return absPath
}

// New function to handle working directory logic
func determineWorkingDir(absPath string, isDir bool) string {
	var working string

	if isDir {
		working = absPath
	} else {
		working = filepath.Dir(absPath)
	}

	// Handle root directory case
	if working == string(os.PathSeparator) {
		return string(os.PathSeparator)
	}

	return strings.TrimSuffix(working, string(os.PathSeparator))
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
	return p.isDir
}

// OriginalName returns the original name passed in
func (p *FSPath) OriginalName() string {
	return p.filepath
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
//
//	A string representing the parent directory path.
//
// This method traverses up the directory tree by the specified number of levels.
// It works with both absolute and relative paths.
//
// Examples:
//
//  1. For an absolute path "/home/user/documents/file.txt":
//     - Parents(1) returns "/home/user/documents"
//     - Parents(2) returns "/home/user"
//     - Parents(3) returns "/home"
//     - Parents(4) returns "/"
//     - Parents(5) or higher returns "/"
//
//  2. For a relative path "user/documents/file.txt":
//     - Parents(1) returns "user/documents"
//     - Parents(2) returns "user"
//     - Parents(3) or higher returns ""
//
//  3. For an empty path "":
//     - Parents(n) returns "" for any n
//
// Note:
//   - If num is 0, it returns the current directory path.
//   - For absolute paths, if num is greater than or equal to the number of directories in the path,
//     it returns "/".
//   - For relative paths, if num is greater than or equal to the number of directories in the path,
//     it returns "".
func (p *FSPath) Parents(num int) string {
	if num == 0 || p.AbsPath == "" {
		return p.AbsPath
	}

	parts := p.Parts()

	if len(parts) == 0 {
		return ""
	}

	isAbsolute := strings.HasPrefix(p.AbsPath, "/")

	if isAbsolute {
		if len(parts) <= num+1 {
			return "/"
		}

		return "/" + strings.Join(parts[1:len(parts)-num], "/")
	} else {
		if len(parts) <= num {
			return ""
		}

		return strings.Join(parts[:len(parts)-num], "/")
	}
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
func (p *FSPath) Parent() string {
	return p.Parents(1)
}

// MkParentDir creates the parent directory for the given path
func (p *FSPath) MkParentDir() error {
	return p.fs.MkdirAll(filepath.Dir(p.AbsPath), DirMode755)
}

// MkDirs quick create dir for given path.
func (p *FSPath) MkDirs() error {
	return p.fs.MkdirAll(p.AbsPath, DirMode755)
}

// SplitPath splits the given path immediately following the final Separator,
// separating it into a directory and file name component.
//
// Parameters:
//   - pathStr: The path string to split.
//
// Returns:
//   - dir: The directory portion of the path.
//   - name: The file name portion of the path.
//
// This method uses filepath.Split internally and works with both file and directory paths.
// It's useful for separating a path into its directory and file (or final directory) components.
//
// Behavior:
//   - For file paths, it separates the directory and file name.
//   - For directory paths ending with a separator, the name component will be empty.
//   - The dir component always ends with a separator.
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
// Note:
//   - This method does not check if the path actually exists in the file system.
//   - It works purely on string manipulation and doesn't resolve symlinks or relative paths.
//   - The behavior is consistent across different operating systems, but the separator
//     character may vary (e.g., "/" on Unix, "\" on Windows).
func (p *FSPath) SplitPath(pathStr string) (dir, name string) {
	return filepath.Split(pathStr)
}

// GenRelativeFSPath generates a new FSPath instance relative to the current working directory.
//
// Parameters:
//   - name: A string representing the relative path and filename.
//
// Returns:
//   - *FSPath: A new FSPath instance representing the generated path.
//
// This method creates a new FSPath based on the current working directory and the provided relative path.
// It joins the current working directory with the provided name and creates a new FSPath from the result.
//
// Behavior:
//   - Generates a new absolute path based on the current working directory and the provided name.
//   - Creates and returns a new FSPath instance for the generated path.
//   - Resolves relative path components (like ".." or ".") in the provided name.
//   - Does not create any files or directories; it only generates the FSPath object.
//
// Examples:
//
//	Assume current FSPath represents "/tmp/a/b/current.txt":
//
//	1. newPath := currentPath.GenRelativeFSPath("new.txt")
//	   // newPath represents "/tmp/a/b/new.txt"
//
//	2. newPath := currentPath.GenRelativeFSPath("../new.txt")
//	   // newPath represents "/tmp/a/new.txt"
//
//	3. newPath := currentPath.GenRelativeFSPath("subdir/new.txt")
//	   // newPath represents "/tmp/a/b/subdir/new.txt"
//
//	4. newPath := currentPath.GenRelativeFSPath("../../other/new.txt")
//	   // newPath represents "/tmp/other/new.txt"
//
// Note:
//   - This method does not check if the generated path exists in the file system.
//   - It's useful for creating FSPath objects for potential new files or directories.
//   - The generated FSPath is a completely new instance and doesn't inherit properties from the original path.
//
// Use cases:
//   - Generating FSPath objects for related files or directories.
//   - Preparing paths for subsequent file operations.
//   - Creating paths relative to the current working directory, including parent directories.
func (p *FSPath) GenRelativeFSPath(name string) *FSPath {
	return Path(path.Join(p.WorkingDir, name))
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

	return Path(filepath.Join(newFolder, p.Name))
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
	name := p.Stem + "." + suffix
	pwd := p.WorkingDir

	if createTypeFolder {
		dir, last := filepath.Split(p.WorkingDir)

		if dir != "/" {
			name = fmt.Sprintf("%s_%s/%s", last, suffix, name)
			pwd = dir
		} else {
			name = fmt.Sprintf("_%s/%s", suffix, name)
		}
	}

	return Path(path.Join(pwd, name))
}

// CreateSiblingDir creates a new directory as a sibling to the current path.
//
// Parameters:
//   - name: The name of the new directory to create.
//
// Returns:
//   - *FSPath: A new FSPath instance representing the created directory.
//   - error: An error if the directory creation fails, nil otherwise.
//
// This method creates a new directory in the same directory as the current path,
// regardless of whether the current path is a file or a directory.
//
// Behavior:
//   - Creates the new directory as a sibling to the current path.
//   - If the directory already exists, it returns its FSPath without an error.
//   - Creates all necessary parent directories if they don't exist.
//
// Examples:
//
//  1. Current path is a file "/tmp/a/b/current.txt":
//     newDir, err := currentPath.CreateSiblingDir("newFolder")
//     // newDir represents "/tmp/a/b/newFolder/"
//
//  2. Current path is a directory "/tmp/a/b/":
//     newDir, err := currentPath.CreateSiblingDir("newFolder")
//     // newDir represents "/tmp/a/b/newFolder/"
//
// Note:
//   - This method interacts with the file system and may fail due to permission issues or other I/O errors.
//   - The returned FSPath represents a directory, so its IsDir() method will return true.
//   - If a file (not a directory) with the same name already exists, this operation will fail.
func (p *FSPath) CreateSiblingDir(name string) (*FSPath, error) {
	// Use GenRelativeFSPath to create a new FSPath for the new directory
	newDirPath := p.GenRelativeFSPath(name)

	// Use MkDirs method to create the directory
	err := newDirPath.MkDirs()
	if err != nil {
		return nil, fmt.Errorf("failed to create directory %s: %w", newDirPath.AbsPath, err)
	}

	return newDirPath, nil
}

// CreateSiblingDirToParent creates a new directory as a sibling to the parent directory of the current path.
//
// Parameters:
//   - name: The name of the new directory to create.
//
// Returns:
//   - *FSPath: A new FSPath instance representing the created directory.
//   - error: An error if the directory creation fails, nil otherwise.
//
// This method creates a new directory in the same directory as the parent of the current path,
// regardless of whether the current path is a file or a directory.
//
// Behavior:
//   - Creates the new directory as a sibling to the parent of the current path.
//   - If the directory already exists, it returns its FSPath without an error.
//   - Creates all necessary parent directories if they don't exist.
//
// Examples:
//
//  1. Current path is a file "/tmp/a/b/current.txt":
//     newDir, err := currentPath.CreateSiblingDirToParent("newFolder")
//     // newDir represents "/tmp/a/newFolder/"
//
//  2. Current path is a directory "/tmp/a/b/":
//     newDir, err := currentPath.CreateSiblingDirToParent("newFolder")
//     // newDir represents "/tmp/a/newFolder/"
//
// Note:
//   - This method interacts with the file system and may fail due to permission issues or other I/O errors.
//   - If a file (not a directory) with the same name already exists, this operation will fail.
//   - If the current path is at the root directory or one level below, this operation will fail.
func (p *FSPath) CreateSiblingDirToParent(name string) (*FSPath, error) {
	// Get the parent of the parent directory
	grandParentDir := p.Parents(2)
	if grandParentDir == "" || grandParentDir == "/" {
		return nil, ErrCannotCreateSiblingDir
	}

	// Use CreateSiblingDir on the grandparent directory
	return Path(grandParentDir).CreateSiblingDir(name)
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

	return afero.Glob(p.fs, filepath.Join(p.WorkingDir, pattern))
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
