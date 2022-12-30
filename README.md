# GoTEDB
VAT Search in TEDB (Taxes in Europe Database) using Golang

## Installation
```
go get github.com/mrhdias/gotedb
```
## Example
```go
package main

import (
	"fmt"
	"log"
	"time"

	tedb "github.com/mrhdias/gotedb"
)

func main() {

	service := tedb.NewVatRetrievalService("./tedb_cache", true)

	currentTime := time.Now()

	records, err := service.VatSearch([]string{"ES"},
		currentTime.AddDate(0, 0, -1).Format("2006/01/02"),
		currentTime.Format("2006/01/02"),
		[]string{"33049900"})

	if err != nil {
		log.Fatalln(err)
	}

	for _, record := range records {
		fmt.Println(record.MemberState.DefaultCountryCode)
		fmt.Println(record.Type)
		fmt.Println(record.Rate.Key, record.Rate.Value, record.Comments)
	}
}
```
