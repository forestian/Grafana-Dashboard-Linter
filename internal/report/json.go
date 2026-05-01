package report

import (
	"encoding/json"

	"grafana-dashboard-linter/internal/model"
)

func RenderJSON(report model.Report) (string, error) {
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}
