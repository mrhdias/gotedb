//
// Copyright 2024 The GoTeDB Authors. All rights reserved.
// Use of this source code is governed by a MIT License
// license that can be found in the LICENSE file.
// Last Modification: 2024-04-12 18:31:33
//

package tedb

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type TEDB struct {
	Url     string
	Timeout int
	Debug   bool
}

type Criteria struct {
	CountryCodes   []string
	DateFrom       string
	DateTo         string
	Categories     []string
	CommodityCodes []string
}

var CountryCodes = map[string]int{
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

var Categories = map[string]int{
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

type void struct{} // to make a set like in python

func findDuplicates(elements []string) error {

	if len(elements) < 2 {
		return nil
	}

	set := map[string]void{}
	for _, element := range elements {
		e := strings.ToLower(strings.ReplaceAll(element, " ", ""))
		if _, ok := set[e]; ok {
			return fmt.Errorf("the element \"%s\" is duplicted", element)
		}
		set[e] = void{}
	}

	return nil
}

func SplitCn(commodityCode string) ([]string, error) {
	// https://ec.europa.eu/taxation_customs/tedb/codes/CN_CODE/30.json
	// chapter len(commodityCode) == 2
	// heading len(commodityCode) == 4

	if commodityCode == "" {
		return nil, errors.New("the commodity code is empty")
	}

	cc := strings.ReplaceAll(commodityCode, " ", "")

	if len(cc) < 2 || len(cc) > 8 || len(cc)%2 != 0 {
		return nil, fmt.Errorf("the commodity code \"%s\" is incorrect", commodityCode)
	}

	if _, err := strconv.Atoi(cc); err != nil {
		return nil,
			fmt.Errorf("the commodity code \"%s\" is incorrect, only numbers and separator spaces are allowed",
				commodityCode)
	}

	parts := []string{}

	if len(cc) <= 4 {
		// return only chapter if len(cc) == 2 heading or heading if len(cc) == 4
		return append(parts, cc), nil
	}

	parts = append(parts, cc[0:4])

	remainder := cc[4:]

	even := int(len(remainder)/2) * 2
	for i := 0; i < even; i += 2 {
		parts = append(parts, remainder[i:i+1*2])
	}
	if even < len(remainder) {
		parts = append(parts, remainder[even:])
	}

	if strings.ContainsAny(commodityCode, " ") &&
		!strings.EqualFold(strings.Join(parts, " "), commodityCode) {
		return nil, fmt.Errorf("the commodity code \"%s\" is not well formatted", commodityCode)
	}

	return parts, nil
}

func (tedb *TEDB) VatSearchQuery(criteria Criteria) (*TEDBsearchResult, error) {

	// "https://ec.europa.eu/taxation_customs/tedb/rest-api/vatSearch"
	url := fmt.Sprintf("%s/rest-api/vatSearch", tedb.Url)

	selectedMemberStates := []string{}
	for _, countryCode := range criteria.CountryCodes {
		ccId, ok := CountryCodes[countryCode]
		if !ok {
			return nil, fmt.Errorf("the country code \"%s\" is invalid", countryCode)
		}
		selectedMemberStates = append(selectedMemberStates, strconv.Itoa(ccId))
	}

	selectedCategories := []string{}
	for _, category := range criteria.Categories {
		categoryCode, ok := Categories[category]
		if !ok {
			return nil, fmt.Errorf("the category \"%s\" is invalid", category)
		}
		selectedCategories = append(selectedCategories, strconv.Itoa(categoryCode))
	}

	selectedCnCodes := []string{}
	for _, commodityCode := range criteria.CommodityCodes {
		selectedCnCode, err := SplitCn(commodityCode)
		if err != nil {
			return nil, err
		}
		selectedCnCodes = append(selectedCnCodes, strings.Join(selectedCnCode, " "))
	}

	searchForm := TEDBsearchForm{
		SelectedMemberStates: selectedMemberStates,
		DateFrom:             criteria.DateFrom,
		DateTo:               criteria.DateTo,
		SelectedCategories:   selectedCategories,
		SelectedCnCodes:      selectedCnCodes,
		SelectedCpaCodes:     []string{},
	}

	tedbSearch := TEDBSearch{
		SearchForm: searchForm,
	}

	jsonPayload, err := json.Marshal(tedbSearch)
	if err != nil {
		return nil, err
	}

	if tedb.Debug {
		fmt.Println("Json Payload:", string(jsonPayload))
	}

	jsonBody := strings.NewReader(string(jsonPayload))

	req, err := http.NewRequest(http.MethodPost, url, jsonBody)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")

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

	if resp.StatusCode != http.StatusOK {
		// fmt.Println("StatusCode:", resp.StatusCode)
		return nil, fmt.Errorf("the server returned http status code %d when handling the HTTP request", resp.StatusCode)
	}

	if strings.EqualFold(string(respContentBytes), "{}") {
		return nil, errors.New("the service did not return any results for the given criteria")
	}

	var records TEDBsearchResult
	if err := json.Unmarshal(respContentBytes, &records); err != nil {
		return nil, err
	}

	/*
		for _, result := range records.Result {
			for _, rate := range result.Rates {
				fmt.Println("Rate Value:", rate.Value)
				for _, cnCode := range rate.CnCodes {

					fmt.Println("CnCode: ", cnCode.Code, cnCode.CodeMatcher, criteria.CommodityCodes)
					if slices.Contains(criteria.CommodityCodes, cnCode.Code) {
						fmt.Println("CnCode: ", cnCode.Code, cnCode.CodeMatcher)
					}
				}
			}
		}
	*/

	return &records, nil
}

func (tedb *TEDB) VatSearch(criteria Criteria) ([]TEDBSearchVatResult, error) {

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

	/*
		for _, field := range []string{"CountryCodes", "CommodityCodes", "Categories"} {
			r := reflect.ValueOf(criteria)
			values := reflect.Indirect(r).FieldByName(field).Interface().([]string)
			if err := findDuplicates(values); err != nil {
				return nil, err
			}
		}
	*/

	if err := findDuplicates(criteria.CountryCodes); err != nil {
		return nil, err
	}

	if err := findDuplicates(criteria.CommodityCodes); err != nil {
		return nil, err
	}

	if err := findDuplicates(criteria.Categories); err != nil {
		return nil, err
	}

	records, err := tedb.VatSearchQuery(criteria)
	if err != nil {
		return nil, err
	}

	if records.Errors != nil {
		return nil, errors.New("the search result has an error")
	}

	/*
		vatRate, err := func() (*float64, error) {
			selectedValue, err := func() (*float64, error) {
				if strings.EqualFold(records.Result[0].Type, "STANDARD") {
					return &records.Result[0].Rates[0].Value, nil
				}

				return nil, errors.New("no standard rate value for the input criteria")
			}()
			if err != nil {
				return nil, err
			}

			for _, result := range records.Result {
				for _, rate := range result.Rates {

					if strings.EqualFold(result.IsoCode, "ES") &&
						strings.Contains(rate.Comments, "Canary Islands") {
						fmt.Println("Skip:", rate.Comments)
						continue
					}

					for _, cnCode := range rate.CnCodes {
						for _, cCnCode := range criteria.CommodityCodes {
							if strings.HasPrefix(cCnCode, cnCode.Code) {
								fmt.Println("Rate by Specific CnCode:", rate.Value, cCnCode, cnCode.Code)
								return &rate.Value, nil
							}
						}
					}
				}
			}

			return selectedValue, nil
		}()
		fmt.Println("Vat Rate:", *vatRate)
	*/

	if len(records.Result) == 0 {
		return nil, errors.New("there are no results for the input criteria")
	}

	return records.Result, err
}

func NewVatRetrievalService(debugOption ...bool) TEDB {

	debug := false
	if len(debugOption) == 1 {
		debug = debugOption[0]
	}

	tedb := new(TEDB)

	tedb.Url = "https://ec.europa.eu/taxation_customs/tedb"

	tedb.Timeout = 60
	tedb.Debug = debug

	return *tedb
}
