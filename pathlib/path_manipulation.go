package pathlib

import (
	"os"
	"path/filepath"
	"strings"
)

// Home returns a new FsPath representing the user's home directory.
//
// This method is equivalent to calling the Home() function.
//
// Returns:
//   - *FsPath: A new FsPath instance representing the user's home directory.
//   - error: An error if the home directory couldn't be determined.
//
// Example:
//
//	homePath, err := path.Home()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(homePath.absPath)
func (p *FsPath) Home() (*FsPath, error) {
	return Home()
}

// Cwd returns a new FsPath representing the current working directory.
//
// This method is equivalent to calling the Cwd() function.
//
// Returns:
//   - *FsPath: A new FsPath instance representing the current working directory.
//   - error: An error if the current working directory couldn't be determined.
//
// Example:
//
//	cwdPath, err := path.Cwd()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(cwdPath.absPath)
func (p *FsPath) Cwd() (*FsPath, error) {
	return Cwd()
}

// ExpandUser replaces a leading ~ or ~user with the user's home directory.
//
// This method is equivalent to calling the ExpandUser() function on the path.
//
// Returns:
//   - A new FsPath with the expanded path.
func (p *FsPath) ExpandUser() *FsPath {
	expandedPath := ExpandUser(p.absPath)
	return Path(expandedPath)
}

// Expand expands environment variables and user's home directory in the path.
//
// This method is equivalent to calling the Expand() function on the path.
//
// Returns:
//   - A new FsPath with the expanded path.
func (p *FsPath) Expand() *FsPath {
	expandedPath := Expand(p.absPath)
	return Path(expandedPath)
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
func (p *FsPath) BaseDir() string {
	return filepath.Base(p.Dir().absPath)
}

func (p *FsPath) Dir() *FsPath {
	if p.IsDir() {
		return p // If it's already a directory, return itself
	}

	// For files, return the parent directory
	return Path(filepath.Dir(p.absPath))
}

// Join joins one or more path components to the current path.
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
//     path.Join("documents", "file.txt") returns a new FSPath for "/home/user/documents/file.txt"
//
//  2. For a file "/home/user/file.txt":
//     path.Join("documents", "newfile.txt") returns a new FSPath for "/home/user/documents/newfile.txt"
//
//  3. For a path "/":
//     path.Join("etc", "config") returns a new FSPath for "/etc/config"
//
// Note: This method does not modify the original FSPath instance or create any directories.
// It only returns a new FSPath instance representing the joined path.
func (p *FsPath) Join(others ...string) *FsPath {
	if len(others) > 0 && filepath.IsAbs(others[0]) {
		// If the first component is an absolute path, use it as the base
		return Path(filepath.Join(others...))
	}

	components := append([]string{p.absPath}, others...)

	return Path(filepath.Join(components...))
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
func (p *FsPath) Parent() *FsPath {
	if p.absPath == "/" {
		return p // Root directory is its own parent
	}

	parentPath := filepath.Dir(p.absPath)

	return Path(parentPath)
}

// Parents returns an iterator of this path's logical parents.
//
// Returns:
//   - A slice of FsPath instances representing all the logical parents of the path.
//
// The first parent is the immediate parent of the path, and the last parent is the root path.
func (p *FsPath) Parents() []*FsPath {
	var parents []*FsPath

	// If the current path is ".", return an empty slice
	if p.RawPath == "." {
		return parents
	}

	current := p

	cwd, err := os.Getwd()
	if err != nil {
		// If we can't get the current working directory, just return an empty slice
		return parents
	}

	for {
		parent := current.Parent()
		if parent.absPath == current.absPath || parent.absPath == cwd {
			// We've reached the root directory or the current working directory
			break
		}

		parents = append(parents, parent)
		current = parent
	}

	return parents
}

// ParentsUpTo returns the parent directory path up to the specified number of levels.
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
//     - ParentsUpTo(1) returns FSPath("/home/user/documents")
//     - ParentsUpTo(2) returns FSPath("/home/user")
//     - ParentsUpTo(3) returns FSPath("/home")
//     - ParentsUpTo(4) or higher returns FSPath("/")
//
// Note:
//   - If num is 0, it returns the current path.
//   - If num is greater than or equal to the number of directories in the path,
//     it returns the root directory for absolute paths or "." for relative paths.
func (p *FsPath) ParentsUpTo(num int) *FsPath {
	if num <= 0 {
		return p
	}

	current := p
	for i := 0; i < num; i++ {
		parent := current.Parent()
		if parent.absPath == current.absPath {
			// We've reached the root directory or "."
			return parent
		}

		current = parent
	}

	return current
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
func (p *FsPath) Parts() []string {
	// If the path is empty, return an empty slice
	if p.absPath == "" {
		return []string{}
	}

	// Trim trailing slashes
	trimmed := strings.TrimRight(p.absPath, "/")

	// If the path was just "/", return a slice with only "/"
	if trimmed == "" {
		return []string{"/"}
	}

	// Split the path
	parts := strings.Split(trimmed, "/")

	// If the original path started with a slash, add it as the first element
	if strings.HasPrefix(p.absPath, "/") {
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

// RelativeTo returns a relative path to p from the given path.
func (p *FsPath) RelativeTo(other string) (string, error) {
	otherAbs, err := ResolveAbsPath(other)
	if err != nil {
		return "", err
	}

	return filepath.Rel(otherAbs, p.absPath)
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
func (p *FsPath) Split(pathStr string) (dir, name string) {
	return filepath.Split(pathStr)
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
func (p *FsPath) WithName(name string) *FsPath {
	return p.Parent().Join(name)
}

func (p *FsPath) WithStem(stem string) *FsPath {
	newName := stem + p.Suffix

	return p.Parent().Join(newName)
}

func (p *FsPath) WithSuffix(suffix string) *FsPath {
	if suffix != "" && !strings.HasPrefix(suffix, ".") {
		suffix = "." + suffix
	}

	return Path(strings.TrimSuffix(p.absPath, p.Suffix) + suffix)
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
func (p *FsPath) WithRenamedParentDir(newParentName string) *FsPath {
	// If the current path is the root directory, return the original FSPath
	if p.absPath == "/" {
		return p
	}

	// Get the parent directory
	parentDir := p.Parent()

	// Create a new FSPath for the new directory within the parent
	newDir := parentDir.WithName(newParentName)

	// Join the new directory with the current file name
	return newDir.Join(p.Name)
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
func (p *FsPath) WithSuffixAndSuffixedParentDir(newSuffix string) *FsPath {
	// If it's a directory, return the original path
	if p.IsDir() {
		return nil
	}

	// Ensure the new suffix starts with a dot
	if newSuffix != "" && !strings.HasPrefix(newSuffix, ".") {
		newSuffix = "." + newSuffix
	}

	// Change the file suffix
	newPath := p.WithSuffix(newSuffix)

	// Remove the dot from the suffix for the directory name
	dirSuffix := strings.TrimPrefix(newSuffix, ".")

	// Handle root directory case
	if p.Parent().absPath == "/" {
		return p.Parent().Join("_"+dirSuffix, newPath.Name)
	}

	// Create new directory name and rename the parent directory
	newDirName := p.Parent().Name + "_" + dirSuffix

	// Use WithRenamedParentDir to create the new path
	return newPath.WithRenamedParentDir(newDirName)
}

func (p *FsPath) WithReplacedDirAndSuffix(dirName, newSuffix string) *FsPath {
	// If it's a directory, return the original path
	if p.IsDir() {
		return nil
	}

	// Ensure the new suffix starts with a dot
	if newSuffix != "" && !strings.HasPrefix(newSuffix, ".") {
		newSuffix = "." + newSuffix
	}

	// Change the file suffix
	newPath := p.WithSuffix(newSuffix)

	// Handle root directory case
	if p.Parent().absPath == "/" {
		return p.Parent().Join(dirName, newPath.Name)
	}

	// Use WithRenamedParentDir to create the new path with the new directory name
	return newPath.WithRenamedParentDir(dirName)
}

// LastNSegments returns the last n segments of the path.
//
//   - If n is 0, it returns just the file name.
//   - If n is greater than or equal to the number of segments, it returns the full path.
func (p *FsPath) LastNSegments(num int) string {
	if p.RawPath == "" {
		return p.RawPath
	}

	if num <= 0 {
		return p.Name
	}

	// Split the path into segments
	segments := strings.Split(p.RawPath, string(os.PathSeparator))

	if num >= len(segments) {
		return p.RawPath
	}

	// Use ParentsUpTo to get the parent directory n levels up
	parentPath := p.ParentsUpTo(num)

	// If parentPath is the same as p, it means we've reached the root or n is greater than the number of segments
	if parentPath.RawPath == p.RawPath {
		return p.RawPath
	}

	// Get the relative path from parentPath to p
	relPath, err := p.RelativeTo(parentPath.RawPath)
	if err != nil {
		// If there's an error, fall back to returning the full path
		return p.RawPath
	}

	// Join the relative path with the last segment of parentPath
	return filepath.Join(filepath.Base(parentPath.RawPath), relPath)
}

// LastSegment returns the last segment of the path.
//
//	This is equivalent to calling LastNSegments(1).
func (p *FsPath) LastSegment() string {
	return p.LastNSegments(1)
}
