# ODS decoder for Go

This package only convers my own needs. It decodes a spreadsheet into a [][]string nothing more than that.
It does handle table cells with extra formating and stylning in the sense that it just strips them away.

Opening a ODF/ODS file is decoupled from the actual decoding so you can inject what ever middle reader you need (ex. different character encoding).

```go
package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/ksev/ods"
)

func main() {
	reader, err := ods.OpenReader("file.ods")
	if err != nil {
		log.Fatal(err)
	}
	defer reader.Close()

	table, err := ods.Decode(reader)
	if err != nil {
		log.Fatal(err)
	}

	for _, row := range table {
		fmt.Println(strings.Join(row, "|"))
	}
}
```
