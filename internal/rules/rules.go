package rules

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"grafana-dashboard-linter/internal/model"
)

type Options struct {
	Strict       bool
	IgnoredRules map[string]bool
}

type lintContext struct {
	file           string
	dashboardTitle string
	ignoredRules   map[string]bool
	findings       []model.Finding
}

func LintDashboard(file string, dashboard map[string]any, opts Options) model.DashboardReport {
	normalizedFile := normalizePath(file)
	title := stringField(dashboard, "title")
	uid := stringField(dashboard, "uid")
	panels := collectPanels(dashboard)

	ctx := lintContext{
		file:           normalizedFile,
		dashboardTitle: title,
		ignoredRules:   opts.IgnoredRules,
	}

	lintDashboardLevel(&ctx, dashboard, panels, opts.Strict)
	lintPanels(&ctx, panels, opts.Strict)
	lintTemplating(&ctx, dashboard)

	return model.DashboardReport{
		File:       normalizedFile,
		Title:      title,
		UID:        uid,
		PanelCount: len(panels),
		Findings:   ctx.findings,
	}
}

func (ctx *lintContext) add(finding model.Finding) {
	if ctx.ignoredRules[finding.RuleID] {
		return
	}
	finding.File = ctx.file
	finding.Dashboard = ctx.dashboardTitle
	ctx.findings = append(ctx.findings, finding)
}

func normalizePath(path string) string {
	if path == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func stringField(object map[string]any, key string) string {
	value, ok := object[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}

func objectField(object map[string]any, key string) map[string]any {
	value, ok := object[key]
	if !ok {
		return nil
	}
	nested, _ := value.(map[string]any)
	return nested
}

func arrayField(object map[string]any, key string) []any {
	value, ok := object[key]
	if !ok {
		return nil
	}
	array, _ := value.([]any)
	return array
}

func boolField(object map[string]any, key string) bool {
	value, ok := object[key]
	if !ok {
		return false
	}
	boolean, _ := value.(bool)
	return boolean
}

func evidenceValue(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	case json.Number:
		return typed.String()
	case fmt.Stringer:
		return typed.String()
	default:
		return fmt.Sprint(typed)
	}
}

func panelIDKey(value any) (string, bool) {
	switch typed := value.(type) {
	case nil:
		return "", false
	case json.Number:
		text := strings.TrimSpace(typed.String())
		return text, text != ""
	case float64:
		if typed == float64(int64(typed)) {
			return strconv.FormatInt(int64(typed), 10), true
		}
		return strconv.FormatFloat(typed, 'f', -1, 64), true
	case string:
		text := strings.TrimSpace(typed)
		return text, text != ""
	default:
		text := strings.TrimSpace(fmt.Sprint(typed))
		return text, text != ""
	}
}

func hasUnit(panel map[string]any) bool {
	fieldConfig := objectField(panel, "fieldConfig")
	if fieldConfig == nil {
		return false
	}
	defaults := objectField(fieldConfig, "defaults")
	if defaults == nil {
		return false
	}
	return stringField(defaults, "unit") != ""
}

func pathJoin(base string, child string) string {
	if base == "" {
		return child
	}
	if strings.HasPrefix(child, "[") {
		return base + child
	}
	return base + "." + child
}
