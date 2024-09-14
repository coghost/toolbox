package pathlib

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/afero"
)

const (
	maxSize = 1 * 1024 * 1024 * 1024 // 1GB max size
)

var (
	ErrIllegalFilePath = errors.New("illegal file path")
	ErrFileTooLarge    = errors.New("file exceeded maximum allowed size")
	ErrIncompleteWrite = errors.New("incomplete write: consider increasing the maxSize parameter or checking for disk space issues")
)

// CompressOptions holds the options for compression and decompression operations
type CompressOptions struct {
	MaxSize int64
}

// defaultCompressOptions returns the default options for compression and decompression
func defaultCompressOptions() CompressOptions {
	return CompressOptions{
		MaxSize: maxSize,
	}
}

// CompressOption defines the method to modify CompressOptions
type CompressOption func(*CompressOptions)

// WithMaxSize sets the MaxSize option
func WithMaxSize(maxSize int64) CompressOption {
	return func(o *CompressOptions) {
		o.MaxSize = maxSize
	}
}

func applyCompressOptions(opts ...CompressOption) CompressOptions {
	options := defaultCompressOptions()
	for _, opt := range opts {
		opt(&options)
	}

	return options
}

// ZipDir compresses the directory represented by this FsPath into a zip file.
// The zip file is created in the same folder as the directory.
//
// Parameters:
//   - zipFileName: The name of the zip file to be created. If it doesn't end with ".zip",
//     the ".zip" extension will be automatically added.
//
// Returns:
//   - *FsPath: A new FsPath representing the created zip file.
//   - int: The total number of files compressed into the zip file.
//   - error: An error if any step of the compression process fails.
//
// The function performs the following steps:
//  1. Checks if the current FsPath represents a directory.
//  2. Prepares the compression by creating the zip file.
//  3. Creates a new zip writer.
//  4. Walks through the directory, adding each file to the zip archive.
//  5. Closes the zip writer and the file.
//
// Possible errors include:
//   - ErrNotDirectory if the current FsPath is not a directory.
//   - I/O errors during file creation or writing.
//   - Errors encountered while walking the directory structure.
//
// Example usage:
//
//	dirPath := Path("/path/to/directory")
//	zipPath, fileCount, err := dirPath.ZipDir("archive.zip")
//	if err != nil {
//	    log.Fatalf("Failed to create zip: %v", err)
//	}
//	fmt.Printf("Created zip at %s with %d files\n", zipPath, fileCount)
func (p *FsPath) ZipDir(zipFileName string) (*FsPath, int, error) {
	zipPath, zipFile, err := p.prepareCompression(zipFileName, ".zip")
	if err != nil {
		return nil, 0, err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	totalFiles, err := p.compressDirectoryToWriter(
		func(name string, info os.FileInfo) (io.Writer, error) {
			return zipWriter.Create(name)
		},
	)
	if err != nil {
		return nil, 0, err
	}

	return zipPath, totalFiles, nil
}

// TarGzDir compresses the directory represented by this FsPath into a tar.gz file.
// The tar.gz file is created in the same folder as the directory.
//
// Parameters:
//   - tarGzFileName: The name of the tar.gz file to be created. If it doesn't end with ".tar.gz",
//     the ".tar.gz" extension will be automatically added.
//
// Returns:
//   - *FsPath: A new FsPath representing the created tar.gz file.
//   - int: The total number of files compressed into the tar.gz file.
//   - error: An error if any step of the compression process fails.
//
// The function performs the following steps:
//  1. Checks if the current FsPath represents a directory.
//  2. Prepares the compression by creating the tar.gz file.
//  3. Creates a new gzip writer.
//  4. Creates a new tar writer.
//  5. Walks through the directory, adding each file to the tar archive.
//  6. Closes the tar writer, gzip writer, and the file.
//
// Possible errors include:
//   - ErrNotDirectory if the current FsPath is not a directory.
//   - I/O errors during file creation or writing.
//   - Errors encountered while walking the directory structure.
//   - Errors in creating tar headers or writing tar entries.
//
// Example usage:
//
//	dirPath := Path("/path/to/directory")
//	tarGzPath, fileCount, err := dirPath.TarGzDir("archive.tar.gz")
//	if err != nil {
//	    log.Fatalf("Failed to create tar.gz: %v", err)
//	}
//	fmt.Printf("Created tar.gz at %s with %d files\n", tarGzPath, fileCount)
//
// Note: This function compresses the entire directory structure, including subdirectories.
// Empty directories are included in the archive.
func (p *FsPath) TarGzDir(tarGzFileName string) (*FsPath, int, error) {
	tarGzPath, file, err := p.prepareCompression(tarGzFileName, ".tar.gz")
	if err != nil {
		return nil, 0, err
	}
	defer file.Close()

	gzipWriter := gzip.NewWriter(file)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	totalFiles, err := p.compressDirectoryToWriter(
		func(name string, info os.FileInfo) (io.Writer, error) {
			header, err := tar.FileInfoHeader(info, name)
			if err != nil {
				return nil, fmt.Errorf("failed to create tar header: %w", err)
			}

			header.Name = name

			if err := tarWriter.WriteHeader(header); err != nil {
				return nil, fmt.Errorf("failed to write tar header: %w", err)
			}

			return tarWriter, nil
		},
	)
	if err != nil {
		return nil, 0, err
	}

	return tarGzPath, totalFiles, nil
}

// Untar extracts the contents of the tar file to the specified destination directory.
func (p *FsPath) Untar(destDir string, opts ...CompressOption) error {
	options := applyCompressOptions(opts...)

	// Create the subdirectory for extraction
	subDir := Path(destDir).Join(strings.TrimSuffix(p.Name, ".tar.gz"))
	if err := subDir.MkdirAll(_mode755); err != nil {
		return fmt.Errorf("failed to create subdirectory: %w", err)
	}

	// Prepare the tar reader
	tarReader, cleanup, err := p.prepareUntarEnvironment()
	if err != nil {
		return err
	}
	defer cleanup()

	for {
		header, err := tarReader.Next()

		switch {
		case errors.Is(err, io.EOF):
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}

		target := subDir.Join(header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := target.MkdirAll(os.FileMode(header.Mode & _mode777)); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := target.MkParentDir(); err != nil {
				return err
			}

			if err := p.extractTarFile(target, header, tarReader, options.MaxSize); err != nil {
				return err
			}
		}
	}
}

// Unzip extracts the contents of the zip file to the specified destination directory.
func (p *FsPath) Unzip(destDir string, opts ...CompressOption) error {
	options := applyCompressOptions(opts...)

	subDir := Path(destDir).Join(strings.TrimSuffix(p.Name, ".zip"))
	if err := subDir.MkdirAll(_mode755); err != nil {
		return fmt.Errorf("failed to create subdirectory: %w", err)
	}

	reader, err := zip.OpenReader(p.absPath)
	if err != nil {
		return fmt.Errorf("failed to open zip file: %w", err)
	}
	defer reader.Close()

	for _, file := range reader.File {
		err := p.extractZipFile(file, subDir, options.MaxSize)
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *FsPath) extractTarFile(target *FsPath, header *tar.Header, tarReader *tar.Reader, maxSize int64) error {
	if header.Size > maxSize {
		return fmt.Errorf("%w: %s (size: %d bytes, max allowed: %d bytes)", ErrFileTooLarge, header.Name, header.Size, maxSize)
	}

	file, err := target.fs.OpenFile(target.absPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode&_mode777))
	if err != nil {
		return err
	}
	defer file.Close()

	written, err := io.Copy(file, io.LimitReader(tarReader, header.Size))
	if err != nil {
		return err
	}

	if written != header.Size {
		return fmt.Errorf("%w: %s (wrote %d of %d bytes)", ErrIncompleteWrite, header.Name, written, header.Size)
	}

	return nil
}

func (p *FsPath) prepareUntarEnvironment() (*tar.Reader, func(), error) {
	file, err := p.fs.Open(p.absPath)
	if err != nil {
		return nil, nil, err
	}

	gzr, err := gzip.NewReader(file)
	if err != nil {
		file.Close()
		return nil, nil, err
	}

	tarReader := tar.NewReader(gzr)

	cleanup := func() {
		gzr.Close()
		file.Close()
	}

	return tarReader, cleanup, nil
}

func (p *FsPath) extractZipFile(file *zip.File, destDir *FsPath, maxSize int64) error {
	// Use FsPath.Join instead of filepath.Join
	filePath := destDir.Join(file.Name)

	// Check for path traversal
	if !strings.HasPrefix(filePath.absPath, destDir.absPath) {
		return fmt.Errorf("%w: illegal file path %s", ErrIllegalFilePath, file.Name)
	}

	if file.FileInfo().IsDir() {
		return filePath.MkdirAll(_mode755)
	}

	if file.UncompressedSize64 > uint64(maxSize) {
		return fmt.Errorf("%w: %s (size: %d bytes, max allowed: %d bytes)", ErrFileTooLarge, file.Name, file.UncompressedSize64, maxSize)
	}

	srcFile, err := file.Open()
	if err != nil {
		return fmt.Errorf("failed to open file in zip: %w", err)
	}
	defer srcFile.Close()

	if err := filePath.MkParentDir(); err != nil {
		return err
	}

	destFile, err := filePath.fs.OpenFile(filePath.absPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, file.Mode())
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	written, err := io.Copy(destFile, io.LimitReader(srcFile, maxSize))
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	if written != int64(file.UncompressedSize64) {
		return fmt.Errorf("%w: %s (wrote %d of %d bytes)", ErrIncompleteWrite, file.Name, written, file.UncompressedSize64)
	}

	return nil
}

func (p *FsPath) prepareCompression(fileName, extension string) (*FsPath, afero.File, error) {
	if !p.IsDir() {
		return nil, nil, fmt.Errorf("%w: %s", ErrNotDirectory, p.absPath)
	}

	// Ensure the fileName has the correct extension
	if !strings.HasSuffix(fileName, extension) {
		fileName += extension
	}

	// Create the path for the compressed file
	compressedPath := p.Parent().Join(fileName)

	// Create the file
	file, err := p.fs.Create(compressedPath.absPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create %s file: %w", extension, err)
	}

	return compressedPath, file, nil
}

// Update the type definition for writerFactory
type (
	writerFactory func(name string, info os.FileInfo) (io.Writer, error)
)

func (p *FsPath) compressDirectoryToWriter(createWriter writerFactory) (int, error) {
	totalFiles := 0

	err := p.Walk(func(relPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := p.fs.Open(p.Join(relPath).absPath)
		if err != nil {
			return fmt.Errorf("failed to open file: %w", err)
		}
		defer file.Close()

		writer, err := createWriter(relPath, info)
		if err != nil {
			return fmt.Errorf("failed to create writer for %s: %w", relPath, err)
		}

		_, err = io.Copy(writer, file)
		if err != nil {
			return fmt.Errorf("failed to write file %s: %w", relPath, err)
		}

		totalFiles++

		return nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to compress directory: %w", err)
	}

	return totalFiles, nil
}
