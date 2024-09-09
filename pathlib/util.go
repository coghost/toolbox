package pathlib

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

var userLookup = os.UserHomeDir // This can be overridden in tests

// ResolveAbsPath takes a file path and returns its absolute path.
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
//	absPath := ResolveAbsPath("~/documents/file.txt")
//	// On a Unix system, this might return "/home/user/documents/file.txt"
//
// Note: This function does not check if the path actually exists in the file system.
func ResolveAbsPath(filePath string) (string, error) {
	// Expand the path first
	expandedPath := Expand(filePath)

	// Convert to absolute path without resolving symlinks
	absPath, err := filepath.Abs(expandedPath)
	if err != nil {
		// If we can't get the absolute path, use the expanded path
		return "", err
	}

	return absPath, nil
}

// Home returns the home directory of the current user.
func Home() (*FsPath, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	return Path(home), nil
}

// Cwd returns a new FsPath representing the current working directory.
func Cwd() (*FsPath, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	return Path(dir), nil
}

// Expand takes a file path and expands any environment variables and user home directory references.
//
// This function performs the following expansions:
//  1. Expands the user's home directory (e.g., "~" or "~username")
//  2. Expands environment variables (e.g., "$HOME" or "${HOME}")
//
// The function first expands the user's home directory using ExpandUser,
// then expands any environment variables using os.ExpandEnv.
//
// Parameters:
//   - path: A string representing the file path to be expanded.
//
// Returns:
//   - string: The expanded path.
//
// Example:
//
//	expanded := Expand("~/documents/$USER/file.txt")
//	// This might return "/home/username/documents/username/file.txt"
//
// Note: This function does not check if the expanded path actually exists in the file system.
func Expand(path string) string {
	// Return immediately if path is empty
	if path == "" {
		return path
	}

	// Handle home directory expansion first
	expandedPath := ExpandUser(path)

	// Protect escaped dollar signs
	expandedPath = strings.ReplaceAll(expandedPath, "\\$", "\u0001")
	// Expand environment variables
	expandedPath = os.ExpandEnv(expandedPath)
	// Restore protected dollar signs
	expandedPath = strings.ReplaceAll(expandedPath, "\u0001", "$")

	return expandedPath
}

// ExpandUser replaces a leading ~ or ~user with the user's home directory.
//
// This function handles the following cases:
//  1. "~" or "~/..." expands to the current user's home directory
//  2. "~username" or "~username/..." expands to the specified user's home directory
//
// If the user's home directory cannot be determined, the original path is returned.
//
// Parameters:
//   - path: A string representing the file path to be expanded.
//
// Returns:
//   - string: The expanded path with the user's home directory.
//
// Example:
//
//	expanded := ExpandUser("~/documents")
//	// This might return "/home/username/documents"
//
//	expanded := ExpandUser("~otheruser/documents")
//	// This might return "/home/otheruser/documents"
//
// Note: This function does not expand environment variables. Use Expand for full expansion.
func ExpandUser(path string) string {
	if !strings.HasPrefix(path, "~") {
		return path
	}

	var (
		homeDir string
		err     error
	)

	if path == "~" || strings.HasPrefix(path, "~/") {
		homeDir, err = userLookup()
		if err != nil {
			return path
		}

		if path == "~" {
			return homeDir
		}

		return filepath.Join(homeDir, path[2:])
	}

	// Handle ~user case
	parts := strings.SplitN(path[1:], "/", 2)
	username := parts[0]
	restPath := ""

	if len(parts) > 1 {
		restPath = parts[1]
	}

	homeDir, err = getUserHomeDir(username)
	if err != nil {
		return path
	}

	return filepath.Join(homeDir, restPath)
}

func getUserHomeDir(username string) (string, error) {
	u, err := user.Lookup(username)
	if err != nil {
		return filepath.Join("/home", username), err
	}

	return u.HomeDir, nil
}
