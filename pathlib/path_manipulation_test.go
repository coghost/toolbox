package pathlib

import (
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
