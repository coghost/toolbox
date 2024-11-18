package pathlib

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
)

const (
	_testContent = "Hello, World!"
)

var errTest = errors.New("test error")

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
			expectedStem: ".config",
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
		{"hidden file", "/tmp/.hidden", ""},
		{"hidden file with ext", "/tmp/.hidden.txt", ".txt"},
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

func (s *PathSuite) TestSuffixes() {
	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{"no suffix", "/path/to/file", []string{}},
		{"single suffix", "/path/to/file.txt", []string{".txt"}},
		{"multiple suffixes", "/path/to/file.tar.gz", []string{".tar", ".gz"}},
		{"hidden file", "/path/to/.hidden", []string{}},
		{"hidden file with suffix", "/path/to/.hidden.txt", []string{".txt"}},
		{"directory", "/path/to/dir/", []string{}},
		{"root", "/", []string{}},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			s.Equal(tt.expected, file.Suffixes())
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
			actualPath := strings.TrimPrefix(file.absPath, "/private")

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
		expected FsPath
	}{
		{
			name: "file path",
			path: testFile,
			expected: FsPath{
				RawPath: testFile,
				Stem:    "test",
				Name:    "test.txt",
				Suffix:  ".txt",
				absPath: testFile,
				// Note: We don't compare the 'fs' field directly
			},
		},
		{
			name: "directory path",
			path: testDir,
			expected: FsPath{
				RawPath: testDir,
				Stem:    "testdir",
				Name:    "testdir",
				Suffix:  "",
				absPath: testDir,
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
			s.Equal(tt.expected.absPath, result.absPath)

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

	err := file.Mkdirs()
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
			s.Equal(tt.expected, file.Parent().absPath)
		})
	}
}

func (s *PathSuite) TestParents() {
	currentDir, err := Cwd()
	s.Require().NoError(err)

	tests := []struct {
		name     string
		path     string
		expected []string
	}{
		{
			name: "multi-level directory",
			path: "/home/user/documents/file.txt",
			expected: []string{
				"/home/user/documents",
				"/home/user",
				"/home",
				"/",
			},
		},
		{
			name:     "root directory",
			path:     "/",
			expected: []string{},
		},
		{
			name: "single-level directory",
			path: "/home",
			expected: []string{
				"/",
			},
		},
		{
			name: "relative path",
			path: "user/documents/file.txt",
			expected: []string{
				currentDir.Join("user/documents").absPath,
				currentDir.Join("user").absPath,
			},
		},
		{
			name:     "current directory",
			path:     ".",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			parents := file.Parents()

			s.Equal(len(tt.expected), len(parents), "Number of parents doesn't match for path: %s", tt.path)

			for i, parent := range parents {
				s.Equal(tt.expected[i], parent.absPath, "Parent path doesn't match at index %d for path: %s", i, tt.path)
			}
		})
	}
}

func (s *PathSuite) TestParentsUpTo() {
	currentDir, err := Cwd()
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
			want: currentDir.Join("documents").absPath,
		},
		{
			name: "beyond relative path",
			raw:  "documents/subdirectory/file.txt",
			n:    5,
			want: currentDir.Join("documents").ParentsUpTo(3).absPath,
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
			got := file.ParentsUpTo(tt.n)

			if filepath.IsAbs(tt.raw) {
				s.Equal(tt.want, got.absPath, "Unexpected parent path")
			} else {
				// For relative paths, we need to compare with the resolved absolute path
				expectedAbs, err := filepath.Abs(tt.want)
				s.Require().NoError(err)
				s.Equal(expectedAbs, got.absPath, "Unexpected parent path")
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
		file.e(nil, errTest)
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
	s.Require().Error(err)
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
			s.Equal(tt.expected, newFile.absPath)

			// Additional checks
			s.Equal(filepath.Base(tt.expected), newFile.Name)
			s.Equal(filepath.Dir(tt.expected), newFile.Parent().absPath)

			// Check that the original path hasn't changed
			s.NotEqual(newFile.absPath, file.absPath, "Original path should not be modified")
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
			result := file.Join(tt.others...)
			s.Equal(tt.expected, result.absPath)
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
			s.Equal(tt.expected, newFile.absPath)
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
			s.Equal(tt.expected, newFile.absPath)
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
			s.Equal(tt.want, got.absPath)
		})
	}
}

func (s *PathSuite) TestRelativeTo() {
	tests := []struct {
		name     string
		path     string
		other    string
		expected string
		hasError bool
	}{
		{"same directory", "/home/user/file.txt", "/home/user", "file.txt", false},
		{"subdirectory", "/home/user/docs/file.txt", "/home/user", "docs/file.txt", false},
		{"parent directory", "/home/user/file.txt", "/home", "user/file.txt", false},
		{"unrelated paths", "/home/user/file.txt", "/var/log", "../../home/user/file.txt", false},
		{"to root", "/home/user/file.txt", "/", "home/user/file.txt", false},
		{"from root", "/", "/home/user", "../..", false},
		{"same file", "/home/user/file.txt", "/home/user/file.txt", ".", false},
		// {"invalid path", "/home/user/file.txt", "~invaliduser", "", true}, // Changed this line
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.path)
			result, err := file.RelativeTo(tt.other)
			if tt.hasError {
				s.Require().Error(err)
			} else {
				s.Require().NoError(err)
				s.Equal(tt.expected, result)
			}
		})
	}
}

func (s *PathSuite) TestMkdir() {
	// Test creating a single directory
	singleDir := filepath.Join(s.tempDir, "singleDir")
	singlePath := Path(singleDir)
	err := singlePath.Mkdir(0o755, false)
	s.Require().NoError(err)
	s.DirExists(singleDir)

	// Test creating a directory that already exists
	err = singlePath.Mkdir(0o755, false)
	s.Require().Error(err)
	s.True(os.IsExist(err), "Expected 'file exists' error, got: %v", err)

	// Test creating nested directories with parents=true
	nestedDir := filepath.Join(s.tempDir, "parent", "child", "grandchild")
	nestedPath := Path(nestedDir)
	err = nestedPath.Mkdir(0o755, true)
	s.Require().NoError(err)
	s.DirExists(nestedDir)

	// Test creating a nested directory without parents=true
	failDir := filepath.Join(s.tempDir, "fail", "dir")
	failPath := Path(failDir)
	err = failPath.Mkdir(0o755, false)
	s.Require().Error(err)
	s.True(os.IsNotExist(err), "Expected 'no such file or directory' error, got: %v", err)

	// Test creating an existing directory with parents=true
	err = nestedPath.Mkdir(0o755, true)
	s.Require().NoError(err, "Creating an existing directory with parents=true should not error")

	// Test creating a directory with different permissions
	permDir := filepath.Join(s.tempDir, "permDir")
	permPath := Path(permDir)
	err = permPath.Mkdir(0o700, false)
	s.Require().NoError(err)
	info, err := os.Stat(permDir)
	s.Require().NoError(err)
	s.Equal(os.FileMode(0o700), info.Mode().Perm())

	// Test creating nested directories with different permissions
	nestedPermDir := filepath.Join(s.tempDir, "nestedPerm", "child")
	nestedPermPath := Path(nestedPermDir)
	err = nestedPermPath.Mkdir(0o700, true)
	s.Require().NoError(err)
	info, err = os.Stat(nestedPermDir)
	s.Require().NoError(err)
	s.Equal(os.FileMode(0o700), info.Mode().Perm())

	// Test creating a directory in a read-only location (if possible)
	if os.Geteuid() != 0 { // Skip this test if running as root
		readOnlyDir := "/tmp/readOnlyDir"

		err := os.Mkdir(readOnlyDir, _mode555)
		s.Require().NoError(err)

		defer os.RemoveAll(readOnlyDir)

		readOnlyPath := Path(filepath.Join(readOnlyDir, "newDir"))
		err = readOnlyPath.Mkdir(0o755, false)
		s.Require().Error(err)
	}
}

func (s *PathSuite) TestTouch() {
	s.T().Parallel() // Mark this test as parallel

	// Create a unique temporary directory for this test
	testDir, err := os.MkdirTemp("", "TestTouch")
	s.Require().NoError(err)
	defer os.RemoveAll(testDir) // Clean up after the test

	testFile := filepath.Join(testDir, "testTouch.txt")
	path := Path(testFile)

	// Test creating a new file
	err = path.Touch()
	s.Require().NoError(err)
	s.FileExists(testFile)

	// Get initial modification time
	initialStat, err := os.Stat(testFile)
	s.Require().NoError(err)
	initialModTime := initialStat.ModTime()

	// Wait a moment to ensure the modification time can change
	time.Sleep(time.Millisecond * 100)

	// Test updating an existing file
	err = path.Touch()
	s.Require().NoError(err)

	// Check if modification time has been updated
	newStat, err := os.Stat(testFile)
	s.Require().NoError(err)
	s.True(newStat.ModTime().After(initialModTime),
		"New mod time (%v) should be after initial mod time (%v)",
		newStat.ModTime(), initialModTime)
}

func (s *PathSuite) TestChmod() {
	testFile := filepath.Join(s.tempDir, "testChmod.txt")
	path := Path(testFile)

	// Create a test file
	err := os.WriteFile(testFile, []byte("test content"), _mode644)
	s.Require().NoError(err)

	// Test changing permissions
	err = path.Chmod(_mode600)
	s.Require().NoError(err)

	// Check if permissions were changed
	info, err := os.Stat(testFile)
	s.Require().NoError(err)
	s.Equal(os.FileMode(_mode600), info.Mode().Perm())

	// Test changing permissions on a non-existent file
	nonExistentPath := Path(filepath.Join(s.tempDir, "nonexistent.txt"))
	err = nonExistentPath.Chmod(_mode644)
	s.Require().Error(err)
}

func (s *PathSuite) TestUnlink() {
	s.T().Parallel()

	testDir := s.T().TempDir()

	// Test unlinking an existing file
	existingFile := Path(filepath.Join(testDir, "existing.txt"))
	err := existingFile.WriteText("test content")
	s.Require().NoError(err)

	err = existingFile.Unlink(false)
	s.Require().NoError(err)
	s.False(existingFile.Exists())

	// Test unlinking a non-existent file with force=false
	nonExistentFile := Path(filepath.Join(testDir, "nonexistent.txt"))
	err = nonExistentFile.Unlink(false)
	s.Require().Error(err)
	s.True(os.IsNotExist(err))

	// Test unlinking a non-existent file with force=true
	err = nonExistentFile.Unlink(true)
	s.Require().NoError(err)

	// Test unlinking a directory
	dirPath := Path(filepath.Join(testDir, "testdir"))
	err = dirPath.Mkdir(0o755, false)
	s.Require().NoError(err)

	err = dirPath.Unlink(false)
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrCannotUnlinkDir)
	s.True(dirPath.Exists())

	// Test unlinking a symlink
	targetFile := Path(filepath.Join(testDir, "targetfile"))
	err = targetFile.WriteText("target content")
	s.Require().NoError(err)

	symlinkPath := Path(filepath.Join(testDir, "symlink"))
	err = os.Symlink(targetFile.absPath, symlinkPath.absPath)
	s.Require().NoError(err)

	err = symlinkPath.Unlink(false)
	s.Require().NoError(err)

	// Use os.Lstat to check if the symlink itself is gone
	_, err = os.Lstat(symlinkPath.absPath)
	s.True(os.IsNotExist(err), "Symlink should not exist after unlinking")

	// Ensure the target file still exists
	s.True(targetFile.Exists(), "Target file should still exist after unlinking symlink")
}

func (s *PathSuite) TestRmdir() {
	s.T().Parallel()

	testDir := s.T().TempDir()

	// Test removing an empty directory
	emptyDir := Path(filepath.Join(testDir, "emptyDir"))
	err := emptyDir.Mkdir(0o755, false)
	s.Require().NoError(err)

	err = emptyDir.Rmdir()
	s.Require().NoError(err)
	s.False(emptyDir.Exists())

	// Test removing a non-existent directory
	nonExistentDir := Path(filepath.Join(testDir, "nonExistentDir"))
	err = nonExistentDir.Rmdir()
	s.Require().Error(err)
	s.True(os.IsNotExist(err))

	// Test removing a non-empty directory
	nonEmptyDir := Path(filepath.Join(testDir, "nonEmptyDir"))
	err = nonEmptyDir.Mkdir(0o755, false)
	s.Require().NoError(err)
	fileInDir := nonEmptyDir.Join("file.txt")
	err = fileInDir.WriteText("test content")
	s.Require().NoError(err)

	err = nonEmptyDir.Rmdir()
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrDirectoryNotEmpty)
	s.True(nonEmptyDir.Exists())

	// Test removing a file
	filePath := Path(filepath.Join(testDir, "file.txt"))
	err = filePath.WriteText("test content")
	s.Require().NoError(err)

	err = filePath.Rmdir()
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrNotDirectory)
	s.True(filePath.Exists())

	// Test removing a symlink to a directory
	symlinkDir := Path(filepath.Join(testDir, "symlinkDir"))
	err = os.Symlink(nonEmptyDir.absPath, symlinkDir.absPath)
	s.Require().NoError(err)

	err = symlinkDir.Rmdir()
	s.Require().Error(err)
	s.Require().ErrorIs(err, ErrDirectoryNotEmpty)
	s.True(symlinkDir.Exists())
}

func (s *PathSuite) Test00Manual() {
	raw := "~/films/golang pathlib/learning.go"
	p := Path(raw)
	s.NotNil(p)
}
