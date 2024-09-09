package pathlib

import (
	"encoding/csv"
	"io"
	"os"

	"github.com/spf13/afero"
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

func (p *FsPath) GetString() (string, error) {
	data, err := p.GetBytes()
	if err != nil {
		return "", err
	}

	return string(data), nil
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
	return p.readDelimitedFile(',')
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
	return p.readDelimitedFile('\t')
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

	r := csv.NewReader(reader)
	r.Comma = separator
	r.Comment = '#'

	return r.ReadAll()
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
