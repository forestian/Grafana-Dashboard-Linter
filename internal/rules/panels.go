package rules

import (
	"fmt"
	"strings"

	"grafana-dashboard-linter/internal/model"
)

type panelRef struct {
	panel map[string]any
	path  string
	id    any
	title string
	typ   string
}

func collectPanels(dashboard map[string]any) []panelRef {
	return collectPanelsFromArray(arrayField(dashboard, "panels"), "panels")
}

func collectPanelsFromArray(panels []any, basePath string) []panelRef {
	var refs []panelRef
	for i, raw := range panels {
		panel, ok := raw.(map[string]any)
		if !ok {
			continue
		}

		panelPath := fmt.Sprintf("%s[%d]", basePath, i)
		ref := panelRef{
			panel: panel,
			path:  panelPath,
			id:    panel["id"],
			title: stringField(panel, "title"),
			typ:   strings.ToLower(stringField(panel, "type")),
		}
		refs = append(refs, ref)

		nested := collectPanelsFromArray(arrayField(panel, "panels"), pathJoin(panelPath, "panels"))
		refs = append(refs, nested...)
	}
	return refs
}

func lintPanels(ctx *lintContext, panels []panelRef, strict bool) {
	lintDuplicatePanelIDs(ctx, panels)

	for _, panel := range panels {
		if panel.title == "" {
			ctx.add(panelFinding(panel, string(model.RiskMedium), "missing-panel-title", pathJoin(panel.path, "title"), "Panel title is missing", "", "Panels without titles are hard to understand during incidents.", "Add a clear panel title."))
		}

		if panel.typ == "" {
			ctx.add(panelFinding(panel, string(model.RiskLow), "missing-panel-type", pathJoin(panel.path, "type"), "Panel type is missing", "", "Panel type is missing.", "Re-export or fix the panel definition."))
		}

		targets := arrayField(panel.panel, "targets")
		if len(targets) > 0 && !hasDatasource(panel.panel["datasource"]) && !targetsHaveDatasource(targets) {
			ctx.add(panelFinding(panel, string(model.RiskMedium), "missing-datasource", pathJoin(panel.path, "datasource"), "Panel datasource is missing", "", "Panel has no datasource, so queries may fail depending on Grafana defaults.", "Set a datasource or use dashboard datasource variables intentionally."))
		}

		if evidence, ok := hardcodedDatasourceEvidence(panel.panel["datasource"]); ok {
			ctx.add(panelFinding(panel, string(model.RiskLow), "hardcoded-datasource", pathJoin(panel.path, "datasource"), "Datasource is hardcoded", evidence, "Hardcoded datasources can make dashboards less portable across environments.", "Consider using a datasource variable such as ${DS_PROMETHEUS} or a consistent datasource UID."))
		}

		if len(targets) > 12 {
			ctx.add(panelFinding(panel, string(model.RiskMedium), "too-many-targets", pathJoin(panel.path, "targets"), "Panel has too many targets", "target count: "+intToString(len(targets)), "Panels with many queries can be expensive and hard to debug.", "Split the panel or simplify queries."))
		}

		lintTargets(ctx, panel, targets)

		if needsUnit(panel.typ) && !hasUnit(panel.panel) {
			ctx.add(panelFinding(panel, string(model.RiskLow), "missing-unit", pathJoin(panel.path, "fieldConfig.defaults.unit"), "Panel unit is missing", "", "Panels without units are harder to interpret.", "Set a meaningful unit such as percent, bytes, seconds, requests/sec, or ops/sec."))
		}

		if strict && stringField(panel.panel, "description") == "" {
			ctx.add(panelFinding(panel, string(model.RiskLow), "missing-description-strict", pathJoin(panel.path, "description"), "Panel description is missing", "", "Descriptions help users understand dashboard intent and panel meaning.", "Add descriptions for important dashboards and panels."))
		}

		if boolField(panel.panel, "transparent") {
			ctx.add(panelFinding(panel, string(model.RiskLow), "hidden-panel", pathJoin(panel.path, "transparent"), "Panel is transparent", "transparent: true", "Hidden or collapsed panels may be missed during review.", "Verify whether hidden or collapsed panels are intentional."))
		}

		if panel.typ == "row" && boolField(panel.panel, "collapsed") && len(arrayField(panel.panel, "panels")) > 5 {
			ctx.add(panelFinding(panel, string(model.RiskLow), "hidden-panel", pathJoin(panel.path, "collapsed"), "Collapsed row contains many panels", "nested panels: "+intToString(len(arrayField(panel.panel, "panels"))), "Hidden or collapsed panels may be missed during review.", "Verify whether hidden or collapsed panels are intentional."))
		}
	}
}

func lintDuplicatePanelIDs(ctx *lintContext, panels []panelRef) {
	byID := map[string][]panelRef{}
	for _, panel := range panels {
		key, ok := panelIDKey(panel.id)
		if !ok {
			continue
		}
		byID[key] = append(byID[key], panel)
	}

	for _, duplicates := range byID {
		if len(duplicates) < 2 {
			continue
		}
		for _, panel := range duplicates {
			ctx.add(panelFinding(panel, string(model.RiskHigh), "duplicate-panel-id", pathJoin(panel.path, "id"), "Duplicate panel ID", "id: "+evidenceValue(panel.id), "Duplicate panel IDs can cause confusing dashboard behavior and broken panel links.", "Regenerate panel IDs or re-export the dashboard."))
		}
	}
}

func panelFinding(panel panelRef, risk string, ruleID string, path string, title string, evidence string, explanation string, suggestion string) model.Finding {
	return model.Finding{
		PanelID:     panel.id,
		PanelTitle:  panel.title,
		Path:        path,
		Risk:        risk,
		RuleID:      ruleID,
		Title:       title,
		Evidence:    evidence,
		Explanation: explanation,
		Suggestion:  suggestion,
	}
}

func hasDatasource(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(typed) != ""
	case map[string]any:
		return stringField(typed, "uid") != "" || stringField(typed, "name") != "" || stringField(typed, "type") != ""
	default:
		return true
	}
}

func targetsHaveDatasource(targets []any) bool {
	for _, raw := range targets {
		target, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if hasDatasource(target["datasource"]) {
			return true
		}
	}
	return false
}

func hardcodedDatasourceEvidence(value any) (string, bool) {
	switch typed := value.(type) {
	case nil:
		return "", false
	case string:
		text := strings.TrimSpace(typed)
		if text == "" || datasourceLooksVariable(text) {
			return "", false
		}
		return text, true
	case map[string]any:
		for _, key := range []string{"uid", "name"} {
			text := stringField(typed, key)
			if text == "" || datasourceLooksVariable(text) {
				continue
			}
			return key + ": " + text, true
		}
	}
	return "", false
}

func datasourceLooksVariable(value string) bool {
	return strings.Contains(value, "$")
}

func needsUnit(panelType string) bool {
	normalized := strings.ReplaceAll(strings.ToLower(panelType), " ", "")
	switch normalized {
	case "timeseries", "stat", "gauge", "bargauge", "graph":
		return true
	default:
		return false
	}
}

func intToString(value int) string {
	return fmt.Sprintf("%d", value)
}
