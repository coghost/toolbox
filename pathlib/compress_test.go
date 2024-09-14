package pathlib

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type CompressSuite struct {
	suite.Suite
	tempDir string
}

func TestCompressSuite(t *testing.T) {
	suite.Run(t, new(CompressSuite))
}

func (s *CompressSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
}

func (s *CompressSuite) TearDownSuite() {
}

func (s *CompressSuite) createTestFiles(dirPath *FsPath) []struct {
	name    string
	content string
} {
	testFiles := []struct {
		name    string
		content string
	}{
		{"file1.txt", "Content of file 1"},
		{"file2.txt", "Content of file 2"},
		{"subdir/file3.txt", "Content of file 3 in subdirectory"},
		{"subdir1/subdir2/file4.txt", "Content of file 4 in sub/subdirectory"},
	}

	for i, tf := range testFiles {
		filePath := dirPath.Join(tf.name)
		s.Require().NoError(filePath.MkParentDir())
		s.Require().NoError(filePath.WriteText(tf.content))
		relPath, err := filePath.RelativeTo(dirPath.absPath)
		s.Require().NoError(err, "Failed to get relative path")
		testFiles[i].name = relPath
	}

	return testFiles
}

func (s *CompressSuite) TestCompression() {
	testCases := []struct {
		name          string
		compressFunc  func(*FsPath, string) (*FsPath, int, error)
		extractFunc   func(*FsPath, string, ...CompressOption) error
		fileExtension string
	}{
		{
			name: "Tar.gz",
			compressFunc: func(dirPath *FsPath, fileName string) (*FsPath, int, error) {
				return dirPath.TarGzDir(fileName)
			},
			extractFunc: func(archivePath *FsPath, destDir string, opts ...CompressOption) error {
				return archivePath.Untar(destDir, opts...)
			},
			fileExtension: ".tar.gz",
		},
		{
			name: "Zip",
			compressFunc: func(dirPath *FsPath, fileName string) (*FsPath, int, error) {
				return dirPath.ZipDir(fileName)
			},
			extractFunc: func(archivePath *FsPath, destDir string, opts ...CompressOption) error {
				return archivePath.Unzip(destDir, opts...)
			},
			fileExtension: ".zip",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			dirPath := Path(s.tempDir)
			testFiles := s.createTestFiles(dirPath)
			wantedLen := len(testFiles)

			s.Run("Successful compression", func() {
				compressedPath, totalFiles, err := tc.compressFunc(dirPath, "test_archive"+tc.fileExtension)
				s.Require().NoError(err)
				s.NotNil(compressedPath)
				s.Equal(wantedLen, totalFiles)

				s.True(compressedPath.Exists())

				s.verifyCompressedContent(compressedPath, testFiles, tc.extractFunc)
			})

			s.Run("Overwrite existing compressed file", func() {
				_, _, err := tc.compressFunc(dirPath, "overwrite_test"+tc.fileExtension)
				s.Require().NoError(err)

				compressedPath, totalFiles, err := tc.compressFunc(dirPath, "overwrite_test"+tc.fileExtension)
				s.Require().NoError(err)
				s.NotNil(compressedPath)
				s.Equal(wantedLen, totalFiles)

				s.True(compressedPath.Exists())
			})

			s.Run("Non-directory path", func() {
				filePath := dirPath.Join("file1.txt")
				_, _, err := tc.compressFunc(filePath, "should_fail"+tc.fileExtension)
				s.Require().Error(err)
				s.ErrorIs(err, ErrNotDirectory)
			})

			s.Run("Empty directory", func() {
				emptyDir := dirPath.Join("empty_dir")
				s.Require().NoError(emptyDir.Mkdir(0o755, true))

				compressedPath, totalFiles, err := tc.compressFunc(emptyDir, "empty"+tc.fileExtension)
				s.Require().NoError(err)
				s.NotNil(compressedPath)
				s.Equal(0, totalFiles)

				s.True(compressedPath.Exists())

				_ = compressedPath.Unlink(true)
			})

			s.Run("Extract with custom max size", func() {
				compressedPath, _, err := tc.compressFunc(dirPath, "max_size_test"+tc.fileExtension)
				s.Require().NoError(err)

				extractDir := s.T().TempDir()
				smallMaxSize := int64(10) // Only allow 10 bytes
				err = tc.extractFunc(compressedPath, extractDir, WithMaxSize(smallMaxSize))
				s.Require().Error(err)
				s.ErrorIs(err, ErrFileTooLarge)
			})
		})
	}
}

func (s *CompressSuite) verifyCompressedContent(compressedPath *FsPath, expectedFiles []struct{ name, content string }, extractFunc func(*FsPath, string, ...CompressOption) error) {
	s.T().Helper()
	extractPath := Path(s.T().TempDir())

	err := extractFunc(compressedPath, extractPath.absPath)
	s.Require().NoError(err, "Failed to extract archive")

	archiveName := strings.TrimSuffix(strings.TrimSuffix(compressedPath.Name, ".tar.gz"), ".zip")
	extractedContentPath := extractPath.Join(archiveName)

	for _, ef := range expectedFiles {
		extractedFilePath := extractedContentPath.Join(ef.name)
		s.True(extractedFilePath.Exists(), "File not found in extracted content: "+ef.name)

		content, err := extractedFilePath.ReadText()
		s.Require().NoError(err, "Failed to read extracted file: "+ef.name)
		s.Equal(ef.content, content, "Content mismatch for file: "+ef.name)
	}

	var foundFiles []string
	err = extractedContentPath.Walk(func(relPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			foundFiles = append(foundFiles, relPath)
		}
		return nil
	})
	s.Require().NoError(err, "Failed to walk extracted directory")

	s.Equal(len(expectedFiles), len(foundFiles), "Number of extracted files doesn't match expected")

	// Print found files for debugging
	if len(foundFiles) != len(expectedFiles) {
		s.T().Logf("Found files: %v", foundFiles)
		s.T().Logf("Expected files: %v", expectedFiles)
	}
}
