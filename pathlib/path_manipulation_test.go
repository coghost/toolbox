package pathlib

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/suite"
)

type PathManipulationSuite struct {
	suite.Suite
}

func TestPathManipulation(t *testing.T) {
	suite.Run(t, new(PathManipulationSuite))
}

func (s *PathManipulationSuite) SetupSuite() {
}

func (s *PathManipulationSuite) TearDownSuite() {
}

func (s *PathManipulationSuite) TestWithSuffixAndSuffixedParentDir() {
	tests := []struct {
		name      string
		filePath  string
		newSuffix string
		want      string
		wantNil   bool
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
			want:      "/tmp/a/b",
			wantNil:   true,
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

			if tt.wantNil {
				s.Nil(got, tt.filePath)
				return
			}

			s.Equal(tt.want, got.absPath, "For input: %s", tt.filePath)

			// Additional checks
			s.Equal(filepath.Base(tt.want), got.Name)
			s.Equal(filepath.Dir(tt.want), got.Parent().absPath)

			// Check that the original path hasn't changed
			s.NotEqual(got.absPath, file.absPath, "Original path should not be modified")

			// Check that the new file has the correct suffix
			if tt.newSuffix != "" && !got.IsDir() {
				s.Equal("."+strings.TrimPrefix(tt.newSuffix, "."), filepath.Ext(got.Name))
			} else {
				s.Equal("", filepath.Ext(got.Name))
			}
		})
	}
}

func (s *PathManipulationSuite) TestWithReplacedDirAndSuffix() {
	tests := []struct {
		path      string
		dirName   string
		newSuffix string
		want      string
		wantNil   bool
	}{
		// Basic cases
		{
			path:      "/path/to/file.txt",
			dirName:   "data",
			newSuffix: "json",
			want:      "/path/data/file.json",
		},
		{
			path:      "/path/to/file.txt",
			dirName:   "output",
			newSuffix: ".pdf",
			want:      "/path/output/file.pdf",
		},

		// Root directory cases
		{
			path:      "/file.txt",
			dirName:   "data",
			newSuffix: "json",
			want:      "/data/file.json",
		},
		{
			path:      "/file.txt",
			dirName:   "docs",
			newSuffix: ".md",
			want:      "/docs/file.md",
		},

		// Deep path cases
		{
			path:      "/usr/local/share/file.txt",
			dirName:   "data",
			newSuffix: "json",
			want:      "/usr/local/data/file.json",
		},

		// Cases with no extension
		{
			path:      "/path/to/file",
			dirName:   "data",
			newSuffix: "txt",
			want:      "/path/data/file.txt",
		},

		// Cases with empty suffix
		{
			path:      "/path/to/file.txt",
			dirName:   "data",
			newSuffix: "",
			want:      "/path/data/file",
		},

		// Cases with special characters in directory name
		{
			path:      "/path/to/file.txt",
			dirName:   "data-2023",
			newSuffix: "json",
			want:      "/path/data-2023/file.json",
		},

		// Cases with multiple dots in filename
		{
			path:      "/path/to/file.name.txt",
			dirName:   "data",
			newSuffix: "json",
			want:      "/path/data/file.name.json",
		},

		// Directory paths (should remain unchanged)
		{
			path:      "/path/to/dir/",
			dirName:   "data",
			newSuffix: "json",
			want:      "/path/to/dir",
			wantNil:   true,
		},
		{
			path:      "/",
			dirName:   "data",
			newSuffix: "txt",
			want:      "/",
			wantNil:   true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.path, func() {
			p := Path(tt.path)
			got := p.WithReplacedDirAndSuffix(tt.dirName, tt.newSuffix)

			if tt.wantNil {
				s.Nil(got, tt.path)
				return
			}

			s.Equal(tt.want, got.absPath,
				"WithReplacedDirAndSuffix(%s, %s) for path %s",
				tt.dirName, tt.newSuffix, tt.path)
		})
	}
}

func (s *PathManipulationSuite) TestLastNSegments() {
	tests := []struct {
		path     string
		n        int
		expected string
	}{
		// File paths
		{"/home/user/documents/file.txt", 0, "file.txt"},
		{"/home/user/documents/file.txt", 1, "documents/file.txt"},
		{"/home/user/documents/file.txt", 2, "user/documents/file.txt"},
		{"/home/user/documents/file.txt", 3, "home/user/documents/file.txt"},
		{"/home/user/documents/file.txt", 4, "/home/user/documents/file.txt"},
		{"/home/user/documents/file.txt", 5, "/home/user/documents/file.txt"},

		// Directory paths
		{"/home/user/documents/", 0, "documents"},
		{"/home/user/documents/", 1, "user/documents"},
		{"/home/user/documents/", 2, "home/user/documents"},
		{"/home/user/documents/", 3, "/home/user/documents"},
		{"/home/user/documents/", 4, "/home/user/documents"},

		// Root directory
		{"/", 0, "/"},
		{"/", 1, "/"},
		{"/", 2, "/"},

		// Relative paths
		{"user/documents/file.txt", 0, "file.txt"},
		{"user/documents/file.txt", 1, "documents/file.txt"},
		{"user/documents/file.txt", 2, "user/documents/file.txt"},
		{"user/documents/file.txt", 3, "user/documents/file.txt"},

		// Edge cases
		{"", 0, ""},
		{"", 1, ""},
		{"file.txt", 0, "file.txt"},
		{"file.txt", 1, "file.txt"},
		{"file.txt", 2, "file.txt"},
	}

	for _, tt := range tests {
		s.Run(tt.path, func() {
			p := Path(tt.path)
			result := p.LastNSegments(tt.n)
			s.Equal(tt.expected, result, "LastNSegments(%d) for path %s", tt.n, tt.path)
		})
	}
}

func (s *PathManipulationSuite) TestLastSegment() {
	tests := []struct {
		path     string
		expected string
	}{
		{"/home/user/documents/file.txt", "documents/file.txt"},
		{"/home/user/documents/", "user/documents"},
		{"/home/user/", "home/user"},
		{"/", "/"},
		{"user/documents/file.txt", "documents/file.txt"},
		{"file.txt", "file.txt"},
		{"", ""},
	}

	for _, tt := range tests {
		s.Run(tt.path, func() {
			p := Path(tt.path)
			result := p.LastSegment()
			s.Equal(tt.expected, result, "LastSegment() for path %s", tt.path)
		})
	}
}
