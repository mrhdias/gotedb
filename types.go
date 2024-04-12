package tedb

type TEDBcnCode struct {
	ID                int    `json:"id"`
	Code              string `json:"code"`
	Description       string `json:"description"`
	ParentDescription string `json:"parentDescription"`
	CodeMatcher       any    `json:"codeMatcher"`
}

type TEDBrate struct {
	Key         string       `json:"key"`
	Value       float64      `json:"value"`
	SituationOn string       `json:"situationOn"`
	CnCodes     []TEDBcnCode `json:"cnCodes"`
	CpaCodes    []any        `json:"cpaCodes"`
	Category    string       `json:"category"`
	Comments    string       `json:"comments"`
}

type TEDBSearchVatResult struct {
	CountryName string     `json:"countryName"`
	IsoCode     string     `json:"isoCode"`
	Historized  string     `json:"historized"`
	Type        string     `json:"type"`
	Rates       []TEDBrate `json:"rates"`
}

type TEDBsearchForm struct {
	SelectedMemberStates []string `json:"selectedMemberStates"`
	DateFrom             any      `json:"dateFrom"`
	DateTo               any      `json:"dateTo"`
	SelectedCategories   []string `json:"selectedCategories"`
	SelectedCnCodes      []string `json:"selectedCnCodes"`
	SelectedCpaCodes     []string `json:"selectedCpaCodes"`
}

type TEDBsearchResult struct {
	Errors        any `json:"errors"`
	InitialSearch struct {
		SearchForm      TEDBsearchForm `json:"searchForm"`
		SelectedFacets  any            `json:"selectedFacets"`
		AvailableFacets []struct {
			ID     string `json:"id"`
			Label  string `json:"label"`
			Facets []struct {
				ID            string `json:"id"`
				Value         string `json:"value"`
				CountElements int    `json:"countElements"`
				Type          string `json:"type"`
				Label         string `json:"label"`
				GroupID       string `json:"groupId"`
				Children      any    `json:"children"`
			} `json:"facets"`
		} `json:"availableFacets"`
	} `json:"initialSearch"`
	Result                []TEDBSearchVatResult `json:"result"`
	CnCodeNonCompliantMs  []any                 `json:"cnCodeNonCompliantMs"`
	CpaCodeNonCompliantMs []any                 `json:"cpaCodeNonCompliantMs"`
	CnCodes               []struct {
		ID                any    `json:"id"`
		Code              string `json:"code"`
		Description       any    `json:"description"`
		ParentDescription string `json:"parentDescription"`
		CodeMatcher       any    `json:"codeMatcher"`
	} `json:"cnCodes"`
	CpaCodes []any `json:"cpaCodes"`
}

type TEDBSearch struct {
	SearchForm      TEDBsearchForm `json:"searchForm"`
	AvailableFacets any            `json:"availableFacets"`
	SelectedFacets  any            `json:"selectedFacets"`
}
