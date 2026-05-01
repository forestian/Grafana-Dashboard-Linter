package risk

import "grafana-dashboard-linter/internal/model"

func ValidThreshold(value string) bool {
	switch value {
	case string(model.RiskNone), string(model.RiskLow), string(model.RiskMedium), string(model.RiskHigh):
		return true
	default:
		return false
	}
}

func Rank(value string) int {
	switch value {
	case string(model.RiskHigh):
		return 3
	case string(model.RiskMedium):
		return 2
	case string(model.RiskLow):
		return 1
	default:
		return 0
	}
}

func ThresholdMet(findings []model.Finding, threshold string) bool {
	if threshold == string(model.RiskNone) {
		return false
	}

	minRank := Rank(threshold)
	for _, finding := range findings {
		if Rank(finding.Risk) >= minRank {
			return true
		}
	}
	return false
}
