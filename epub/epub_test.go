package epub

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/suite"
	"github.com/ungerik/go-dry"
)

type EBookSuite struct {
	suite.Suite
}

func TestEBook(t *testing.T) {
	suite.Run(t, new(EBookSuite))
}

func (s *EBookSuite) SetupSuite() {
}

func (s *EBookSuite) TearDownSuite() {
}

func (s *EBookSuite) TestNewEBook() {
	root := "/tmp/shuba/txt/16927"
	files, err := dry.ListDirFiles(root)
	s.Nil(err)

	arr := []int{}
	for _, fi := range files {
		arr = append(arr, cast.ToInt(strings.Split(fi, ".")[0]))
	}

	sort.Ints(arr)

	files = []string{}
	for _, v := range arr {
		files = append(files, fmt.Sprintf("%s/%d.html", root, v))
	}

	NewEBookWithFiles("xx", "zxy", files)
}
