//
// Copyright 2023 The GoTeDB Authors. All rights reserved.
// Use of this source code is governed by a MIT License
// license that can be found in the LICENSE file.
// Last Modification: 2023-01-10 12:02:46
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
	Categories     map[string]int
	CacheDir       string
	CreateCacheDir bool
	Timeout        int
	Debug          bool
}

type Criteria struct {
	CountryCodes   []string
	DateFrom       string
	DateTo         string
	Categories     []string
	CommodityCodes []string
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

func (tedb TEDB) VatSearchResult(criteria Criteria) ([]byte, error) {

	url := fmt.Sprintf("%s/vatSearchResult.json", tedb.Url)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return nil, err
	}

	// req.Header.Add("User-Agent", fmt.Sprintf("%s/%s", userAgent, version))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded;charset=UTF-8")

	values := req.URL.Query()
	for _, countryCode := range criteria.CountryCodes {
		ccId, ok := tedb.CountryCodes[countryCode]
		if !ok {
			return nil, fmt.Errorf("the country code \"%s\" is invalid", countryCode)
		}
		values.Add("selectedMemberStates", strconv.Itoa(ccId))
	}

	values.Add("dateFrom", criteria.DateFrom)
	values.Add("dateTo", criteria.DateTo)

	for _, category := range criteria.Categories {
		if id, ok := tedb.Categories[category]; ok {
			values.Add("selectedCategories", strconv.Itoa(id))
		}
	}

	for _, commodityCode := range criteria.CommodityCodes {
		cnId, err := tedb.GetCnId(commodityCode)
		if err != nil {
			return nil, err
		}
		if tedb.Debug {
			fmt.Println("Commodity Code:", commodityCode, "CnId:", cnId)
		}
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

func (tedb TEDB) VatSearch(criteria Criteria) ([]TEDBVatSearchResult, error) {

	if criteria.DateTo == "" {
		currentTime := time.Now()
		criteria.DateTo = currentTime.Format("2006/01/02")
	}

	dt, err := time.Parse("2006/01/02", criteria.DateTo)
	if err != nil {
		return nil, err
	}

	if criteria.DateFrom == "" {
		criteria.DateFrom = dt.AddDate(0, 0, -1).Format("2006/01/02")
	}

	df, err := time.Parse("2006/01/02", criteria.DateFrom)
	if err != nil {
		return nil, err
	}

	if df.After(dt) {
		return nil, fmt.Errorf("date from \"%s\" is after date to \"%s\"",
			criteria.DateFrom, criteria.DateTo)
	}

	result, err := tedb.VatSearchResult(criteria)
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

	tedb.Categories = map[string]int{
		"100_years_old":           288,
		"accommodation":           262,
		"agricultural_production": 261,
		"bicycles_repair":         270,
		"broadcasting_services":   256,
		"ceramics":                283,
		"children_car_seats":      250,
		"clothing_repair":         272,
		"cultural_events":         255,
		"domestic_care":           273,
		"enamels":                 284,
		"foodstuffs":              246,
		"hairdressing":            274,
		"housing_provision":       258,
		"impressions":             279,
		"loan_libraries":          252,
		"medical_care":            268,
		"medical_equipment":       249,
		"newspapers":              253,
		"parking":                 572,
		"periodicals":             254,
		"pharmaceutical_products": 248,
		"photographs":             285,
		"pictures":                278,
		"postage":                 286,
		"private_dwellings":       259,
		"region":                  574,
		"restaurant":              263,
		"sculpture_casts":         281,
		"sculptures":              280,
		"shoes_repair":            271,
		"social_wellbeing":        266,
		"sporting_events":         264,
		"sporting_facilities":     265,
		"street_cleaning":         269,
		"super_temporary":         571,
		"supply_electricity":      276,
		"supply_gas":              275,
		"supply_heating":          277,
		"supply_water":            247,
		"tapestries":              282,
		"temporary":               573,
		"transport_passengers":    251,
		"undertakers_services":    267,
		"window_cleaning":         260,
		"writers_services":        257,
		"zero_rate":               570,
		"zero_reduced_rate":       569,
		"zoological":              287,
	}

	tedb.CacheDir = cacheDir
	tedb.CreateCacheDir = createCacheDir
	tedb.Timeout = 60
	tedb.Debug = debug

	return *tedb
}
