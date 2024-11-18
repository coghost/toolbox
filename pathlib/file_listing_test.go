package pathlib

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func (s *PathSuite) TestListFilesWithGlobStatic() {
	s.createTempFile("file1.txt", "")
	s.createTempFile("file2.txt", "")
	s.createTempFile("file3.json", "")

	files, err := ListFilesWithGlob(nil, s.tempDir, "*.txt")
	s.Require().NoError(err)
	s.Len(files, 2)

	// Use filepath.Base to compare just the file names
	fileNames := make([]string, len(files))
	for i, file := range files {
		fileNames[i] = filepath.Base(file)
	}

	s.ElementsMatch([]string{"file1.txt", "file2.txt"}, fileNames)
}

func (s *PathSuite) TestListFilesWithGlob() {
	// Setup test files
	s.createTempFile("file1.txt", "")
	s.createTempFile("file2.txt", "")
	s.createTempFile("file3.json", "")
	s.createTempFile(".hiddenfile", "")

	file := Path(s.tempDir)

	tests := []struct {
		name     string
		pattern  string
		expected int
	}{
		{"all files (empty pattern)", "", 4},
		{"all files (asterisk)", "*", 4},
		{"txt files", "*.txt", 2},
		{"json files", "*.json", 1},
		{"hidden files", ".*", 1},
		{"non-existent pattern", "*.go", 0},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result, err := file.ListFilesWithGlob(tt.pattern)
			s.Require().NoError(err)
			s.Len(result, tt.expected)

			if tt.pattern == "*.txt" {
				s.True(strings.HasSuffix(result[0], "file1.txt") || strings.HasSuffix(result[0], "file2.txt"))
				s.True(strings.HasSuffix(result[1], "file1.txt") || strings.HasSuffix(result[1], "file2.txt"))
			}
		})
	}
}

func (s *PathSuite) TestWalk() {
	// Create a temporary directory structure for testing
	tempDir := s.T().TempDir()
	rootPath := Path(tempDir)

	// Create test directory structure
	testFiles := []struct {
		path    string
		content string
		isDir   bool
	}{
		{"file1.txt", "Content of file 1", false},
		{"dir1", "", true},
		{"dir1/file2.txt", "Content of file 2", false},
		{"dir1/subdir", "", true},
		{"dir1/subdir/file3.txt", "Content of file 3", false},
	}

	for _, tf := range testFiles {
		filePath := rootPath.Join(tf.path)
		if tf.isDir {
			err := filePath.MkdirAll(0o755)
			s.Require().NoError(err)
		} else {
			err := filePath.MkParentDir()
			s.Require().NoError(err)
			err = filePath.WriteText(tf.content)
			s.Require().NoError(err)
		}
	}

	// Test walking the directory
	var visited []string
	err := rootPath.Walk(func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		visited = append(visited, path)
		return nil
	})

	s.Require().NoError(err)

	// Check if all paths were visited
	expectedPaths := []string{
		".",
		"file1.txt",
		"dir1",
		filepath.Join("dir1", "file2.txt"),
		filepath.Join("dir1", "subdir"),
		filepath.Join("dir1", "subdir", "file3.txt"),
	}

	s.ElementsMatch(expectedPaths, visited)

	// Test error handling
	err = rootPath.Walk(func(path string, info fs.FileInfo, err error) error {
		if path == "dir1" {
			return errTest
		}
		return nil
	})

	s.Require().Error(err)
	s.Equal(errTest, err)
}
