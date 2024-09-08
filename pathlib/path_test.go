package pathlib

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

const (
	_testContent = "Hello, World!"
)

type PathSuite struct {
	suite.Suite
	tempDir string
}

func TestPath(t *testing.T) {
	suite.Run(t, new(PathSuite))
}

func (s *PathSuite) SetupTest() {
	s.tempDir = s.T().TempDir()
}

func (s *PathSuite) SetupSuite() {
}

func (s *PathSuite) TearDownSuite() {
}

func (s *PathSuite) createTempFile(name, content string) string {
	s.T().Helper()
	path := filepath.Join(s.tempDir, name)
	err := os.WriteFile(path, []byte(content), _mode644)
	s.Require().NoError(err)
	return path
}

func (s *PathSuite) TestNameAndNameWithSuffix() {
	tests := []struct {
		name         string
		filePath     string
		expectedStem string
		expectedName string
	}{
		{
			name:         "simple file",
			filePath:     "/tmp/test.txt",
			expectedStem: "test",
			expectedName: "test.txt",
		},
		{
			name:         "file with multiple extensions",
			filePath:     "/home/user/document.tar.gz",
			expectedStem: "document.tar",
			expectedName: "document.tar.gz",
		},
		{
			name:         "hidden file",
			filePath:     "/home/user/.config",
			expectedStem: "",
			expectedName: ".config",
		},
		{
			name:         "directory",
			filePath:     "/var/log/",
			expectedStem: "log",
			expectedName: "log",
		},
		{
			name:         "file without extension",
			filePath:     "/bin/bash",
			expectedStem: "bash",
			expectedName: "bash",
		},
		{
			name:         "root directory",
			filePath:     "/",
			expectedStem: "/",
			expectedName: "/",
		},
		{
			name:         "file with dot in name",
			filePath:     "/home/user/file.name.with.dots.txt",
			expectedStem: "file.name.with.dots",
			expectedName: "file.name.with.dots.txt",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.filePath)
			s.Equal(tt.expectedStem, file.Stem, "Unexpected Stem for %s", tt.filePath)
			s.Equal(tt.expectedName, file.Name, "Unexpected Name for %s", tt.filePath)
		})
	}
}

func (s *PathSuite) TestSuffix() {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"file with extension", "/tmp/test.txt", ".txt"},
		{"file without extension", "/tmp/test", ""},
		{"hidden file", "/tmp/.hidden.txt", ".txt"},
		{"directory", "/tmp/dir/", ""},
		{"file with multiple extensions", "/tmp/archive.tar.gz", ".gz"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			s.Equal(tt.expected, file.Suffix)
		})
	}
}

func (s *PathSuite) TestAbsPath() {
	tmpDir := s.T().TempDir()
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"absolute path", "/tmp/test.txt", "/tmp/test.txt"},
		{"relative path", "test.txt", filepath.Join(tmpDir, "test.txt")},
		{"dot path", ".", tmpDir},
		{"parent path", "..", filepath.Dir(tmpDir)},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			currentDir, err := os.Getwd()
			s.Require().NoError(err)

			err = os.Chdir(tmpDir)
			s.Require().NoError(err)

			defer func() {
				err := os.Chdir(currentDir)
				s.Require().NoError(err)
			}()

			file := Path(tt.path)

			// Remove "/private" prefix from both expected and actual paths
			expectedPath := strings.TrimPrefix(tt.expected, "/private")
			actualPath := strings.TrimPrefix(file.AbsPath, "/private")

			s.Equal(expectedPath, actualPath)
		})
	}
}

func (s *PathSuite) TestPath() {
	// Setup: Create necessary directories and files
	tmpDir := s.T().TempDir()

	// Create /tmp/testdir/
	testDir := filepath.Join(tmpDir, "testdir")
	err := os.MkdirAll(testDir, _mode755)
	s.Require().NoError(err)

	// Create /tmp/test.txt
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), _mode644)
	s.Require().NoError(err)

	tests := []struct {
		name     string
		path     string
		expected FSPath
	}{
		{
			name: "file path",
			path: testFile,
			expected: FSPath{
				RawPath: testFile,
				Stem:    "test",
				Name:    "test.txt",
				Suffix:  ".txt",
				AbsPath: testFile,
				// Note: We don't compare the 'fs' field directly
			},
		},
		{
			name: "directory path",
			path: testDir,
			expected: FSPath{
				RawPath: testDir,
				Stem:    "testdir",
				Name:    "testdir",
				Suffix:  "",
				AbsPath: testDir,
				// Note: We don't compare the 'fs' field directly
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := Path(tt.path)

			// Compare fields individually, excluding 'fs'
			s.Equal(tt.expected.RawPath, result.RawPath)
			s.Equal(tt.expected.Stem, result.Stem)
			s.Equal(tt.expected.Name, result.Name)
			s.Equal(tt.expected.Suffix, result.Suffix)
			s.Equal(tt.expected.AbsPath, result.AbsPath)

			// Check that 'fs' is not nil and is of the expected type
			s.NotNil(result.fs)
			_, ok := result.fs.(*afero.OsFs)
			s.True(ok, "Expected fs to be of type *afero.OsFs")
		})
	}
}

func (s *PathSuite) TestStat() {
	path := s.createTempFile("stattest.txt", "content")
	fspath := Path(path)

	info, err := fspath.Stat()
	s.Require().NoError(err)
	s.NotNil(info)
	s.Equal("stattest.txt", info.Name())
	s.Equal(int64(7), info.Size()) // "content" is 7 bytes
}

func (s *PathSuite) TestExpand() {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"home directory", "~/documents", filepath.Join(os.Getenv("HOME"), "documents")},
		{"environment variable", "$HOME/documents", filepath.Join(os.Getenv("HOME"), "documents")},
		{"no expansion needed", "/tmp/file.txt", "/tmp/file.txt"},
		{"empty path", "", ""},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			expanded := Expand(tt.path)
			s.Equal(tt.expected, expanded)
		})
	}
}

func (s *PathSuite) TestExists() {
	existingFile := s.createTempFile("existing.txt", "content")
	nonExistingFile := filepath.Join(s.tempDir, "non_existing.txt")

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing directory", s.tempDir, true},
		{"existing file", existingFile, true},
		{"non-existing file", nonExistingFile, false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			s.Equal(tt.expected, file.Exists())
		})
	}
}

func (s *PathSuite) TestIsDir() {
	file := s.createTempFile("file.txt", "content")

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"directory", s.tempDir, true},
		{"file", file, false},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			path := Path(tt.path)
			s.Equal(tt.expected, path.IsDir())
		})
	}
}

func (s *PathSuite) TestMkParentDir() {
	path := filepath.Join(s.tempDir, "new", "parent", "dir", "file.txt")
	file := Path(path)

	err := file.MkParentDir()
	s.Require().NoError(err)
	s.DirExists(filepath.Dir(path))
}

func (s *PathSuite) TestMkDirs() {
	path := filepath.Join(s.tempDir, "new", "nested", "dir")
	file := Path(path)

	err := file.MkDirs()
	s.Require().NoError(err)
	s.DirExists(path)
}

func (s *PathSuite) TestSetGetString() {
	path := filepath.Join(s.tempDir, "test.txt")
	file := Path(path)

	err := file.SetString(_testContent)
	s.Require().NoError(err)

	result, err := file.GetString()
	s.Require().NoError(err)
	s.Equal(_testContent, result)
}

func (s *PathSuite) TestSetGetBytes() {
	path := filepath.Join(s.tempDir, "test.bin")
	file := Path(path)
	content := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}

	err := file.SetBytes(content)
	s.Require().NoError(err)

	result, err := file.GetBytes()
	s.Require().NoError(err)
	s.Equal(content, result)
}

func (s *PathSuite) TestReader() {
	path := s.createTempFile("test.txt", _testContent)
	file := Path(path)

	reader, err := file.Reader()
	s.Require().NoError(err)

	result, err := io.ReadAll(reader)
	s.Require().NoError(err)
	s.Equal(_testContent, string(result))
}

func (s *PathSuite) TestCopy() {
	srcPath := s.createTempFile("src.txt", _testContent)
	dstPath := filepath.Join(s.tempDir, "dst.txt")

	file := Path(srcPath)
	err := file.Copy(dstPath)
	s.Require().NoError(err)

	dstContent, err := os.ReadFile(dstPath)
	s.Require().NoError(err)
	s.Equal(_testContent, string(dstContent))
}

func (s *PathSuite) TestMove() {
	srcPath := s.createTempFile("src.txt", _testContent)
	dstPath := filepath.Join(s.tempDir, "dst.txt")

	file := Path(srcPath)
	err := file.Move(dstPath)
	s.Require().NoError(err)

	s.NoFileExists(srcPath)
	s.FileExists(dstPath)

	dstContent, err := os.ReadFile(dstPath)
	s.Require().NoError(err)
	s.Equal(_testContent, string(dstContent))
}

func (s *PathSuite) TestRename() {
	oldPath := s.createTempFile("oldfile.txt", "content")
	newPath := filepath.Join(filepath.Dir(oldPath), "newfile.txt")

	fspath := Path(oldPath)
	err := fspath.Rename(newPath)
	s.Require().NoError(err)
	s.NoFileExists(oldPath)
	s.FileExists(newPath)
}

func (s *PathSuite) TestListFilesWithGlob() {
	s.createTempFile("file1.txt", "")
	s.createTempFile("file2.txt", "")
	s.createTempFile("file3.json", "")

	file := Path(s.tempDir)

	result, err := file.ListFilesWithGlob("*.txt")
	s.Require().NoError(err)
	s.Len(result, 2)
	s.True(strings.HasSuffix(result[0], "file1.txt") || strings.HasSuffix(result[0], "file2.txt"))
	s.True(strings.HasSuffix(result[1], "file1.txt") || strings.HasSuffix(result[1], "file2.txt"))
}

func (s *PathSuite) TestListFilesWithGlobEmptyPattern() {
	s.createTempFile("file1.txt", "")
	s.createTempFile("file2.txt", "")
	s.createTempFile("file3.json", "")

	file := Path(s.tempDir)

	result, err := file.ListFilesWithGlob("")
	s.Require().NoError(err)
	s.Len(result, 3)
}

func (s *PathSuite) TestListFilesWithGlobStatic() {
	s.createTempFile("file1.txt", "")
	s.createTempFile("file2.txt", "")
	s.createTempFile("file3.json", "")

	files, err := ListFilesWithGlob(s.tempDir, "*.txt")
	s.Require().NoError(err)
	s.Len(files, 2)

	// Use filepath.Base to compare just the file names
	fileNames := make([]string, len(files))
	for i, file := range files {
		fileNames[i] = filepath.Base(file)
	}

	s.ElementsMatch([]string{"file1.txt", "file2.txt"}, fileNames)
}

func (s *PathSuite) TestWithSuffixInNewDir() {
	tests := []struct {
		name      string
		filePath  string
		newSuffix string
		want      string
	}{
		{
			name:      "file in subdirectory",
			filePath:  "/tmp/a/b/file.txt",
			newSuffix: "json",
			want:      "/tmp/a/b_json/file.json",
		},
		{
			name:      "file in root directory",
			filePath:  "/file.txt",
			newSuffix: "json",
			want:      "/_json/file.json",
		},
		{
			name:      "directory path",
			filePath:  "/tmp/a/b/",
			newSuffix: "backup",
			want:      "/tmp/a_backup/b.backup",
		},
		{
			name:      "file without extension",
			filePath:  "/tmp/a/b/file",
			newSuffix: "txt",
			want:      "/tmp/a/b_txt/file.txt",
		},
		{
			name:      "file with multiple extensions",
			filePath:  "/tmp/a/b/archive.tar.gz",
			newSuffix: "bak",
			want:      "/tmp/a/b_bak/archive.tar.bak",
		},
		{
			name:      "empty new suffix",
			filePath:  "/tmp/a/b/file.txt",
			newSuffix: "",
			want:      "/tmp/a/b_/file",
		},
		{
			name:      "suffix with dot",
			filePath:  "/tmp/a/b/file.txt",
			newSuffix: ".json",
			want:      "/tmp/a/b_json/file.json",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.filePath)
			got := file.WithSuffixAndSuffixedParentDir(tt.newSuffix)
			s.Equal(tt.want, got.AbsPath, "For input: %s", tt.filePath)

			// Additional checks
			s.Equal(filepath.Base(tt.want), got.Name)
			s.Equal(filepath.Dir(tt.want), got.Parent().AbsPath)

			// Check that the original path hasn't changed
			s.NotEqual(got.AbsPath, file.AbsPath, "Original path should not be modified")

			// Check that the new file has the correct suffix
			if tt.newSuffix != "" {
				s.Equal("."+strings.TrimPrefix(tt.newSuffix, "."), filepath.Ext(got.Name))
			} else {
				s.Equal("", filepath.Ext(got.Name))
			}
		})
	}
}

func (s *PathSuite) TestParts() {
	currentDir, err := os.Getwd()
	s.Require().NoError(err)

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name:     "simple absolute path",
			path:     "/usr/bin/golang",
			expected: []string{"/", "usr", "bin", "golang"},
		},
		{
			name:     "path with trailing slash",
			path:     "/usr/local/",
			expected: []string{"/", "usr", "local"},
		},
		{
			name:     "root directory",
			path:     "/",
			expected: []string{"/"},
		},
		{
			name:     "path without leading slash",
			path:     "home/user/documents",
			expected: append([]string{"/"}, strings.Split(filepath.Join(currentDir, "home", "user", "documents"), string(os.PathSeparator))[1:]...),
		},
		{
			name:     "path with multiple consecutive slashes",
			path:     "/var///log/messages",
			expected: []string{"/", "var", "log", "messages"},
		},
		{
			name:     "path with dot",
			path:     "/etc/./config",
			expected: []string{"/", "etc", "config"},
		},
		{
			name:     "path with double dot",
			path:     "/usr/local/../bin",
			expected: []string{"/", "usr", "bin"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: append([]string{"/"}, strings.Split(currentDir, string(os.PathSeparator))[1:]...),
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			got := file.Parts()

			s.Equal(tt.expected, got, "For path: %s", tt.path)
		})
	}
}

func (s *PathSuite) TestParent() {
	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"file in directory", "/tmp/a/b/test.txt", "/tmp/a/b"},
		{"directory", "/tmp/a/b/c/", "/tmp/a/b"},
		{"root directory", "/", "/"},
		{"file in root", "/test.txt", "/"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			s.Equal(tt.expected, file.Parent().AbsPath)
		})
	}
}

func (s *PathSuite) TestParents() {
	currentDir, err := os.Getwd()
	s.Require().NoError(err)

	tests := []struct {
		name string
		raw  string
		n    int
		want string
	}{
		{
			name: "zero levels up",
			raw:  "/tmp/b/93877/c/4696890",
			n:    0,
			want: "/tmp/b/93877/c/4696890",
		},
		{
			name: "one level up",
			raw:  "/tmp/b/93877/c/4696890",
			n:    1,
			want: "/tmp/b/93877/c",
		},
		{
			name: "two levels up",
			raw:  "/tmp/b/93877/c/4696890",
			n:    2,
			want: "/tmp/b/93877",
		},
		{
			name: "all the way up",
			raw:  "/tmp/b/93877/c/4696890",
			n:    5,
			want: "/",
		},
		{
			name: "beyond root",
			raw:  "/tmp/b/93877/c/4696890",
			n:    10,
			want: "/",
		},
		{
			name: "from root",
			raw:  "/",
			n:    1,
			want: "/",
		},
		{
			name: "relative path",
			raw:  "documents/subdirectory/file.txt",
			n:    2,
			want: filepath.Join(currentDir, "documents"),
		},
		{
			name: "with trailing slash",
			raw:  "/tmp/b/93877/c/4696890/",
			n:    1,
			want: "/tmp/b/93877/c",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.raw)
			got := file.Parents(tt.n)

			if filepath.IsAbs(tt.raw) {
				s.Equal(tt.want, got.AbsPath, "Unexpected parent path")
			} else {
				// For relative paths, we need to compare with the resolved absolute path
				expectedAbs, err := filepath.Abs(tt.want)
				s.Require().NoError(err)
				s.Equal(expectedAbs, got.AbsPath, "Unexpected parent path")
			}
		})
	}
}

func (s *PathSuite) TestBaseDir() {
	tests := []struct {
		name     string
		filePath string
		wantBase string
	}{
		{
			name:     "file in subdirectory",
			filePath: "/tmp/a/b/c.txt",
			wantBase: "b",
		},
		{
			name:     "directory",
			filePath: "/tmp/a/b/c/",
			wantBase: "c",
		},
		{
			name:     "file in root",
			filePath: "/tmp/c.txt",
			wantBase: "tmp",
		},
		{
			name:     "root directory",
			filePath: "/",
			wantBase: "/",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.filePath)
			s.Equal(tt.wantBase, file.BaseDir(), "Unexpected BaseDir")
		})
	}
}

func (s *PathSuite) TestOriginalName() {
	tests := []struct {
		name     string
		filePath string
	}{
		{"file path", "/tmp/test.txt"},
		{"directory path", "/tmp/testdir/"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.filePath)
			s.Equal(tt.filePath, file.RawPath)
		})
	}
}

func (s *PathSuite) TestSplitPath() {
	tests := []struct {
		name     string
		filePath string
		wantDir  string
		wantName string
	}{
		{"file in directory", "/tmp/test.txt", "/tmp/", "test.txt"},
		{"root file", "/test.txt", "/", "test.txt"},
		{"directory", "/tmp/testdir/", "/tmp/testdir/", ""},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.filePath)
			dir, name := file.Split(tt.filePath)
			s.Equal(tt.wantDir, dir)
			s.Equal(tt.wantName, name)
		})
	}
}

func (s *PathSuite) TestMustSetString() {
	path := filepath.Join(s.tempDir, "test.txt")
	file := Path(path)

	file.MustSetString(_testContent)

	result, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Equal(_testContent, string(result))
}

func (s *PathSuite) TestMustGetBytes() {
	path := s.createTempFile("test.txt", _testContent)
	file := Path(path)

	result := file.MustGetBytes()
	s.Equal([]byte(_testContent), result)
}

func (s *PathSuite) TestCSVGetSlices() {
	content := "a,b,c\n1,2,3\n4,5,6"
	path := s.createTempFile("test.csv", content)
	file := Path(path)

	result, err := file.CSVGetSlices()
	s.Require().NoError(err)
	s.Equal([][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}}, result)
}

func (s *PathSuite) TestMustCSVGetSlices() {
	content := "a,b,c\n1,2,3\n4,5,6"
	path := s.createTempFile("test.csv", content)
	file := Path(path)

	result := file.MustCSVGetSlices()
	s.Equal([][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}}, result)
}

func (s *PathSuite) TestCSVAndTSVWithComments() {
	csvContent := "# This is a comment\na,b,c\n1,2,3\n# Another comment\n4,5,6"
	tsvContent := "# This is a comment\na\tb\tc\n1\t2\t3\n# Another comment\n4\t5\t6"

	csvPath := s.createTempFile("test.csv", csvContent)
	tsvPath := s.createTempFile("test.tsv", tsvContent)

	csvFile := Path(csvPath)
	tsvFile := Path(tsvPath)

	expectedResult := [][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}}

	csvResult, err := csvFile.CSVGetSlices()
	s.Require().NoError(err)
	s.Equal(expectedResult, csvResult)

	tsvResult, err := tsvFile.TSVGetSlices()
	s.Require().NoError(err)
	s.Equal(expectedResult, tsvResult)
}

func (s *PathSuite) TestListFilesWithGlobPatterns() {
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
		{"all files", "", 4},
		{"all files", "*", 4},
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
		})
	}
}

func (s *PathSuite) TestTSVGetSlices() {
	content := "a\tb\tc\n1\t2\t3\n4\t5\t6"
	path := s.createTempFile("test.tsv", content)
	file := Path(path)

	result, err := file.TSVGetSlices()
	s.Require().NoError(err)
	s.Equal([][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}}, result)
}

func (s *PathSuite) TestMustTSVGetSlices() {
	content := "a\tb\tc\n1\t2\t3\n4\t5\t6"
	path := s.createTempFile("test.tsv", content)
	file := Path(path)

	result := file.MustTSVGetSlices()
	s.Equal([][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}}, result)
}

func (s *PathSuite) TestEPanic() {
	file := Path("/tmp/test.txt")
	s.Panics(func() {
		file.e(nil, errors.New("test error")) //nolint
	})
}

func (s *PathSuite) TestENoError() {
	fspath := Path("/tmp/test.txt")
	s.NotPanics(func() {
		fspath.e(nil) // Should not panic
	})
}

func (s *PathSuite) TestReadDelimitedFile() {
	tests := []struct {
		name      string
		content   string
		delimiter rune
		expected  [][]string
	}{
		{
			name:      "comma delimiter",
			content:   "a,b,c\n1,2,3\n4,5,6",
			delimiter: ',',
			expected:  [][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}},
		},
		{
			name:      "tab delimiter",
			content:   "a\tb\tc\n1\t2\t3\n4\t5\t6",
			delimiter: '\t',
			expected:  [][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}},
		},
		{
			name:      "semicolon delimiter",
			content:   "a;b;c\n1;2;3\n4;5;6",
			delimiter: ';',
			expected:  [][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			path := s.createTempFile("test.txt", tt.content)
			file := Path(path)

			result, err := file.readDelimitedFile(tt.delimiter)
			s.Require().NoError(err)
			s.Equal(tt.expected, result)
		})
	}
}

func (s *PathSuite) TestReadDelimitedFileErrors() {
	nonExistentPath := filepath.Join(s.tempDir, "nonexistent.csv")
	fspath := Path(nonExistentPath)

	_, err := fspath.readDelimitedFile(',')
	s.Error(err)
}

func (s *PathSuite) TestMustMethodsPanic() {
	nonExistentPath := filepath.Join(s.tempDir, "nonexistent.txt")
	fspath := Path(nonExistentPath)

	s.Panics(func() { fspath.MustGetBytes() })
	s.Panics(func() { fspath.MustCSVGetSlices() })
	s.Panics(func() { fspath.MustTSVGetSlices() })
}

func (s *PathSuite) TestWithName() {
	tmpDir := s.T().TempDir()

	tests := []struct {
		name     string
		path     string
		newName  string
		expected string
	}{
		{"change file name", "/home/user/file.txt", "newfile.txt", "/home/user/newfile.txt"},
		{"change directory name", "/home/user/docs/", "newdocs", "/home/user/newdocs"},
		{"change root-level file name", "/file.txt", "newfile.txt", "/newfile.txt"},
		{"change to name with extension", "/home/user/file.txt", "newfile.csv", "/home/user/newfile.csv"},
		{"change to name without extension", "/home/user/file.txt", "newfile", "/home/user/newfile"},
		{"change root-level file name", "/", "newfile.txt", "/newfile.txt"},
		{"sibling", "/tmp/a/b/current.txt", "newFolder/abc.txt", "/tmp/a/b/newFolder/abc.txt"},
		{
			name:     "create sibling dir for file path",
			path:     filepath.Join(tmpDir, "a", "b", "test.txt"),
			newName:  "newFolder",
			expected: filepath.Join(tmpDir, "a", "b", "newFolder"),
		},
		{
			name:     "create sibling dir for directory path",
			path:     filepath.Join(tmpDir, "x", "y", "z"),
			newName:  "newFolder",
			expected: filepath.Join(tmpDir, "x", "y", "newFolder"),
		},
		{
			name:     "create new dir at root",
			path:     "/",
			newName:  "newFolder",
			expected: "/newFolder",
		},
		{
			name:     "change name of hidden file",
			path:     "/home/user/.config",
			newName:  ".newconfig",
			expected: "/home/user/.newconfig",
		},
		{
			name:     "change name with spaces",
			path:     "/home/user/my documents",
			newName:  "my new documents",
			expected: "/home/user/my new documents",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			newFile := file.WithName(tt.newName)
			s.Equal(tt.expected, newFile.AbsPath)

			// Additional checks
			s.Equal(filepath.Base(tt.expected), newFile.Name)
			s.Equal(filepath.Dir(tt.expected), newFile.Parent().AbsPath)

			// Check that the original path hasn't changed
			s.NotEqual(newFile.AbsPath, file.AbsPath, "Original path should not be modified")
		})
	}
}

func (s *PathSuite) TestJoinPath() {
	tests := []struct {
		name     string
		path     string
		others   []string
		expected string
	}{
		{
			name:     "join relative path to file",
			path:     "/home/user/file.html",
			others:   []string{"..", "newfile.txt"},
			expected: "/home/user/newfile.txt",
		},
		{
			name:     "join multiple components to directory",
			path:     "/home/user/",
			others:   []string{"documents", "file.txt"},
			expected: "/home/user/documents/file.txt",
		},
		{
			name:     "join to file as if it were a directory",
			path:     "/home/user/file.txt",
			others:   []string{"documents", "newfile.txt"},
			expected: "/home/user/file.txt/documents/newfile.txt",
		},
		{
			name:     "join to root directory",
			path:     "/",
			others:   []string{"var", "log"},
			expected: "/var/log",
		},
		{
			name:     "join with empty component",
			path:     "/home/user/",
			others:   []string{"", "documents"},
			expected: "/home/user/documents",
		},
		{
			name:     "join absolute path overrides base path",
			path:     "/home/user/",
			others:   []string{"/var/log"},
			expected: "/var/log",
		},
		{
			name:     "join with parent directory reference",
			path:     "/home/user/",
			others:   []string{"..", "documents"},
			expected: "/home/documents",
		},
		{
			name:     "join to file with parent directory reference",
			path:     "/home/user/file.txt",
			others:   []string{"..", "documents"},
			expected: "/home/user/documents",
		},
		{
			name:     "join to root-level file",
			path:     "/file.txt",
			others:   []string{"documents", "newfile.txt"},
			expected: "/file.txt/documents/newfile.txt",
		},
		{
			name:     "join to root-level file with parent reference",
			path:     "/file.txt",
			others:   []string{"..", "documents", "newfile.txt"},
			expected: "/documents/newfile.txt",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			result := file.JoinPath(tt.others...)
			s.Equal(tt.expected, result.AbsPath)
		})
	}
}

func (s *PathSuite) TestWithStem() {
	tests := []struct {
		name     string
		path     string
		newStem  string
		expected string
	}{
		{"change file stem", "/home/user/file.txt", "newfile", "/home/user/newfile.txt"},
		{"change directory stem", "/home/user/docs/", "newdocs", "/home/user/newdocs"},
		{"change stem of file without extension", "/home/user/file", "newfile", "/home/user/newfile"},
		{"change stem of file with multiple extensions", "/home/user/archive.tar.gz", "newarchive", "/home/user/newarchive.gz"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			newFile := file.WithStem(tt.newStem)
			s.Equal(tt.expected, newFile.AbsPath)
		})
	}
}

func (s *PathSuite) TestWithSuffix() {
	tests := []struct {
		name      string
		path      string
		newSuffix string
		expected  string
	}{
		{"change file suffix", "/home/user/file.txt", ".md", "/home/user/file.md"},
		{"add suffix to directory", "/home/user/docs/", ".txt", "/home/user/docs.txt"},
		{"remove suffix", "/home/user/file.txt", "", "/home/user/file"},
		{"change suffix of file without extension", "/home/user/file", ".txt", "/home/user/file.txt"},
		{"change suffix without dot", "/home/user/file.txt", "md", "/home/user/file.md"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			newFile := file.WithSuffix(tt.newSuffix)
			s.Equal(tt.expected, newFile.AbsPath)
		})
	}
}

func (s *PathSuite) TestWithRenamedParentDir() {
	tests := []struct {
		name       string
		filePath   string
		newDirName string
		want       string
	}{
		{
			name:       "file in subdirectory",
			filePath:   "/tmp/a/b/file.txt",
			newDirName: "c",
			want:       "/tmp/a/c/file.txt",
		},
		{
			name:       "directory path",
			filePath:   "/tmp/a/b/",
			newDirName: "c",
			want:       "/tmp/c/b",
		},
		{
			name:       "file in root directory",
			filePath:   "/file.txt",
			newDirName: "newdir",
			want:       "/newdir/file.txt",
		},
		{
			name:       "root directory",
			filePath:   "/",
			newDirName: "newdir",
			want:       "/",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.filePath)
			got := file.WithRenamedParentDir(tt.newDirName)
			s.Equal(tt.want, got.AbsPath)
		})
	}
}

func (s *PathSuite) Test00Manual() {
}
