package report

import (
	"encoding/json"
	"strings"
	"testing"

	"grafana-dashboard-linter/internal/model"
)

func TestRenderJSON(t *testing.T) {
	input := sampleReport()

	rendered, err := RenderJSON(input)
	if err != nil {
		t.Fatalf("RenderJSON returned error: %v", err)
	}

	var decoded model.Report
	if err := json.Unmarshal([]byte(rendered), &decoded); err != nil {
		t.Fatalf("rendered JSON did not unmarshal: %v", err)
	}
	if decoded.Summary.Findings != 1 {
		t.Fatalf("findings = %d, want 1", decoded.Summary.Findings)
	}
}

func TestRenderMarkdown(t *testing.T) {
	rendered := RenderMarkdown(sampleReport())

	for _, want := range []string{
		"# Grafana Dashboard Linter",
		"| Scanned files | 1 |",
		"### HIGH / empty-query",
		"**Suggestion:** Add a query.",
	} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("markdown missing %q:\n%s", want, rendered)
		}
	}
}

func sampleReport() model.Report {
	finding := model.Finding{
		File:        "dashboards/bad.json",
		Dashboard:   "Bad",
		PanelTitle:  "CPU",
		Path:        "panels[0].targets[0].expr",
		Risk:        string(model.RiskHigh),
		RuleID:      "empty-query",
		Title:       "Empty query",
		Explanation: "A panel target has an empty query.",
		Suggestion:  "Add a query.",
	}
	return model.Report{
		Summary: model.Summary{
			FilesScanned: 1,
			Dashboards:   1,
			Panels:       1,
			Findings:     1,
			HighRisk:     1,
		},
		Dashboards: []model.DashboardReport{
			{
				File:       "dashboards/bad.json",
				Title:      "Bad",
				UID:        "bad",
				PanelCount: 1,
				Findings:   []model.Finding{finding},
			},
		},
		Findings: []model.Finding{finding},
	}
}
