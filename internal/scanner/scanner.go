package scanner

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"grafana-dashboard-linter/internal/model"
	"grafana-dashboard-linter/internal/parser"
	"grafana-dashboard-linter/internal/risk"
	"grafana-dashboard-linter/internal/rules"
)

type Options struct {
	Strict      bool
	IgnoreRules []string
}

type Report = model.Report

type scanResult struct {
	dashboardReport model.DashboardReport
	dashboardCount  int
	fileCount       int
}

func ScanFile(path string, opts Options) (model.Report, error) {
	result, err := scanFile(path, opts)
	if err != nil {
		return model.Report{}, err
	}
	return buildReport([]scanResult{result}), nil
}

func ScanDir(root string, opts Options) (model.Report, error) {
	files, err := FindJSONFiles(root)
	if err != nil {
		return model.Report{}, err
	}

	results := make([]scanResult, 0, len(files))
	for _, file := range files {
		result, err := scanFile(file, opts)
		if err != nil {
			return model.Report{}, err
		}
		results = append(results, result)
	}

	return buildReport(results), nil
}

func scanFile(path string, opts Options) (scanResult, error) {
	ignored := ignoredRuleSet(opts.IgnoreRules)
	normalizedFile := normalizePath(path)

	data, err := os.ReadFile(path)
	if err != nil {
		return scanResult{}, fmt.Errorf("read %s: %w", path, err)
	}

	dashboard, err := parser.ParseDashboardBytes(data)
	if err != nil {
		finding := model.Finding{
			File:        normalizedFile,
			Path:        "$",
			Risk:        string(model.RiskHigh),
			RuleID:      "invalid-json",
			Title:       "Dashboard file is not valid JSON",
			Evidence:    err.Error(),
			Explanation: "Dashboard file is not valid JSON.",
			Suggestion:  "Fix JSON syntax and export the dashboard again if necessary.",
		}
		findings := []model.Finding{}
		if !ignored["invalid-json"] {
			findings = append(findings, finding)
		}
		return scanResult{
			dashboardReport: model.DashboardReport{
				File:     normalizedFile,
				Findings: findings,
			},
			fileCount: 1,
		}, nil
	}

	dashboardReport := rules.LintDashboard(path, dashboard, rules.Options{
		Strict:       opts.Strict,
		IgnoredRules: ignored,
	})

	return scanResult{
		dashboardReport: dashboardReport,
		dashboardCount:  1,
		fileCount:       1,
	}, nil
}

func buildReport(results []scanResult) model.Report {
	report := model.Report{}

	for _, result := range results {
		report.Summary.FilesScanned += result.fileCount
		report.Summary.Dashboards += result.dashboardCount
		report.Summary.Panels += result.dashboardReport.PanelCount
		report.Dashboards = append(report.Dashboards, result.dashboardReport)
		report.Findings = append(report.Findings, result.dashboardReport.Findings...)
	}

	sortDashboardReports(report.Dashboards)
	sortFindings(report.Findings)
	for i := range report.Dashboards {
		sortFindings(report.Dashboards[i].Findings)
	}

	report.Summary.Findings = len(report.Findings)
	for _, finding := range report.Findings {
		switch finding.Risk {
		case string(model.RiskHigh):
			report.Summary.HighRisk++
		case string(model.RiskMedium):
			report.Summary.MediumRisk++
		case string(model.RiskLow):
			report.Summary.LowRisk++
		}
	}

	return report
}

func ignoredRuleSet(values []string) map[string]bool {
	ignored := map[string]bool{}
	for _, value := range values {
		if value != "" {
			ignored[value] = true
		}
	}
	return ignored
}

func sortDashboardReports(reports []model.DashboardReport) {
	sort.SliceStable(reports, func(i, j int) bool {
		return fileLess(reports[i].File, reports[j].File)
	})
}

func sortFindings(findings []model.Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		left := findings[i]
		right := findings[j]
		if left.File != right.File {
			return fileLess(left.File, right.File)
		}
		if risk.Rank(left.Risk) != risk.Rank(right.Risk) {
			return risk.Rank(left.Risk) > risk.Rank(right.Risk)
		}
		if left.Path != right.Path {
			return left.Path < right.Path
		}
		return left.RuleID < right.RuleID
	})
}

func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func fileLess(left string, right string) bool {
	leftBase := filepath.Base(left)
	rightBase := filepath.Base(right)
	if leftBase != rightBase {
		return leftBase < rightBase
	}
	return filepath.ToSlash(left) < filepath.ToSlash(right)
}
