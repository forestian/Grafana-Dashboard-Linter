package risk

import (
	"testing"

	"grafana-dashboard-linter/internal/model"
)

func TestThresholdMet(t *testing.T) {
	findings := []model.Finding{
		{Risk: string(model.RiskLow)},
		{Risk: string(model.RiskMedium)},
	}

	if ThresholdMet(findings, "none") {
		t.Fatalf("none threshold should never fail")
	}
	if !ThresholdMet(findings, "low") {
		t.Fatalf("low threshold should fail on low findings")
	}
	if !ThresholdMet(findings, "medium") {
		t.Fatalf("medium threshold should fail on medium findings")
	}
	if ThresholdMet(findings, "high") {
		t.Fatalf("high threshold should not fail without high findings")
	}
}
