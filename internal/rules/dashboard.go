package rules

import "grafana-dashboard-linter/internal/model"

func lintDashboardLevel(ctx *lintContext, dashboard map[string]any, panels []panelRef, strict bool) {
	if stringField(dashboard, "title") == "" {
		ctx.add(model.Finding{
			Path:        "title",
			Risk:        string(model.RiskHigh),
			RuleID:      "missing-dashboard-title",
			Title:       "Dashboard title is missing",
			Explanation: "Dashboard title is required for humans to identify the dashboard.",
			Suggestion:  "Set a clear dashboard title.",
		})
	}

	if stringField(dashboard, "uid") == "" {
		ctx.add(model.Finding{
			Path:        "uid",
			Risk:        string(model.RiskMedium),
			RuleID:      "missing-dashboard-uid",
			Title:       "Dashboard UID is missing",
			Explanation: "A stable dashboard UID helps provisioning, linking, and GitOps workflows.",
			Suggestion:  "Set a stable uid for the dashboard.",
		})
	}

	if _, ok := dashboard["schemaVersion"]; !ok {
		ctx.add(model.Finding{
			Path:        "schemaVersion",
			Risk:        string(model.RiskLow),
			RuleID:      "missing-schema-version",
			Title:       "Schema version is missing",
			Explanation: "Missing schemaVersion may make dashboard compatibility harder to understand.",
			Suggestion:  "Export the dashboard from Grafana or add schemaVersion.",
		})
	}

	if tags := arrayField(dashboard, "tags"); len(tags) == 0 {
		ctx.add(model.Finding{
			Path:        "tags",
			Risk:        string(model.RiskLow),
			RuleID:      "missing-tags",
			Title:       "Dashboard tags are missing",
			Explanation: "Tags help users find dashboards.",
			Suggestion:  "Add useful tags such as service, platform, kubernetes, loki, mimir, tempo, or team name.",
		})
	}

	if topLevelPanels := arrayField(dashboard, "panels"); len(topLevelPanels) == 0 {
		ctx.add(model.Finding{
			Path:        "panels",
			Risk:        string(model.RiskMedium),
			RuleID:      "no-panels",
			Title:       "Dashboard has no panels",
			Explanation: "Dashboard has no panels.",
			Suggestion:  "Add panels or remove unused dashboard files.",
		})
	}

	if refresh := stringField(dashboard, "refresh"); isAggressiveInterval(refresh) {
		ctx.add(model.Finding{
			Path:        "refresh",
			Risk:        string(model.RiskMedium),
			RuleID:      "excessive-interval",
			Title:       "Dashboard refresh interval is too aggressive",
			Evidence:    refresh,
			Explanation: "Very aggressive refresh or query intervals can overload Grafana and data sources.",
			Suggestion:  "Use a safer refresh interval such as 30s, 1m, or higher unless there is a strong reason.",
		})
	}

	if strict && stringField(dashboard, "description") == "" {
		ctx.add(model.Finding{
			Path:        "description",
			Risk:        string(model.RiskLow),
			RuleID:      "missing-description-strict",
			Title:       "Dashboard description is missing",
			Explanation: "Descriptions help users understand dashboard intent and panel meaning.",
			Suggestion:  "Add descriptions for important dashboards and panels.",
		})
	}

	if len(panels) > 40 {
		ctx.add(model.Finding{
			Path:        "panels",
			Risk:        string(model.RiskMedium),
			RuleID:      "too-many-panels",
			Title:       "Dashboard has too many panels",
			Evidence:    panelCountEvidence(len(panels)),
			Explanation: "Very large dashboards can be hard to navigate and expensive to load.",
			Suggestion:  "Split the dashboard by service, signal, or audience.",
		})
	}
}

func panelCountEvidence(count int) string {
	return "panel count: " + intToString(count)
}
