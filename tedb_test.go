//
// Copyright 2024 The GoTeDB Authors. All rights reserved.
// Use of this source code is governed by a MIT License
// license that can be found in the LICENSE file.
// Last Modification: 2024-01-16 12:20:43
//

package tedb

import (
	"testing"
	"time"
)

func TestTedb(t *testing.T) {

	service := NewVatRetrievalService()

	currentTime := time.Now()

	criteria := Criteria{
		CountryCodes:   []string{"ES"},
		DateFrom:       currentTime.AddDate(0, 0, -1).Format("2006/01/02"),
		DateTo:         currentTime.Format("2006/01/02"),
		Categories:     []string{"foodstuffs"},
		CommodityCodes: []string{"33049900", "0402 29 11"},
	}
	records, err := service.VatSearch(criteria)

	if err != nil {
		t.Fatal(err)
	}

	want := 2
	if got := len(records); got != want {
		t.Errorf("Records = %d, want %d", got, want)
	}

	want = 2
	if got := len(records[0].Rates); got != want {
		t.Errorf("Rates = %d, want %d", got, want)
	}
}
