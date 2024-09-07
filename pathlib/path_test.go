package pathlib

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
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
	path := filepath.Join(s.tempDir, name)
	err := os.WriteFile(path, []byte(content), 0o644)
	s.Require().NoError(err)
	return path
}

func (s *PathSuite) TestNameAndNameWithSuffix() {
	tests := []struct {
		name               string
		filePath           string
		expectedName       string
		expectedWithSuffix string
	}{
		{
			name:               "simple file",
			filePath:           "/tmp/test.txt",
			expectedName:       "test",
			expectedWithSuffix: "test.txt",
		},
		{
			name:               "file with multiple extensions",
			filePath:           "/home/user/document.tar.gz",
			expectedName:       "document.tar",
			expectedWithSuffix: "document.tar.gz",
		},
		{
			name:               "hidden file",
			filePath:           "/home/user/.config",
			expectedName:       "",
			expectedWithSuffix: ".config",
		},
		{
			name:               "directory",
			filePath:           "/var/log/",
			expectedName:       "log",
			expectedWithSuffix: "log",
		},
		{
			name:               "file without extension",
			filePath:           "/bin/bash",
			expectedName:       "bash",
			expectedWithSuffix: "bash",
		},
		{
			name:               "root directory",
			filePath:           "/",
			expectedName:       "/",
			expectedWithSuffix: "/",
		},
		{
			name:               "file with dot in name",
			filePath:           "/home/user/file.name.with.dots.txt",
			expectedName:       "file.name.with.dots",
			expectedWithSuffix: "file.name.with.dots.txt",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.filePath)
			s.Equal(tt.expectedName, file.Name, "Unexpected Name for %s", tt.filePath)
			s.Equal(tt.expectedWithSuffix, file.NameWithSuffix, "Unexpected NameWithSuffix for %s", tt.filePath)
		})
	}
}

func (s *PathSuite) TestPath() {
	// Setup: Create necessary directories and files
	tmpDir := s.T().TempDir()

	// Create /tmp/testdir/
	testDir := filepath.Join(tmpDir, "testdir")
	err := os.MkdirAll(testDir, 0755)
	s.Require().NoError(err)

	// Create /tmp/test.txt
	testFile := filepath.Join(tmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
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
				filepath:       testFile,
				Name:           "test",
				NameWithSuffix: "test.txt",
				WorkingDir:     filepath.Dir(testFile),
				BaseDir:        filepath.Base(filepath.Dir(testFile)),
				Suffix:         ".txt",
				AbsPath:        testFile,
				isDir:          false,
			},
		},
		{
			name: "directory path",
			path: testDir,
			expected: FSPath{
				filepath:       testDir,
				Name:           "testdir",
				NameWithSuffix: "testdir",
				WorkingDir:     testDir,
				BaseDir:        "testdir",
				Suffix:         "",
				AbsPath:        testDir,
				isDir:          true,
			},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			result := Path(tt.path)
			s.Equal(tt.expected, *result)
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
	s.NoError(err)
	s.DirExists(filepath.Dir(path))
}

func (s *PathSuite) TestMkDirs() {
	path := filepath.Join(s.tempDir, "new", "nested", "dir")
	file := Path(path)

	err := file.MkDirs()
	s.NoError(err)
	s.DirExists(path)
}

func (s *PathSuite) TestSetGetString() {
	path := filepath.Join(s.tempDir, "test.txt")
	file := Path(path)
	content := "Hello, World!"

	err := file.SetString(content)
	s.NoError(err)

	result, err := file.GetString()
	s.NoError(err)
	s.Equal(content, result)
}

func (s *PathSuite) TestSetGetBytes() {
	path := filepath.Join(s.tempDir, "test.bin")
	file := Path(path)
	content := []byte{0x48, 0x65, 0x6C, 0x6C, 0x6F}

	err := file.SetBytes(content)
	s.NoError(err)

	result, err := file.GetBytes()
	s.NoError(err)
	s.Equal(content, result)
}

func (s *PathSuite) TestReader() {
	content := "Hello, World!"
	path := s.createTempFile("test.txt", content)
	file := Path(path)

	reader, err := file.Reader()
	s.NoError(err)

	result, err := io.ReadAll(reader)
	s.NoError(err)
	s.Equal(content, string(result))
}

func (s *PathSuite) TestCopy() {
	content := "Hello, World!"
	srcPath := s.createTempFile("src.txt", content)
	dstPath := filepath.Join(s.tempDir, "dst.txt")

	file := Path(srcPath)
	err := file.Copy(dstPath)
	s.NoError(err)

	dstContent, err := os.ReadFile(dstPath)
	s.NoError(err)
	s.Equal(content, string(dstContent))
}

func (s *PathSuite) TestMove() {
	content := "Hello, World!"
	srcPath := s.createTempFile("src.txt", content)
	dstPath := filepath.Join(s.tempDir, "dst.txt")

	file := Path(srcPath)
	err := file.Move(dstPath)
	s.NoError(err)

	s.NoFileExists(srcPath)
	s.FileExists(dstPath)

	dstContent, err := os.ReadFile(dstPath)
	s.NoError(err)
	s.Equal(content, string(dstContent))
}

func (s *PathSuite) TestCSVGetSlices() {
	content := "a,b,c\n1,2,3\n4,5,6"
	path := s.createTempFile("test.csv", content)
	file := Path(path)

	result, err := file.CSVGetSlices()
	s.NoError(err)
	s.Equal([][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}}, result)
}

func (s *PathSuite) TestListFilesWithGlob() {
	s.createTempFile("file1.txt", "")
	s.createTempFile("file2.txt", "")
	s.createTempFile("file3.json", "")

	file := Path(s.tempDir)

	result, err := file.ListFilesWithGlob("*.txt")
	s.NoError(err)
	s.Len(result, 2)
	s.True(strings.HasSuffix(result[0], "file1.txt") || strings.HasSuffix(result[0], "file2.txt"))
	s.True(strings.HasSuffix(result[1], "file1.txt") || strings.HasSuffix(result[1], "file2.txt"))
}

func (s *PathSuite) TestGenRelativeFile() {
	tests := []struct {
		name     string
		filePath string
		newName  string
		want     string
	}{
		{
			name:     "in current folder with src file",
			filePath: "/tmp/b/93877/c/4696890.html",
			newName:  "./abc.json",
			want:     "/tmp/b/93877/c/abc.json",
		},
		{
			name:     "in current folder with src folder",
			filePath: "/tmp/b/93877/c/4696890/",
			newName:  "./abc.json",
			want:     "/tmp/b/93877/c/4696890/abc.json",
		},
		{
			name:     "in parent folder with src file",
			filePath: "/tmp/b/93877/c/4696890.html",
			newName:  "../abc.json",
			want:     "/tmp/b/93877/abc.json",
		},
		{
			name:     "in parent folder with src folder",
			filePath: "/tmp/b/93877/c/4696890/",
			newName:  "../abc.json",
			want:     "/tmp/b/93877/c/abc.json",
		},
		{
			name:     "in /tmp folder",
			filePath: "/tmp/b/",
			newName:  "../abc.json",
			want:     "/tmp/abc.json",
		},
		{
			name:     "in /tmp folder",
			filePath: "/tmp/",
			newName:  "../../abc.json",
			want:     "/abc.json",
		},
	}
	for _, tt := range tests {
		file := Path(tt.filePath)
		got := file.GenRelativeFSPath(tt.newName)
		s.Equal(tt.want, got.AbsPath, tt.name)
	}
}

func (s *PathSuite) TestGenFileWithType() {
	tests := []struct {
		name      string
		filePath  string
		newType   string
		newFolder bool
		want      string
	}{
		{
			name:     "in current folder",
			filePath: "/tmp/a/b/c.txt",
			newType:  "json",
			want:     "/tmp/a/b/c.json",
		},
		{
			name:      "in parent folder",
			filePath:  "/tmp/a/b/c.txt",
			newType:   "json",
			newFolder: true,
			want:      "/tmp/a/b_json/c.json",
		},
		{
			name:      "in top-level folder",
			filePath:  "/tmp/c.txt",
			newType:   "json",
			newFolder: true,
			want:      "/tmp/_json/c.json",
		},
		{
			name:      "in top-level folder",
			filePath:  "/tmp/c.txt",
			newType:   "yaml",
			newFolder: true,
			want:      "/tmp/_yaml/c.yaml",
		},
	}

	for _, tt := range tests {
		file := Path(tt.filePath)
		got := file.GenFilePathWithNewSuffix(tt.newType, tt.newFolder)
		s.Equal(tt.want, got.filepath, tt.name)
	}
}

func (s *PathSuite) TestGenPathInSiblingDir() {
	tests := []struct {
		name       string
		filePath   string
		folderName string
		want       string
	}{
		{
			name:       "file in subdirectory",
			filePath:   "/tmp/a/b/c.txt",
			folderName: "new_folder",
			want:       "/tmp/a/new_folder/c.txt",
		},
		{
			name:       "file in root directory",
			filePath:   "/tmp/c.txt",
			folderName: "new_folder",
			want:       "/new_folder/c.txt",
		},
		{
			name:       "directory path",
			filePath:   "/tmp/a/b/c/",
			folderName: "new_folder",
			want:       "/tmp/a/b/new_folder/c",
		},
		{
			name:       "file with no extension",
			filePath:   "/tmp/a/b/c",
			folderName: "new_folder",
			want:       "/tmp/a/new_folder/c",
		},
		{
			name:       "folder name with spaces",
			filePath:   "/tmp/a/b/c.txt",
			folderName: "new folder",
			want:       "/tmp/a/new folder/c.txt",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.filePath)
			got := file.GenPathInSiblingDir(tt.folderName)
			s.Equal(tt.want, got.AbsPath, tt.name)
			s.Equal(filepath.Base(tt.want), got.NameWithSuffix, tt.name)
		})
	}
}

func (s *PathSuite) TestParts() {
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
			expected: []string{"home", "user", "documents"},
		},
		{
			name:     "path with multiple consecutive slashes",
			path:     "/var///log/messages",
			expected: []string{"/", "var", "log", "messages"},
		},
		{
			name:     "path with dot",
			path:     "/etc/./config",
			expected: []string{"/", "etc", ".", "config"},
		},
		{
			name:     "path with double dot",
			path:     "/usr/local/../bin",
			expected: []string{"/", "usr", "local", "..", "bin"},
		},
		{
			name:     "empty path",
			path:     "",
			expected: []string{},
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
			s.Equal(tt.expected, file.Parent())
		})
	}
}

func (s *PathSuite) TestParents() {
	tests := []struct {
		name     string
		raw      string
		n        int
		want     string
		wantSame bool
	}{
		{
			name:     "zero levels up",
			raw:      "/tmp/b/93877/c/4696890",
			n:        0,
			want:     "/tmp/b/93877/c/4696890",
			wantSame: true,
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
			want: "documents",
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

			if tt.wantSame {
				s.Equal(file.AbsPath, got, "Path should remain the same for n=0")
			} else {
				s.Equal(tt.want, got, "Unexpected parent path")
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
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			file := Path(tt.filePath)
			s.Equal(tt.wantBase, file.BaseDir, "Unexpected BaseDir")
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
			s.Equal(tt.filePath, file.OriginalName())
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
			dir, name := file.SplitPath(tt.filePath)
			s.Equal(tt.wantDir, dir)
			s.Equal(tt.wantName, name)
		})
	}
}

func (s *PathSuite) TestMustSetString() {
	path := filepath.Join(s.tempDir, "test.txt")
	file := Path(path)
	content := "Hello, World!"

	file.MustSetString(content)

	result, err := os.ReadFile(path)
	s.NoError(err)
	s.Equal(content, string(result))
}

func (s *PathSuite) TestMustGetBytes() {
	content := "Hello, World!"
	path := s.createTempFile("test.txt", content)
	file := Path(path)

	result := file.MustGetBytes()
	s.Equal([]byte(content), result)
}

func (s *PathSuite) TestMustCSVGetSlices() {
	content := "a,b,c\n1,2,3\n4,5,6"
	path := s.createTempFile("test.csv", content)
	file := Path(path)

	result := file.MustCSVGetSlices()
	s.Equal([][]string{{"a", "b", "c"}, {"1", "2", "3"}, {"4", "5", "6"}}, result)
}

func (s *PathSuite) TestCreateSiblingDir() {
	tmpDir := s.T().TempDir()

	tests := []struct {
		name        string
		setup       func() *FSPath
		newDirName  string
		expectedDir string
		expectError bool
	}{
		{
			name: "Create sibling dir for file path",
			setup: func() *FSPath {
				filePath := filepath.Join(tmpDir, "a", "b", "test.txt")
				s.Require().NoError(os.MkdirAll(filepath.Dir(filePath), 0o755))
				s.Require().NoError(os.WriteFile(filePath, []byte("test"), 0o644))
				return Path(filePath)
			},
			newDirName:  "newFolder",
			expectedDir: filepath.Join(tmpDir, "a", "b", "newFolder"),
		},
		{
			name: "Create sibling dir for directory path",
			setup: func() *FSPath {
				dirPath := filepath.Join(tmpDir, "x", "y", "z")
				s.Require().NoError(os.MkdirAll(dirPath, 0o755))
				return Path(dirPath)
			},
			newDirName:  "newFolder",
			expectedDir: filepath.Join(tmpDir, "x", "y", "z", "newFolder"),
		},
		{
			name: "Create sibling dir when it already exists",
			setup: func() *FSPath {
				dirPath := filepath.Join(tmpDir, "m", "n")
				newDirPath := filepath.Join(tmpDir, "m", "n", "existingFolder")
				s.Require().NoError(os.MkdirAll(dirPath, 0o755))
				s.Require().NoError(os.MkdirAll(newDirPath, 0o755))
				return Path(dirPath)
			},
			newDirName:  "existingFolder",
			expectedDir: filepath.Join(tmpDir, "m", "n", "existingFolder"),
		},
		{
			name: "Attempt to create sibling dir with file name conflict",
			setup: func() *FSPath {
				dirPath := filepath.Join(tmpDir, "p", "q")
				conflictPath := filepath.Join(tmpDir, "p", "q", "conflict")
				s.Require().NoError(os.MkdirAll(dirPath, 0o755))
				s.Require().NoError(os.WriteFile(conflictPath, []byte("test"), 0o644))
				return Path(dirPath)
			},
			newDirName:  "conflict",
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			path := tt.setup()
			newDir, err := path.CreateSiblingDir(tt.newDirName)

			if tt.expectError {
				s.Error(err)
			} else {
				s.Require().NoError(err)
				s.Equal(tt.expectedDir, newDir.AbsPath)
				// Check if the directory exists using the file system
				info, err := os.Stat(newDir.AbsPath)
				s.Require().NoError(err)
				s.True(info.IsDir(), "Created path should be a directory")
				s.DirExists(newDir.AbsPath)
			}
		})
	}
}

func (s *PathSuite) TestCreateSiblingDirToParent() {
	tmpDir := s.T().TempDir()

	tests := []struct {
		name        string
		setup       func() *FSPath
		newDirName  string
		expectedDir string
		expectError bool
	}{
		{
			name: "Create sibling dir to parent for file path",
			setup: func() *FSPath {
				filePath := filepath.Join(tmpDir, "a", "b", "test.txt")
				s.Require().NoError(os.MkdirAll(filepath.Dir(filePath), 0o755))
				s.Require().NoError(os.WriteFile(filePath, []byte("test"), 0o644))
				return Path(filePath)
			},
			newDirName:  "newFolder",
			expectedDir: filepath.Join(tmpDir, "a", "newFolder"),
		},
		{
			name: "Create sibling dir to parent for directory path",
			setup: func() *FSPath {
				dirPath := filepath.Join(tmpDir, "x", "y", "z")
				s.Require().NoError(os.MkdirAll(dirPath, 0o755))
				return Path(dirPath)
			},
			newDirName:  "newFolder",
			expectedDir: filepath.Join(tmpDir, "x", "newFolder"),
		},
		{
			name: "Create sibling dir to parent when it already exists",
			setup: func() *FSPath {
				dirPath := filepath.Join(tmpDir, "m", "n", "o")
				newDirPath := filepath.Join(tmpDir, "m", "existingFolder")
				s.Require().NoError(os.MkdirAll(dirPath, 0o755))
				s.Require().NoError(os.MkdirAll(newDirPath, 0o755))
				return Path(dirPath)
			},
			newDirName:  "existingFolder",
			expectedDir: filepath.Join(tmpDir, "m", "existingFolder"),
		},
		{
			name: "Attempt to create sibling dir to parent with file name conflict",
			setup: func() *FSPath {
				dirPath := filepath.Join(tmpDir, "p", "q", "r")
				conflictPath := filepath.Join(tmpDir, "p", "conflict")
				s.Require().NoError(os.MkdirAll(dirPath, 0o755))
				s.Require().NoError(os.WriteFile(conflictPath, []byte("test"), 0o644))
				return Path(dirPath)
			},
			newDirName:  "conflict",
			expectError: true,
		},
		{
			name: "Attempt to create sibling dir to parent at root",
			setup: func() *FSPath {
				return Path("/")
			},
			newDirName:  "newFolder",
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			path := tt.setup()
			newDir, err := path.CreateSiblingDirToParent(tt.newDirName)

			if tt.expectError {
				s.Error(err)
			} else {
				s.Require().NoError(err)
				s.Equal(tt.expectedDir, newDir.AbsPath)
				info, err := os.Stat(newDir.AbsPath)
				s.Require().NoError(err)
				s.True(info.IsDir(), "Created path should be a directory")
				s.DirExists(newDir.AbsPath)
			}
		})
	}
}

func (s *PathSuite) TestTSVGetSlices() {
	content := "a\tb\tc\n1\t2\t3\n4\t5\t6"
	path := s.createTempFile("test.tsv", content)
	file := Path(path)

	result, err := file.TSVGetSlices()
	s.NoError(err)
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
		file.e(nil, errors.New("test error"))
	})
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
			s.NoError(err)
			s.Len(result, tt.expected)
		})
	}
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
			s.NoError(err)
			s.Equal(tt.expected, result)
		})
	}
}
