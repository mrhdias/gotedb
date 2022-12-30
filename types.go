package tedb

type CodeRecord struct {
	ModifiedBy        string      `json:"modifiedBy"`
	ModifiedOn        int64       `json:"modifiedOn"`
	CreatedBy         interface{} `json:"createdBy"`
	CreatedOn         interface{} `json:"createdOn"`
	Code              string      `json:"code"`
	Order             int         `json:"order"`
	StartDate         interface{} `json:"startDate"`
	EndDate           interface{} `json:"endDate"`
	Description       string      `json:"description"`
	ID                int         `json:"id"`
	ParentDescription string      `json:"parentDescription"`
	Iso8601CreatedOn  interface{} `json:"iso8601CreatedOn"`
	Iso8601ModifiedOn string      `json:"iso8601ModifiedOn"`
}

type CnCode struct {
	Key struct {
		ModifiedBy        string      `json:"modifiedBy"`
		ModifiedOn        int64       `json:"modifiedOn"`
		CreatedBy         interface{} `json:"createdBy"`
		CreatedOn         interface{} `json:"createdOn"`
		Code              string      `json:"code"`
		Order             int         `json:"order"`
		StartDate         interface{} `json:"startDate"`
		EndDate           interface{} `json:"endDate"`
		Description       string      `json:"description"`
		ID                int         `json:"id"`
		ParentDescription string      `json:"parentDescription"`
		Iso8601CreatedOn  interface{} `json:"iso8601CreatedOn"`
		Iso8601ModifiedOn string      `json:"iso8601ModifiedOn"`
	} `json:"key"`
	Value interface{} `json:"value"`
}

type TEDBVatSearchResult struct {
	SelectedMemberStates []interface{} `json:"selectedMemberStates"`
	Historized           bool          `json:"historized"`
	MemberState          struct {
		ModifiedBy             string      `json:"modifiedBy"`
		ModifiedOn             int64       `json:"modifiedOn"`
		ID                     int         `json:"id"`
		Name                   string      `json:"name"`
		DefaultCountryCode     string      `json:"defaultCountryCode"`
		AlternativeCountryCode interface{} `json:"alternativeCountryCode"`
		Email                  string      `json:"email"`
		DefaultCurrency        struct {
			ModifiedBy        interface{} `json:"modifiedBy"`
			ModifiedOn        interface{} `json:"modifiedOn"`
			ID                int         `json:"id"`
			IsoCode           string      `json:"isoCode"`
			Description       string      `json:"description"`
			Iso8601ModifiedOn interface{} `json:"iso8601ModifiedOn"`
		} `json:"defaultCurrency"`
		MemberStateLabel  string `json:"memberStateLabel"`
		Iso8601ModifiedOn string `json:"iso8601ModifiedOn"`
	} `json:"memberState"`
	Type string `json:"type"`
	Rate struct {
		Key   string  `json:"key"`
		Value float64 `json:"value"`
	} `json:"rate"`
	CnCodes     []CnCode      `json:"cnCodes"`
	CpaCodes    []interface{} `json:"cpaCodes"`
	Category    string        `json:"category"`
	Comments    string        `json:"comments"`
	SituationOn int64         `json:"situationOn"`
}
