package pathlib

import (
	"path/filepath"

	"github.com/spf13/afero"
)

// ListFilesWithGlob lists files in the working directory matching the given pattern
//
// This method performs a glob operation in the directory of the current FSPath,
// using the pattern provided. It leverages the underlying file system associated
// with this FSPath instance.
//
// Parameters:
//   - pattern: The glob pattern to match files against. If empty, defaults to "*".
//
// Returns:
//   - []string: A slice of matched file paths.
//   - error: An error if the glob operation fails.
//
// The method uses the directory of the current FSPath as the root for the glob operation.
// If the pattern is an empty string, it defaults to "*", matching all files in the directory.
//
// Example usage:
//
//	path := Path("/home/user/documents")
//	files, err := path.ListFilesWithGlob("*.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, file := range files {
//	    fmt.Println(file)
//	}
//
// Note: This method does not recurse into subdirectories unless specified in the pattern.
// The returned paths are relative to the working directory of the FSPath.
func (p *FsPath) ListFilesWithGlob(pattern string) ([]string, error) {
	return ListFilesWithGlob(p.fs, p.Dir().absPath, pattern)
}

// ListFilesWithGlob lists files in the specified directory matching the given pattern.
//
// This function uses the provided file system (fs) to perform the glob operation.
// If fs is nil, it defaults to the OS file system.
//
// Parameters:
//   - fs: The file system to use. If nil, uses the OS file system.
//   - rootDir: The root directory in which to perform the glob operation.
//   - pattern: The glob pattern to match files against. If empty, defaults to "*".
//
// Returns:
//   - []string: A slice of matched file paths.
//   - error: An error if the glob operation fails.
//
// The function expands the rootDir to handle home directory references and environment variables.
// It then performs a glob operation using the specified pattern in the given root directory.
//
// If the pattern is an empty string, it defaults to "*", matching all files in the root directory.
//
// Example usage:
//
//	// List all .txt files in the user's home directory
//	files, err := ListFilesWithGlob(nil, "~/Documents", "*.txt")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	for _, file := range files {
//	    fmt.Println(file)
//	}
//
// Note: This function does not recurse into subdirectories unless specified in the pattern.
func ListFilesWithGlob(fs afero.Fs, rootDir, pattern string) ([]string, error) {
	if pattern == "" {
		pattern = "*"
	}

	if fs == nil {
		fs = afero.NewOsFs()
	}

	return afero.Glob(fs, filepath.Join(Expand(rootDir), pattern))
}
