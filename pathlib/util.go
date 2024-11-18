package pathlib

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
)

var (
	_, b, _, _ = runtime.Caller(0)

	// rootDir folder of this project, you must config this in every project, this is just a demo usage.
	rootDir = filepath.Join(filepath.Dir(b), "../..")
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

// IsJSON checks if the given byte slice contains valid JSON data.
//
// This function attempts to unmarshal the input data into a json.RawMessage.
// If the unmarshal operation succeeds, it means the data is valid JSON.
//
// Parameters:
//   - data: A byte slice containing the data to be checked.
//
// Returns:
//   - bool: true if the data is valid JSON, false otherwise.
//
// The function returns true for valid JSON structures including objects,
// arrays, strings, numbers, booleans, and null. It returns false for any
// input that cannot be parsed as valid JSON.
//
// Example:
//
//	validJSON := []byte(`{"name": "John", "age": 30}`)
//	fmt.Println(IsJSON(validJSON))  // Output: true
//
//	invalidJSON := []byte(`{"name": "John", "age": }`)
//	fmt.Println(IsJSON(invalidJSON))  // Output: false
//
// Note: This function does not provide information about why the JSON might
// be invalid. For more detailed error information, use json.Unmarshal directly.
func IsJSON(data []byte) bool {
	var js json.RawMessage
	return json.Unmarshal(data, &js) == nil
}

// StructToJSONMap converts a struct to a map[string]interface{} representation of JSON.
// This function uses JSON marshaling and unmarshaling to perform the conversion,
// which means it respects JSON tags and only includes exported fields.
//
// Parameters:
//   - src: The source struct to convert.
//
// Returns:
//   - map[string]interface{}: A map representing the JSON structure of the input struct.
//   - error: An error if the conversion process fails.
//
// Example usage:
//
//	type Person struct {
//	    Name string `json:"name"`
//	    Age  int    `json:"age"`
//	}
//
//	person := Person{Name: "John Doe", Age: 30}
//	jsonMap, err := StructToJSONMap(person)
//	if err != nil {
//	    // Handle error
//	}
//	// jsonMap will be: map[string]interface{}{"name": "John Doe", "age": 30}
//
// Note: This function will only include fields that would be marshaled to JSON.
// Unexported fields and fields with `json:"-"` tags will be omitted.
func StructToJSONMap(src interface{}) (map[string]interface{}, error) {
	jsonData, err := json.Marshal(src)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal struct to JSON: %w", err)
	}

	// Then unmarshal the JSON to a map
	var result map[string]interface{}

	err = json.Unmarshal(jsonData, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON to map: %w", err)
	}

	return result, nil
}

// SanitizeJSON cleans a JSON string by removing empty fields.
// It unmarshals the raw JSON into the provided struct template,
// then marshals it back to JSON, effectively removing any empty or zero-value fields.
//
// Parameters:
//   - rawJSON: The input JSON string to sanitize.
//   - template: A pointer to a struct that defines the expected schema of the JSON.
//
// Returns:
//   - []byte: The sanitized JSON as a byte slice.
//   - error: Any error encountered during the process.
//
// Example usage:
//
//	type Person struct {
//	    Name string `json:"name,omitempty"`
//	    Age  int    `json:"age,omitempty"`
//	}
//
//	input := `{"name":"John","age":0,"extra":""}`
//	cleaned, err := SanitizeJSON(input, &Person{})
//	if err != nil {
//	    // Handle error
//	}
//	// cleaned will be: `{"name":"John"}`
//
// Note: This function relies on the `omitempty` tag for fields that should be
// removed when empty. Make sure your struct is properly tagged for the desired behavior.
func SanitizeJSON(rawJSON string, template interface{}) ([]byte, error) {
	// Unmarshal raw JSON into the provided struct template
	if err := json.Unmarshal([]byte(rawJSON), template); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	// Marshal the struct back to JSON, which will omit empty fields
	sanitized, err := json.Marshal(template)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal sanitized struct: %w", err)
	}

	return sanitized, nil
}

func ToSlices(reader io.Reader, separator rune) ([][]string, error) {
	r := csv.NewReader(reader)
	r.Comma = separator
	r.Comment = '#'

	return r.ReadAll()
}
