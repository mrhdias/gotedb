//
// Copyright 2023 The GoTeDB Authors. All rights reserved.
// Use of this source code is governed by a MIT License
// license that can be found in the LICENSE file.
// Last Modification: 2023-02-17 22:37:59
//

package tedb

import (
	"os"
	"testing"
	"time"
)

func TestTedb(t *testing.T) {

	cacheDir := "./tedb_cache"
	service := NewVatRetrievalService(
		cacheDir,
		true,
		3,
	)

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

	want := 4
	if got := len(records); got != want {
		t.Errorf("Records = %q, want %q", got, want)
	}

	if _, err := os.Stat(cacheDir); !os.IsNotExist(err) {
		if err = os.RemoveAll(cacheDir); err != nil {
			t.Fatal(err)
		}
	}

}
