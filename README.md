# GoTEDB
VAT Search in TEDB ([Taxes in Europe Database v4](https://ec.europa.eu/taxation_customs/tedb/#/vat-search)) using Golang

The query of goods commodity code can be done in the [TARIC Consultation](https://ec.europa.eu/taxation_customs/dds2/taric/taric_consultation.jsp) page.

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

    service := tedb.NewVatRetrievalService()

    currentTime := time.Now()
    
    criteria := tedb.Criteria{
        CountryCodes:   []string{"ES"},
        DateFrom:       currentTime.AddDate(0, 0, -1).Format("2006/01/02"), // Optional - default today date -1 day
        DateTo:         currentTime.Format("2006/01/02"),                   // Optional - default today date
        Categories:     []string{"foodstuffs"},             // Category(ies) - Optional
        CommodityCodes: []string{"33049900", "0402 29 11"}, // Search by CN Codes (goods) - Optional
    }
    records, err := service.VatSearch(criteria)

    if err != nil {
        log.Fatalln(err)
    }

    for _, record := range records {
        fmt.Println("Country IsoCode:", record.IsoCode,
            "Type:", record.Type,
            "Rate:", func() float64 {
                if len(record.Rates) == 0 {
                    return -1.00
                }
                for _, rate := range record.Rates {
                    if strings.Contains(rate.Comments, "temporary subject to a 0% VAT rate") {
                        return 0.00
                    }
                }
                return record.Rates[0].Value
            }())
    }
}
```
