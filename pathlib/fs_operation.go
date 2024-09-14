package pathlib

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	ErrNotDirectory      = errors.New("path is not a directory")
	ErrDirectoryNotEmpty = errors.New("directory not empty")
	ErrCannotUnlinkDir   = errors.New("cannot unlink directory: use Rmdir() instead")
)

// Copy creates a copy of the file at the current path to a new location.
//
// This method copies the content of the current file to a new file at the specified path.
// It also attempts to preserve the file mode (permissions) of the original file.
//
// Parameters:
//   - newfile: A string representing the path where the new copy should be created.
//
// Returns:
//   - error: An error if the operation failed. Possible reasons for failure include:
//   - The source file cannot be opened for reading.
//   - The destination file cannot be created or opened for writing.
//   - An error occurs during the copy operation.
//   - The file mode (permissions) cannot be set on the new file.
//
// The method performs the following steps:
// 1. Opens the source file for reading.
// 2. Creates the destination file.
// 3. Copies the content from the source to the destination.
// 4. Attempts to set the file mode of the new file to match the original.
//
// Example:
//
//	sourcePath := Path("/path/to/source/file.txt")
//	err := sourcePath.Copy("/path/to/destination/file_copy.txt")
//	if err != nil {
//	    log.Fatalf("Failed to copy file: %v", err)
//	}
//
// Note: This method does not handle copying directories. It's designed for single file operations.
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

// Mkdirs quick create dir for given path with MkdirAll.
func (p *FsPath) Mkdirs() error {
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

// Mkdir creates a new directory with the specified permissions.
// If the directory already exists, Mkdir does nothing and returns nil.
//
// Parameters:
//   - perm: The file mode bits to use for the new directory.
//   - parents: If true, any missing parents of this path are created as needed;
//     if false and the parent directory does not exist, an error is returned.
//
// Returns:
//   - error: An error if the operation failed.
//
// Examples:
//
//	// Create a single directory
//	err := path.Mkdir(0755, false)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Create a directory and all necessary parents
//	err := path.Mkdir(0755, true)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// Note:
//   - When parents is true, this method behaves like MkdirAll in the os package.
//   - When parents is false, it behaves like Mkdir in the os package.
//   - Using parents=true is generally safer and more convenient in most situations.
func (p *FsPath) Mkdir(perm os.FileMode, parents bool) error {
	if parents {
		return p.fs.MkdirAll(p.absPath, perm)
	}

	return p.fs.Mkdir(p.absPath, perm)
}

func (p *FsPath) MkdirAll(perm os.FileMode) error {
	return p.fs.MkdirAll(p.absPath, perm)
}

// Touch creates a new file or updates the modification time of an existing file.
// If the file doesn't exist, it is created with mode 0666 (before umask).
// If the file exists, its modification time is updated to the current time.
//
// Returns:
//   - error: An error if the operation failed.
//
// Example:
//
//	err := path.Touch()
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *FsPath) Touch() error {
	file, err := p.fs.OpenFile(p.absPath, os.O_CREATE|os.O_WRONLY, _mode666)
	if err != nil {
		return err
	}

	defer file.Close()

	currentTime := time.Now().Local()

	return p.fs.Chtimes(p.absPath, currentTime, currentTime)
}

// Chmod changes the mode of the file to the given mode.
//
// Parameters:
//   - mode: The new file mode.
//
// Returns:
//   - error: An error if the operation failed.
//
// Example:
//
//	err := path.Chmod(0644)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (p *FsPath) Chmod(mode os.FileMode) error {
	return p.fs.Chmod(p.absPath, mode)
}

// Unlink removes the file or symbolic link pointed to by the path.
// If the path points to a directory, an error is returned.
//
// Parameters:
//   - missingOK: If true, no error is returned if the file does not exist.
//
// Returns:
//   - error: nil if the operation was successful, otherwise an error explaining the failure.
//
// Possible errors:
//   - If the path does not exist and missingOK is false, it returns an os.ErrNotExist error.
//   - If the path points to a directory, it returns an error wrapping ErrCannotUnlinkDir.
//   - Other errors may be returned depending on the underlying filesystem operation.
//
// Example:
//
//	err := path.Unlink(false)
//	if err != nil {
//	    if errors.Is(err, ErrCannotUnlinkDir) {
//	        log.Println("Cannot unlink a directory")
//	    } else if os.IsNotExist(err) {
//	        log.Println("File does not exist")
//	    } else {
//	        log.Printf("Failed to unlink: %v", err)
//	    }
//	}
//
// Note: This method does not recursively remove directories. Use Rmdir() for removing empty directories.
func (p *FsPath) Unlink(missingOK bool) error {
	fileInfo, err := p.Stat()
	if err != nil {
		if os.IsNotExist(err) && missingOK {
			return nil
		}

		return err
	}

	if fileInfo.IsDir() {
		return fmt.Errorf("%w: %s", ErrCannotUnlinkDir, p.absPath)
	}

	return p.fs.Remove(p.absPath)
}

// Rmdir removes the empty directory pointed to by the path.
// If the directory is not empty, or if the path points to a file or a symbolic link, an error is returned.
//
// Returns:
//   - error: nil if the operation was successful, otherwise an error explaining the failure.
//
// Possible errors:
//   - If the path does not exist, it returns an os.ErrNotExist error.
//   - If the path is not a directory, it returns an error wrapping ErrNotDirectory.
//   - If the directory is not empty, it returns an error wrapping ErrDirectoryNotEmpty.
//   - Other errors may be returned depending on the underlying filesystem operation.
//
// Example:
//
//	err := path.Rmdir()
//	if err != nil {
//	    switch {
//	    case errors.Is(err, ErrNotDirectory):
//	        log.Println("Path is not a directory")
//	    case errors.Is(err, ErrDirectoryNotEmpty):
//	        log.Println("Directory is not empty")
//	    case os.IsNotExist(err):
//	        log.Println("Directory does not exist")
//	    default:
//	        log.Printf("Failed to remove directory: %v", err)
//	    }
//	}
//
// Note: This method only removes empty directories. It does not recursively remove directory contents.
func (p *FsPath) Rmdir() error {
	fileInfo, err := p.Stat()
	if err != nil {
		return err
	}

	if !fileInfo.IsDir() {
		return fmt.Errorf("%w: %s", ErrNotDirectory, p.absPath)
	}

	// Check if the directory is empty
	dir, err := p.fs.Open(p.absPath)
	if err != nil {
		return err
	}

	defer dir.Close()

	_, err = dir.Readdirnames(1) // Try to read one entry
	if err == nil {
		return fmt.Errorf("%w: %s", ErrDirectoryNotEmpty, p.absPath)
	}

	if !errors.Is(err, io.EOF) {
		return err // Some other error occurred
	}

	// Directory is empty, remove it
	return p.fs.Remove(p.absPath)
}
