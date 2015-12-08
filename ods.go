// Read Open Document Format spreadsheets
package ods

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"archive/zip"
	"encoding/xml"
	"io/ioutil"
)

const (
	MIMETYPE = "application/vnd.oasis.opendocument.spreadsheet"
)

type readCloserImpl struct {
	io.Reader
	io.Closer
}

func Decode(reader io.Reader) ([][]string, error) {
	decoder := xml.NewDecoder(reader)

	// Just some state for our simple parser
	insheet := false
	incolumn := false
	repeat := 1

	var cell []string
	var row []string
	var err error
	result := make([][]string, 0)

	for {
		t, _ := decoder.Token()
		if t == nil {
			break
		}
		switch token := t.(type) {
		case xml.StartElement:
			switch {
			// Make sure the table we are parsning is in a spreadsheet
			case token.Name.Local == "spreadsheet":
				insheet = true
			case token.Name.Local == "table-cell" && insheet:
				incolumn = true
				cell = make([]string, 0)
				repeat = 1
				// If content repeats to adjecent cells the formats compact them to one cell
				for _, attr := range token.Attr {
					if attr.Name.Local == "number-columns-repeated" {
						repeat, err = strconv.Atoi(attr.Value)
						if err != nil {
							return nil, err
						}
					}
				}
			case token.Name.Local == "table-row" && insheet:
				row = make([]string, 0)
			case incolumn:
				str := ""
				decoder.DecodeElement(&str, &token)
				if str != "" {
					cell = append(cell, str)
				}
			}
		case xml.EndElement:
			switch {
			case token.Name.Local == "spreadsheet":
				insheet = false
			case token.Name.Local == "table-cell" && insheet:
				incolumn = false
				if len(cell) > 0 {
					for i := 0; i < repeat; i += 1 {
						row = append(row, strings.Join(cell, " "))
					}
				}
			case token.Name.Local == "table-row" && insheet:
				if len(row) > 0 {
					result = append(result, row)
				}
			}
		}
	}

	return result, nil
}

// Opens a reader for the content part of ODF.
// Takes a string path to specified file.
// Some simple sanity checks are made to make sure the file is a ODS spreadsheet.
func OpenReader(path string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	reader, err := NewReader(file)
	if err != nil {
		return nil, err
	}

	return readCloserImpl{reader, file}, nil
}

// NewReader takes an exsisting reader and reads the spreadsheet out of it.
// It's important to note that this will exhaust the underlying Reader and is NOT streaming.
// This happens because ODF/ODS files are ZIP files and they require random read access.
func NewReader(rdr io.Reader) (io.Reader, error) {
	data, err := ioutil.ReadAll(rdr)
	if err != nil {
		return nil, err
	}

	zip, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return nil, err
	}

	mimeType, err := readMimeType(zip)
	if err != nil {
		return nil, errors.New("Could not read mimetype")
	}
	if mimeType != MIMETYPE {
		return nil, errors.New(fmt.Sprintf("Unexpected mimetype. found: %s, expected: ", mimeType, MIMETYPE))
	}

	content := findFile("content.xml", zip.File)
	if content == nil {
		return nil, errors.New("content.xml not found. The file might be corrupted.")
	}

	reader, err := content.Open()
	if err != nil {
		return nil, err
	}

	return reader, nil
}

// Small helper function to extract the mimetype of the ODF file.
func readMimeType(zip *zip.Reader) (string, error) {
	file := findFile("mimetype", zip.File)
	if file == nil {
		return "", errors.New("No mimetype present in input file")
	}

	reader, err := file.Open()
	if err != nil {
		return "", err
	}

	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// Find a specific file in the zip package.
// Returns a nil pointer of no file is found.
func findFile(needle string, haystack []*zip.File) *zip.File {
	for _, file := range haystack {
		if file.Name == needle {
			return file
		}
	}
	return nil
}
