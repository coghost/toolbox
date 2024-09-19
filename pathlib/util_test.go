package pathlib

import (
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/suite"
)

type UtilSuite struct {
	suite.Suite
}

func TestUtil(t *testing.T) {
	suite.Run(t, new(UtilSuite))
}

func (s *UtilSuite) SetupSuite() {
}

func (s *UtilSuite) TearDownSuite() {
}

func (s *UtilSuite) TestHome() {
	home, err := Home()
	s.NoError(err)
	s.NotNil(home)

	userHome, err := os.UserHomeDir()
	s.NoError(err)
	s.Equal(userHome, home.absPath)
}

func (s *UtilSuite) TestCwd() {
	cwd, err := Cwd()
	s.NoError(err)
	s.NotNil(cwd)

	expected, err := os.Getwd()
	s.NoError(err)
	s.Equal(expected, cwd.absPath)
}

func (s *UtilSuite) TestExpandUser() {
	currentUser, err := user.Current()
	s.Require().NoError(err)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"home directory", "~", currentUser.HomeDir},
		{"home subdirectory", "~/documents", filepath.Join(currentUser.HomeDir, "documents")},
		{"current user home", "~" + currentUser.Username, currentUser.HomeDir},
		{"current user subdirectory", "~" + currentUser.Username + "/documents", filepath.Join(currentUser.HomeDir, "documents")},
		{"no expansion needed", "/tmp/file.txt", "/tmp/file.txt"},
		{"tilde in middle", "/tmp/~file.txt", "/tmp/~file.txt"},
		{"environment variable (not expanded)", "~/documents/$USER", filepath.Join(currentUser.HomeDir, "documents/$USER")},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			expanded := ExpandUser(tt.path)
			s.Equal(tt.expected, expanded)
		})
	}
}

func (s *UtilSuite) TestExpand() {
	// Save the original environment
	origHome := os.Getenv("HOME")
	origUser := os.Getenv("USER")
	defer func() {
		os.Setenv("HOME", origHome)
		os.Setenv("USER", origUser)
	}()

	// Set up test environment variables
	os.Setenv("HOME", "/home/testuser")
	os.Setenv("USER", "testuser")
	os.Setenv("CUSTOM_VAR", "custom_value")

	currentUser, err := user.Current()
	s.Require().NoError(err)

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"empty path", "", ""},
		{"home directory", "~", "/home/testuser"},
		{"home subdirectory", "~/documents", "/home/testuser/documents"},
		{"current user home", "~" + currentUser.Username, currentUser.HomeDir},
		{"current user subdirectory", "~" + currentUser.Username + "/documents", filepath.Join(currentUser.HomeDir, "documents")},
		{"environment variable", "$HOME/documents", "/home/testuser/documents"},
		{"multiple environment variables", "$HOME/$USER", "/home/testuser/testuser"},
		{"custom environment variable", "$CUSTOM_VAR", "custom_value"},
		{"mixed expansion", "~/documents/$USER", "/home/testuser/documents/testuser"},
		{"no expansion needed", "/tmp/file.txt", "/tmp/file.txt"},
		{"partial expansion", "/tmp/$USER/file.txt", "/tmp/testuser/file.txt"},
		{"expansion in middle", "/tmp/$USER/file.txt~", "/tmp/testuser/file.txt~"},
		{"multiple consecutive slashes", "//tmp///file.txt", "//tmp///file.txt"},
		{"invalid environment variable", "$NONEXISTENT", ""},
		{"escaped dollar sign", "\\$HOME", "$HOME"},
		{"partially escaped", "\\$HOME/$USER", "$HOME/testuser"},
		{"double escaped", "\\\\$HOME", "\\$HOME"},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			expanded := Expand(tt.path)
			s.Equal(tt.expected, expanded)
		})
	}
}
