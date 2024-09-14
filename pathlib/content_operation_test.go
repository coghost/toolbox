package pathlib

import (
	"os"
	"path/filepath"
)

func (s *PathSuite) TestWriteText() {
	path := filepath.Join(s.tempDir, "writetext.txt")
	file := Path(path)

	err := file.WriteText(_testContent)
	s.Require().NoError(err)

	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Equal(_testContent, string(content))
}

func (s *PathSuite) TestWriteBytes() {
	path := filepath.Join(s.tempDir, "writebytes.txt")
	file := Path(path)
	content := []byte(_testContent)

	err := file.WriteBytes(content)
	s.Require().NoError(err)

	readContent, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Equal(content, readContent)
}

func (s *PathSuite) TestReadText() {
	path := s.createTempFile("readtext.txt", _testContent)
	file := Path(path)

	content, err := file.ReadText()
	s.Require().NoError(err)
	s.Equal(_testContent, content)
}

func (s *PathSuite) TestReadBytes() {
	path := s.createTempFile("readbytes.txt", _testContent)
	file := Path(path)

	content, err := file.ReadBytes()
	s.Require().NoError(err)
	s.Equal([]byte(_testContent), content)
}

func (s *PathSuite) TestMustWriteText() {
	path := filepath.Join(s.tempDir, "mustwritetext.txt")
	file := Path(path)

	s.NotPanics(func() {
		file.MustWriteText(_testContent)
	})

	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Equal(_testContent, string(content))
}

func (s *PathSuite) TestMustWriteBytes() {
	path := filepath.Join(s.tempDir, "mustwritebytes.txt")
	file := Path(path)
	content := []byte(_testContent)

	s.NotPanics(func() {
		file.MustWriteBytes(content)
	})

	readContent, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Equal(content, readContent)
}

func (s *PathSuite) TestMustReadText() {
	path := s.createTempFile("mustreadtext.txt", _testContent)
	file := Path(path)

	content := file.MustReadText()
	s.Equal(_testContent, content)
}

func (s *PathSuite) TestMustReadBytes() {
	path := s.createTempFile("mustreadbytes.txt", _testContent)
	file := Path(path)

	content := file.MustReadBytes()
	s.Equal([]byte(_testContent), content)
}

func (s *PathSuite) TestAppendText() {
	path := s.createTempFile("appendtext.txt", _testContent)
	file := Path(path)

	additionalContent := "Additional content"
	err := file.AppendText(additionalContent)
	s.Require().NoError(err)

	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Equal(_testContent+additionalContent, string(content))
}

func (s *PathSuite) TestAppendBytes() {
	path := s.createTempFile("appendbytes.txt", _testContent)
	file := Path(path)

	additionalContent := []byte("Additional content")
	err := file.AppendBytes(additionalContent)
	s.Require().NoError(err)

	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Equal(append([]byte(_testContent), additionalContent...), content)
}

func (s *PathSuite) TestMustAppendText() {
	path := s.createTempFile("mustappendtext.txt", _testContent)
	file := Path(path)

	additionalContent := "Additional content"
	s.NotPanics(func() {
		file.MustAppendText(additionalContent)
	})

	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Equal(_testContent+additionalContent, string(content))
}

func (s *PathSuite) TestMustAppendBytes() {
	path := s.createTempFile("mustappendbytes.txt", _testContent)
	file := Path(path)

	additionalContent := []byte("Additional content")
	s.NotPanics(func() {
		file.MustAppendBytes(additionalContent)
	})

	content, err := os.ReadFile(path)
	s.Require().NoError(err)
	s.Equal(append([]byte(_testContent), additionalContent...), content)
}

func (s *PathSuite) TestCalculateMD5() {
	// Create a temporary file with known content
	tempDir := s.T().TempDir()
	filePath := Path(tempDir).Join("test_file.txt")
	content := "Hello, World!"
	err := filePath.WriteText(content)
	s.Require().NoError(err)

	// Calculate MD5
	md5hash, err := filePath.GetMD5()
	s.Require().NoError(err)

	// Expected MD5 hash for "Hello, World!"
	expectedMD5 := "65a8e27d8879283831b664bd8b7f0ad4"
	s.Equal(expectedMD5, md5hash)

	// Test with non-existent file
	nonExistentFile := Path(tempDir).Join("non_existent.txt")
	_, err = nonExistentFile.GetMD5()
	s.Require().Error(err)

	// Test with directory
	dirPath := Path(tempDir)
	_, err = dirPath.GetMD5()
	s.Require().Error(err)
}
