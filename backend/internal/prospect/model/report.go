package model

type ReportFilter struct {
	DateFrom         string
	DateTo           string
	SalesExecutiveID string
	Territory        string
}

type ReportKPI struct {
	TotalVisits   int `json:"totalVisits"`
	WithinRadius  int `json:"withinRadius"`
	OutsideRadius int `json:"outsideRadius"`
	WonProspects  int `json:"wonProspects"`
}

type ReportTrend struct {
	Label         string `json:"label"`
	WithinRadius  int    `json:"withinRadius"`
	OutsideRadius int    `json:"outsideRadius"`
}

type ReportStage struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type SalesPerformance struct {
	SalesExecutiveID   string  `json:"salesExecutiveId"`
	SalesExecutiveName string  `json:"salesExecutiveName"`
	Territory          string  `json:"territory"`
	Visits             int     `json:"visits"`
	WithinRadius       int     `json:"withinRadius"`
	ProspectsWon       int     `json:"prospectsWon"`
	Conversion         float64 `json:"conversion"`
	Performance        int     `json:"performance"`
}

type Report struct {
	KPI         ReportKPI          `json:"kpi"`
	Trends      []ReportTrend      `json:"trends"`
	Stages      []ReportStage      `json:"stages"`
	Performance []SalesPerformance `json:"performance"`
	Territories []string           `json:"territories"`
}
