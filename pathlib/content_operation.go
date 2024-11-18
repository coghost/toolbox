package pathlib

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/afero"
)

const (
	SepRuneCsv = ','
	SepRuneTsv = '\t'
)

func (p *FsPath) MustGetBytes() []byte {
	b, err := p.GetBytes()
	p.e(err)

	return b
}

func (p *FsPath) GetBytes() ([]byte, error) {
	return afero.ReadFile(p.fs, p.absPath)
}

func (p *FsPath) SetBytes(data []byte) error {
	if err := p.MkParentDir(); err != nil {
		return err
	}

	return afero.WriteFile(p.fs, p.absPath, data, FileMode644)
}

// MustSetString sets the file content as a string, panics on error
func (p *FsPath) MustSetString(data string) {
	p.e(p.SetString(data))
}

// SetString sets the file content as a string
func (p *FsPath) SetString(data string) error {
	return p.SetBytes([]byte(data))
}

func (p *FsPath) MustGetString() string {
	b, err := p.GetString()
	p.e(err)

	return b
}

func (p *FsPath) GetString() (string, error) {
	data, err := p.GetBytes()
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// GetLines reads the file and returns its contents as a slice of strings.
//
// Each string in the returned slice represents a line in the file.
// The newline characters are stripped from the end of each line.
//
// Returns:
//   - []string: A slice containing each line of the file.
//   - error: An error if the file cannot be read or processed.
//
// Example usage:
//
//	lines, err := path.GetLines()
//	if err != nil {
//		// handle error
//	}
//	for _, line := range lines {
//		fmt.Println(line)
//	}
//
// Note: This method reads the entire file into memory. For very large files,
// consider using a streaming approach instead.
func (p *FsPath) GetLines() ([]string, error) {
	content, err := p.GetString()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(content, "\n")

	// Remove any trailing empty lines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	// Trim any trailing \r from each line (for Windows-style line endings)
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, "\r")
	}

	return lines, nil
}

// MustGetLines reads the file and returns its contents as a slice of strings.
// It panics if an error occurs.
//
// This is a convenience wrapper around GetLines.
//
// Returns:
//   - []string: A slice containing each line of the file.
//
// Example usage:
//
//	lines := path.MustGetLines()
//	for _, line := range lines {
//		fmt.Println(line)
//	}
//
// Note: Use this method only when you're sure the file exists and can be read,
// or if you want to halt execution on error.
func (p *FsPath) MustGetLines() []string {
	lines, err := p.GetLines()
	p.e(err)

	return lines
}

// GetJSON reads the file and unmarshals its content into the provided interface.
//
// This method reads the entire file content using GetBytes, then uses json.Unmarshal
// to parse the JSON data into the provided interface.
//
// Parameters:
//   - v: A pointer to the variable where the unmarshaled data should be stored.
//
// Returns:
//   - error: An error if the file cannot be read or if the JSON unmarshaling fails.
//
// Example usage:
//
//	type Config struct {
//	    Name    string `json:"name"`
//	    Version int    `json:"version"`
//	}
//
//	var config Config
//	err := path.GetJSON(&config)
//	if err != nil {
//	    // handle error
//	}
//	fmt.Printf("Name: %s, Version: %d\n", config.Name, config.Version)
//
// Note: This method reads the entire file into memory. For very large files,
// consider using a streaming JSON parser instead.
func (p *FsPath) GetJSON(raw interface{}) error {
	data, err := p.GetBytes()
	if err != nil {
		return err
	}

	return json.Unmarshal(data, raw)
}

// MustGetJSON reads the file, unmarshals its content into the provided interface, and panics on error.
//
// This method is similar to GetJSON but panics if an error occurs during reading or unmarshaling.
//
// Parameters:
//   - v: A pointer to the variable where the unmarshaled data should be stored.
//
// Example usage:
//
//	type Config struct {
//	    Name    string `json:"name"`
//	    Version int    `json:"version"`
//	}
//
//	var config Config
//	path.MustGetJSON(&config)
//	fmt.Printf("Name: %s, Version: %d\n", config.Name, config.Version)
//
// Note: Use this method only when you're sure the file exists and contains valid JSON,
// or if you want to halt execution on error.
func (p *FsPath) MustGetJSON(v interface{}) {
	p.e(p.GetJSON(v))
}

// Reader returns an io.Reader for the file
func (p *FsPath) Reader() (io.Reader, error) {
	return p.fs.Open(p.absPath)
}

// CSVGetSlices reads the CSV file and returns its content as slices.
//
// This method reads the file at the FsPath's location as a CSV (Comma-Separated Values) file
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
func (p *FsPath) CSVGetSlices() ([][]string, error) {
	return p.readDelimitedFile(SepRuneCsv)
}

func (p *FsPath) MustCSVGetSlices() [][]string {
	arr, err := p.CSVGetSlices()
	p.e(err)

	return arr
}

// TSVGetSlices reads the TSV file and returns its content as slices.
//
// This method reads the file at the FsPath's location as a TSV (Tab-Separated Values) file
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
func (p *FsPath) TSVGetSlices() ([][]string, error) {
	return p.readDelimitedFile(SepRuneTsv)
}

func (p *FsPath) MustTSVGetSlices() [][]string {
	arr, err := p.TSVGetSlices()
	p.e(err)

	return arr
}

func (p *FsPath) readDelimitedFile(separator rune) ([][]string, error) {
	reader, err := p.Reader()
	if err != nil {
		return nil, err
	}

	return ToSlices(reader, separator)
}

// WriteText writes the given string data to the file, creating the file if it doesn't exist,
// and overwriting it if it does.
func (p *FsPath) WriteText(data string) error {
	return p.SetString(data)
}

// WriteBytes writes the given byte slice to the file, creating the file if it doesn't exist,
// and overwriting it if it does.
func (p *FsPath) WriteBytes(data []byte) error {
	return p.SetBytes(data)
}

// ReadText reads the contents of the file and returns it as a string.
func (p *FsPath) ReadText() (string, error) {
	return p.GetString()
}

// ReadBytes reads the contents of the file and returns it as a byte slice.
func (p *FsPath) ReadBytes() ([]byte, error) {
	return p.GetBytes()
}

// MustWriteText writes the given string data to the file, creating the file if it doesn't exist,
// and overwriting it if it does. It panics on error.
func (p *FsPath) MustWriteText(data string) {
	p.e(p.WriteText(data))
}

// MustWriteBytes writes the given byte slice to the file, creating the file if it doesn't exist,
// and overwriting it if it does. It panics on error.
func (p *FsPath) MustWriteBytes(data []byte) {
	p.e(p.WriteBytes(data))
}

// MustReadText reads the contents of the file and returns it as a string. It panics on error.
func (p *FsPath) MustReadText() string {
	text, err := p.ReadText()
	p.e(err)

	return text
}

// MustReadBytes reads the contents of the file and returns it as a byte slice. It panics on error.
func (p *FsPath) MustReadBytes() []byte {
	bytes, err := p.ReadBytes()
	p.e(err)

	return bytes
}

// AppendText appends the given string data to the file, creating the file if it doesn't exist.
func (p *FsPath) AppendText(data string) error {
	return p.appendData([]byte(data))
}

// AppendBytes appends the given byte slice to the file, creating the file if it doesn't exist.
func (p *FsPath) AppendBytes(data []byte) error {
	return p.appendData(data)
}

// MustAppendText appends the given string data to the file, creating the file if it doesn't exist.
// It panics on error.
func (p *FsPath) MustAppendText(data string) {
	p.e(p.AppendText(data))
}

// MustAppendBytes appends the given byte slice to the file, creating the file if it doesn't exist.
// It panics on error.
func (p *FsPath) MustAppendBytes(data []byte) {
	p.e(p.AppendBytes(data))
}

// appendData is a helper function to append data to a file
func (p *FsPath) appendData(data []byte) error {
	if err := p.MkParentDir(); err != nil {
		return err
	}

	file, err := p.fs.OpenFile(p.absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, FileMode644)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)

	return err
}

// GetMD5 calculates the MD5 hash of the file at the given path.
func (p *FsPath) GetMD5() (string, error) {
	file, err := p.fs.Open(p.absPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file for MD5 calculation: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("failed to calculate MD5 hash: %w", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
