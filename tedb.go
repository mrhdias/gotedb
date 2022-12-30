//
// Copyright 2022 The GoTeDB Authors. All rights reserved.
// Use of this source code is governed by a MIT License
// license that can be found in the LICENSE file.
// Last Modification: 2022-12-30 21:51:29
//

package tedb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type TEDB struct {
	Url            string
	CountryCodes   map[string]int
	CacheDir       string
	CreateCacheDir bool
	Timeout        int
	Debug          bool
}

func SplitCn(commodityCode string) []string {
	parts := []string{commodityCode[0:4]}

	if len(commodityCode) <= 4 {
		return parts
	}

	remainder := commodityCode[4:]

	even := int(len(remainder)/2) * 2
	for i := 0; i < even; i += 2 {
		parts = append(parts, remainder[i:i+1*2])
	}
	if even < len(remainder) {
		parts = append(parts, remainder[even:])
	}

	return parts
}

func (tedb TEDB) GetCnId(commodityCode string) (int, error) {

	if len(commodityCode) < 4 {
		return 0, fmt.Errorf("the commodity code %s is incorrect", commodityCode)
	}

	heading := commodityCode[0:4]
	code := strings.Join(SplitCn(commodityCode), " ")

	jsonFilename := fmt.Sprintf("%s.json", heading)

	if tedb.CacheDir != "" {
		jsonFilepath := filepath.Join(tedb.CacheDir, jsonFilename)
		if _, err := os.Stat(jsonFilepath); err == nil {
			contentBytes, err := os.ReadFile(jsonFilepath)
			if err != nil {
				return 0, err
			}

			var records []CodeRecord

			if err := json.Unmarshal([]byte(contentBytes), &records); err != nil {
				return 0, err
			}

			for _, record := range records {
				if strings.EqualFold(record.Code, code) {
					return record.ID, nil
				}
			}

			return 0, nil
		}
	}

	url := fmt.Sprintf("%s/codes/CN_CODE/%s", tedb.Url, jsonFilename)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	// req.Header.Add("User-Agent", fmt.Sprintf("%s/%s", userAgent, version))

	client := &http.Client{
		Timeout: time.Duration(time.Duration(tedb.Timeout).Seconds()),
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}

	defer resp.Body.Close()

	respContentBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// fmt.Println("Content Body:", string(respContentBytes), resp.StatusCode)

	if resp.StatusCode != 200 {
		// fmt.Println("StatusCode:", resp.StatusCode)
		return 0, fmt.Errorf("the server returned http status code %d when handling the HTTP request", resp.StatusCode)
	}

	// cache json file
	if tedb.CacheDir != "" {
		if _, err := os.Stat(tedb.CacheDir); os.IsNotExist(err) {
			if tedb.CreateCacheDir {
				if err := os.Mkdir(tedb.CacheDir, 0755); err != nil {
					return 0, err
				}
			}
		}

		if err := os.WriteFile(filepath.Join(tedb.CacheDir, jsonFilename), respContentBytes, 0644); err != nil {
			return 0, err
		}
	}

	var records []CodeRecord
	if err := json.Unmarshal([]byte(respContentBytes), &records); err != nil {
		return 0, err
	}

	for _, record := range records {
		if strings.EqualFold(record.Code, code) {
			return record.ID, nil
		}
	}

	return 0, nil
}

func (tedb TEDB) VatSearchResult(countryCodes []string,
	dateFrom, dateTo string,
	cnIds []int) ([]byte, error) {

	url := fmt.Sprintf("%s/vatSearchResult.json", tedb.Url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	// req.Header.Add("User-Agent", fmt.Sprintf("%s/%s", userAgent, version))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	values := req.URL.Query()
	for _, countryCode := range countryCodes {
		ccId, ok := tedb.CountryCodes[countryCode]
		if !ok {
			fmt.Println("Error")
		}
		values.Add("selectedMemberStates", strconv.Itoa(ccId))
	}

	values.Add("dateFrom", dateFrom)
	values.Add("dateTo", dateTo)

	for _, cnId := range cnIds {
		values.Add("selectedCnCodes", strconv.Itoa(cnId))
	}

	req.URL.RawQuery = values.Encode()

	client := &http.Client{
		Timeout: time.Duration(time.Duration(tedb.Timeout).Seconds()),
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	respContentBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if tedb.Debug {
		fmt.Println("Content Body Response:", string(respContentBytes), "Status Code:", resp.StatusCode)
	}

	if resp.StatusCode != 200 {
		// fmt.Println("StatusCode:", resp.StatusCode)
		return nil, fmt.Errorf("the server returned http status code %d when handling the HTTP request", resp.StatusCode)
	}

	return respContentBytes, nil
}

func (tedb TEDB) VatSearch(countryCodes []string,
	dateFrom, dateTo string,
	commodityCodes []string) ([]TEDBVatSearchResult, error) {

	df, err := time.Parse("2006/01/02", dateFrom)
	if err != nil {
		return nil, err
	}

	dt, err := time.Parse("2006/01/02", dateTo)
	if err != nil {
		return nil, err
	}

	if df.After(dt) {
		return nil, fmt.Errorf("date from \"%s\" is after date to \"%s\"", dateFrom, dateTo)
	}

	var cnIds []int
	for _, commodityCode := range commodityCodes {
		cnId, err := tedb.GetCnId(commodityCode)
		if err != nil {
			return nil, err
		}
		if tedb.Debug {
			fmt.Println("Commodity Code:", commodityCode, "CnId:", cnId)
		}

		cnIds = append(cnIds, cnId)
	}
	if len(cnIds) != len(commodityCodes) {
		return nil, errors.New("the number of commodity codes differs from the number of id's")
	}

	result, err := tedb.VatSearchResult(countryCodes, dateFrom, dateTo, cnIds)
	if err != nil {
		return nil, err
	}

	if strings.EqualFold(string(result), "{}") {
		return nil, errors.New("the service did not return any results for the given criteria")
	}

	var records []TEDBVatSearchResult
	if err := json.Unmarshal([]byte(result), &records); err != nil {
		return nil, err
	}

	return records, err
}

func NewVatRetrievalService(cacheDir string, createCacheDir bool, debugOption ...bool) TEDB {
	debug := false
	if len(debugOption) == 1 {
		debug = debugOption[0]
	}

	tedb := new(TEDB)

	tedb.Url = "https://ec.europa.eu/taxation_customs/tedb"
	tedb.CountryCodes = map[string]int{
		"AT": 1,
		"BE": 2,
		"BG": 3,
		"CY": 4,
		"CZ": 5,
		"DE": 6,
		"DK": 7,
		"EE": 8,
		"EL": 9,
		"ES": 10,
		"FI": 11,
		"FR": 12,
		// "UK": 13,
		"HR": 14,
		"HU": 15,
		"IE": 16,
		"IT": 17,
		"LT": 18,
		"LU": 19,
		"LV": 20,
		"MT": 21,
		"NL": 22,
		"PL": 23,
		"PT": 24,
		"RO": 25,
		"SE": 26,
		"SI": 27,
		"SK": 28,
		"XI": 30,
	}

	tedb.CacheDir = cacheDir
	tedb.CreateCacheDir = createCacheDir
	tedb.Timeout = 60
	tedb.Debug = debug

	return *tedb
}
