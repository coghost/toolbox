package epub

import (
	"fmt"

	"github.com/go-shiori/go-epub"
	"github.com/ungerik/go-dry"
)

type EBook struct {
	Name   string
	Author string

	Epub *epub.Epub
}

func NewEBookWithFiles(author, filename string, files []string) (string, error) {
	e, err := NewEBook(filename, author)
	if err != nil {
		return "", err
	}

	if err := e.AddFiles(files); err != nil {
		return "", err
	}

	dst := filename + ".epub"
	if err := e.Save(dst); err != nil {
		dry.PanicIfErr(err)
	}

	return dst, nil
}

func NewEBook(bookname string, author string) (*EBook, error) {
	e, err := epub.NewEpub(bookname)
	if err != nil {
		return nil, err
	}

	e.SetAuthor(author)

	return &EBook{
		Name:   bookname,
		Author: author,
		Epub:   e,
	}, nil
}

// AddFiles
//
//	file format:
//	 - first line: must be the chapter name
//	 - other lines: are treated as body content.
func (c *EBook) AddFiles(files []string) error {
	for _, file := range files {
		lines, err := dry.FileGetLines(file)
		if err != nil {
			return err
		}

		if err := c.AddSectionByFile(lines[0], lines[1:]); err != nil {
			return err
		}
	}

	return nil
}

func (c *EBook) AddSectionByFile(header string, paragraphs []string) error {
	body := fmt.Sprintf("<h1>%s</h1>", header)

	for _, l := range paragraphs {
		body += fmt.Sprintf("<p>%s</p>", l)
	}

	_, err := c.Epub.AddSection(body, header, "", "")
	if err != nil {
		return fmt.Errorf("cannot add section: %w", err)
	}

	return nil
}

func (c *EBook) Save(filename string) error {
	return c.Epub.Write(filename)
}
