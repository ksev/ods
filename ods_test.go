package ods

import (
	"testing"
)

func TestOpenReader(test *testing.T) {
	file, err := OpenReader("./test.ods")
	if err != nil {
		test.Fatal(err)
	}
	if file == nil {
		test.Fatal("file is nil")
	}
}

func TestDecoding(test *testing.T) {
	file, _ := OpenReader("./test.ods")

	document, err := Decode(file)
	if err != nil {
		test.Fatal(err)
	}
	if document == nil {
		test.Fatal(err)
	}

	// Make sure all columns are equally wide
	width := len(document[0])

	for i, row := range document {
		if len(row) != width {
			test.Fatalf("Diverging column width on row: %d, number of column: %d, expected: %d", i+1, len(row), width)
		}
	}
}
