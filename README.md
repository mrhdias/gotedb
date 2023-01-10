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

    fmt.Println("Country Codes:", tedb.CountryCodes)
    fmt.Println("Categories:", tedb.Categories)

    service := tedb.NewVatRetrievalService("./tedb_cache", true)

    currentTime := time.Now()
    
    criteria := tedb.Criteria{
        CountryCodes:   []string{"ES"},
        DateFrom:       currentTime.AddDate(0, 0, -1).Format("2006/01/02"), // Optional - default today date -1 day
        DateTo:         currentTime.Format("2006/01/02"),                   // Optional - default today date
        Categories:     []string{"foodstuffs"},  // Category(ies) - Optional
        CommodityCodes: []string{"33049900"},    // CN Code(s) - Optional
    }
    records, err := service.VatSearch(criteria)

    if err != nil {
        log.Fatalln(err)
    }

    for _, record := range records {
        fmt.Println("Country Code:", record.MemberState.DefaultCountryCode,
            "Type:", record.Type,
            "Rate:", record.Rate.Value,
            "Comments:", record.Comments)
    }
}
```
