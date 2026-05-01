package model

type RiskLevel string

const (
	RiskNone   RiskLevel = "none"
	RiskLow    RiskLevel = "low"
	RiskMedium RiskLevel = "medium"
	RiskHigh   RiskLevel = "high"
)

type Finding struct {
	File        string `json:"file"`
	Dashboard   string `json:"dashboard"`
	PanelID     any    `json:"panel_id,omitempty"`
	PanelTitle  string `json:"panel_title,omitempty"`
	Path        string `json:"path"`
	Risk        string `json:"risk"`
	RuleID      string `json:"rule_id"`
	Title       string `json:"title"`
	Evidence    string `json:"evidence,omitempty"`
	Explanation string `json:"explanation"`
	Suggestion  string `json:"suggestion"`
}

type DashboardReport struct {
	File       string    `json:"file"`
	Title      string    `json:"title"`
	UID        string    `json:"uid"`
	PanelCount int       `json:"panel_count"`
	Findings   []Finding `json:"findings"`
}

type Summary struct {
	FilesScanned int `json:"files_scanned"`
	Dashboards   int `json:"dashboards"`
	Panels       int `json:"panels"`
	Findings     int `json:"findings"`
	HighRisk     int `json:"high_risk"`
	MediumRisk   int `json:"medium_risk"`
	LowRisk      int `json:"low_risk"`
}

type Report struct {
	Summary    Summary           `json:"summary"`
	Dashboards []DashboardReport `json:"dashboards"`
	Findings   []Finding         `json:"findings"`
}
