package report

import (
	"fmt"

	"grafana-dashboard-linter/internal/model"
)

func ValidFormat(format string) bool {
	switch format {
	case "text", "json", "markdown":
		return true
	default:
		return false
	}
}

func Render(report model.Report, format string) (string, error) {
	switch format {
	case "text":
		return RenderText(report), nil
	case "json":
		return RenderJSON(report)
	case "markdown":
		return RenderMarkdown(report), nil
	default:
		return "", fmt.Errorf("unsupported report format %q", format)
	}
}
