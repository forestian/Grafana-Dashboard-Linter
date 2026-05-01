package scanner

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"grafana-dashboard-linter/internal/model"
)

func TestInvalidJSONDetection(t *testing.T) {
	path := writeTempDashboard(t, `{"title":`)

	report, err := ScanFile(path, Options{})
	if err != nil {
		t.Fatalf("ScanFile returned error: %v", err)
	}

	if !hasRule(report, "invalid-json") {
		t.Fatalf("expected invalid-json finding, got %#v", report.Findings)
	}
	if report.Summary.HighRisk != 1 {
		t.Fatalf("high risk count = %d, want 1", report.Summary.HighRisk)
	}
}

func TestDashboardRules(t *testing.T) {
	tests := []struct {
		name      string
		json      string
		strict    bool
		wantRules []string
	}{
		{
			name: "missing dashboard title",
			json: `{
				"uid": "uid",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": [{"id": 1, "title": "Notes", "type": "text"}]
			}`,
			wantRules: []string{"missing-dashboard-title"},
		},
		{
			name: "missing dashboard uid",
			json: `{
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": [{"id": 1, "title": "Notes", "type": "text"}]
			}`,
			wantRules: []string{"missing-dashboard-uid"},
		},
		{
			name: "no panels",
			json: `{
				"uid": "uid",
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": []
			}`,
			wantRules: []string{"no-panels"},
		},
		{
			name: "duplicate panel ID",
			json: `{
				"uid": "uid",
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": [
					{"id": 1, "title": "One", "type": "text"},
					{"id": 1, "title": "Two", "type": "text"}
				]
			}`,
			wantRules: []string{"duplicate-panel-id"},
		},
		{
			name: "missing panel title",
			json: `{
				"uid": "uid",
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": [{"id": 1, "title": "", "type": "text"}]
			}`,
			wantRules: []string{"missing-panel-title"},
		},
		{
			name: "missing datasource",
			json: `{
				"uid": "uid",
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": [
					{"id": 1, "title": "Query", "type": "logs", "targets": [{"expr": "up"}]}
				]
			}`,
			wantRules: []string{"missing-datasource"},
		},
		{
			name: "empty query",
			json: `{
				"uid": "uid",
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": [
					{
						"id": 1,
						"title": "Query",
						"type": "logs",
						"datasource": {"type": "loki", "uid": "${DS_LOKI}"},
						"targets": [{"expr": ""}]
					}
				]
			}`,
			wantRules: []string{"empty-query"},
		},
		{
			name: "aggressive refresh interval",
			json: `{
				"uid": "uid",
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"refresh": "1s",
				"panels": [{"id": 1, "title": "Notes", "type": "text"}]
			}`,
			wantRules: []string{"excessive-interval"},
		},
		{
			name: "missing unit",
			json: `{
				"uid": "uid",
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": [
					{
						"id": 1,
						"title": "CPU",
						"type": "timeseries",
						"datasource": {"type": "prometheus", "uid": "${DS_PROMETHEUS}"},
						"targets": [{"expr": "up"}]
					}
				]
			}`,
			wantRules: []string{"missing-unit"},
		},
		{
			name:   "strict missing description",
			strict: true,
			json: `{
				"uid": "uid",
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": [{"id": 1, "title": "Notes", "type": "text"}]
			}`,
			wantRules: []string{"missing-description-strict"},
		},
		{
			name: "loki broad query",
			json: `{
				"uid": "uid",
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": [
					{
						"id": 1,
						"title": "Logs",
						"type": "logs",
						"datasource": {"type": "loki", "uid": "${DS_LOKI}"},
						"targets": [{"expr": "{} |= \"error\""}]
					}
				]
			}`,
			wantRules: []string{"loki-broad-query"},
		},
		{
			name: "prometheus heavy query",
			json: `{
				"uid": "uid",
				"title": "Dashboard",
				"tags": ["platform"],
				"schemaVersion": 39,
				"panels": [
					{
						"id": 1,
						"title": "Requests",
						"type": "timeseries",
						"datasource": {"type": "prometheus", "uid": "${DS_PROMETHEUS}"},
						"fieldConfig": {"defaults": {"unit": "reqps"}},
						"targets": [{"expr": "sum(rate(http_requests_total{job=~\".*\"}[10s]))"}]
					}
				]
			}`,
			wantRules: []string{"prometheus-heavy-query"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			path := writeTempDashboard(t, test.json)
			report, err := ScanFile(path, Options{Strict: test.strict})
			if err != nil {
				t.Fatalf("ScanFile returned error: %v", err)
			}

			for _, ruleID := range test.wantRules {
				if !hasRule(report, ruleID) {
					t.Fatalf("expected rule %s, got rules %v", ruleID, ruleIDs(report))
				}
			}
		})
	}
}

func TestIgnoreRuleBehavior(t *testing.T) {
	path := writeTempDashboard(t, `{
		"uid": "uid",
		"title": "Dashboard",
		"tags": [],
		"schemaVersion": 39,
		"panels": [{"id": 1, "title": "Notes", "type": "text"}]
	}`)

	report, err := ScanFile(path, Options{IgnoreRules: []string{"missing-tags"}})
	if err != nil {
		t.Fatalf("ScanFile returned error: %v", err)
	}
	if hasRule(report, "missing-tags") {
		t.Fatalf("ignored missing-tags rule appeared in findings")
	}
}

func TestDirectoryRecursiveScanning(t *testing.T) {
	root := t.TempDir()
	nested := filepath.Join(root, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("MkdirAll: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "one.json"), []byte(validMinimalDashboard("one")), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(nested, "two.json"), []byte(validMinimalDashboard("two")), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "ignored.txt"), []byte("{}"), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	report, err := ScanDir(root, Options{})
	if err != nil {
		t.Fatalf("ScanDir returned error: %v", err)
	}
	if report.Summary.FilesScanned != 2 {
		t.Fatalf("files scanned = %d, want 2", report.Summary.FilesScanned)
	}
	if len(report.Dashboards) != 2 {
		t.Fatalf("dashboards = %d, want 2", len(report.Dashboards))
	}
	if !strings.HasSuffix(report.Dashboards[0].File, "one.json") || !strings.HasSuffix(report.Dashboards[1].File, "two.json") {
		t.Fatalf("dashboards not sorted by file: %#v", report.Dashboards)
	}
}

func writeTempDashboard(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "dashboard.json")
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func validMinimalDashboard(uid string) string {
	return `{
		"uid": "` + uid + `",
		"title": "Dashboard ` + uid + `",
		"tags": ["platform"],
		"schemaVersion": 39,
		"panels": [{"id": 1, "title": "Notes", "type": "text"}]
	}`
}

func hasRule(report model.Report, ruleID string) bool {
	for _, finding := range report.Findings {
		if finding.RuleID == ruleID {
			return true
		}
	}
	return false
}

func ruleIDs(report model.Report) []string {
	rules := make([]string, 0, len(report.Findings))
	for _, finding := range report.Findings {
		rules = append(rules, finding.RuleID)
	}
	return rules
}
