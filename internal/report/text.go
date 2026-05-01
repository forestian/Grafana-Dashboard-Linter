package report

import (
	"fmt"
	"strings"

	"grafana-dashboard-linter/internal/model"
)

func RenderText(report model.Report) string {
	var builder strings.Builder

	builder.WriteString("Grafana Dashboard Linter\n\n")
	fmt.Fprintf(&builder, "Scanned files: %d\n", report.Summary.FilesScanned)
	fmt.Fprintf(&builder, "Dashboards: %d\n", report.Summary.Dashboards)
	fmt.Fprintf(&builder, "Panels: %d\n\n", report.Summary.Panels)

	builder.WriteString("Summary:\n")
	fmt.Fprintf(&builder, "- Findings: %d\n", report.Summary.Findings)
	fmt.Fprintf(&builder, "- High risk: %d\n", report.Summary.HighRisk)
	fmt.Fprintf(&builder, "- Medium risk: %d\n", report.Summary.MediumRisk)
	fmt.Fprintf(&builder, "- Low risk: %d\n\n", report.Summary.LowRisk)

	builder.WriteString("Findings:\n\n")
	if len(report.Findings) == 0 {
		builder.WriteString("No findings.\n")
		return builder.String()
	}

	for i, finding := range report.Findings {
		if i > 0 {
			builder.WriteString("\n")
		}
		fmt.Fprintf(&builder, "[%s] %s\n", strings.ToUpper(finding.Risk), finding.RuleID)
		fmt.Fprintf(&builder, "File: %s\n", finding.File)
		if finding.Dashboard != "" {
			fmt.Fprintf(&builder, "Dashboard: %s\n", finding.Dashboard)
		}
		if finding.PanelTitle != "" {
			fmt.Fprintf(&builder, "Panel: %s\n", finding.PanelTitle)
		} else if finding.PanelID != nil {
			fmt.Fprintf(&builder, "Panel ID: %v\n", finding.PanelID)
		}
		fmt.Fprintf(&builder, "Path: %s\n", finding.Path)
		if finding.Evidence != "" {
			fmt.Fprintf(&builder, "Evidence: %s\n", finding.Evidence)
		}
		builder.WriteString("\nExplanation:\n")
		builder.WriteString(finding.Explanation)
		builder.WriteString("\n\nSuggestion:\n")
		builder.WriteString(finding.Suggestion)
		builder.WriteString("\n")
	}

	return builder.String()
}
