package rules

import (
	"fmt"
	"strings"

	"grafana-dashboard-linter/internal/model"
)

func lintTemplating(ctx *lintContext, dashboard map[string]any) {
	templating := objectField(dashboard, "templating")
	if templating == nil {
		return
	}

	for i, raw := range arrayField(templating, "list") {
		variable, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		basePath := fmt.Sprintf("templating.list[%d]", i)

		if strings.EqualFold(stringField(variable, "type"), "query") && variableQuery(variable) == "" {
			ctx.add(model.Finding{
				Path:        pathJoin(basePath, "query"),
				Risk:        string(model.RiskMedium),
				RuleID:      "templating-variable-without-query",
				Title:       "Templating query variable has no query",
				Explanation: "A dashboard variable has no query and may not work.",
				Suggestion:  "Add a valid query or remove the variable.",
			})
		}

		if allValue := stringField(variable, "allValue"); allValue == ".*" {
			ctx.add(model.Finding{
				Path:        pathJoin(basePath, "allValue"),
				Risk:        string(model.RiskLow),
				RuleID:      "risky-regex-all-values",
				Title:       "Variable all value is broad",
				Evidence:    allValue,
				Explanation: "Overly broad variables can create expensive queries.",
				Suggestion:  "Constrain variable values where possible.",
			})
		}

		if regex := stringField(variable, "regex"); regexLooksBroad(regex) {
			ctx.add(model.Finding{
				Path:        pathJoin(basePath, "regex"),
				Risk:        string(model.RiskLow),
				RuleID:      "risky-regex-all-values",
				Title:       "Variable regex is broad",
				Evidence:    regex,
				Explanation: "Overly broad variables can create expensive queries.",
				Suggestion:  "Constrain variable values where possible.",
			})
		}
	}
}

func variableQuery(variable map[string]any) string {
	switch typed := variable["query"].(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]any:
		return stringField(typed, "query")
	default:
		return ""
	}
}

func regexLooksBroad(regex string) bool {
	switch strings.TrimSpace(regex) {
	case ".*", "/.*/", "(.*)", "^.*$":
		return true
	default:
		return false
	}
}
