package rules

import (
	"fmt"
	"regexp"
	"strings"

	"grafana-dashboard-linter/internal/model"
)

var (
	prometheusBroadSelectorRE = regexp.MustCompile(`\{\s*(job|namespace)\s*=~\s*"\.\*"\s*\}`)
	prometheusShortRateRE     = regexp.MustCompile(`(?i)\b[[:alpha:]]*rate\s*\([^)]*\[(10s|15s)\]`)
	lokiBroadSelectorRE       = regexp.MustCompile(`\{\s*(job|namespace)\s*=~\s*"\.\*"\s*\}`)
	lokiSelectorRE            = regexp.MustCompile(`\{([^}]*)\}`)
)

var highCardinalityLokiLabels = []string{
	"request_id",
	"trace_id",
	"user_id",
	"session_id",
	"path",
	"url",
	"pod_uid",
	"container_id",
}

func lintTargets(ctx *lintContext, panel panelRef, targets []any) {
	for i, raw := range targets {
		target, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		targetPath := fmt.Sprintf("%s.targets[%d]", panel.path, i)

		if evidence, ok := hardcodedDatasourceEvidence(target["datasource"]); ok {
			ctx.add(panelFinding(panel, string(model.RiskLow), "hardcoded-datasource", pathJoin(targetPath, "datasource"), "Datasource is hardcoded", evidence, "Hardcoded datasources can make dashboards less portable across environments.", "Consider using a datasource variable such as ${DS_PROMETHEUS} or a consistent datasource UID."))
		}

		if interval := stringField(target, "interval"); isAggressiveInterval(interval) {
			ctx.add(panelFinding(panel, string(model.RiskMedium), "excessive-interval", pathJoin(targetPath, "interval"), "Target interval is too aggressive", interval, "Very aggressive refresh or query intervals can overload Grafana and data sources.", "Use a safer refresh interval such as 30s, 1m, or higher unless there is a strong reason."))
		}

		queryFields := []string{"expr", "query", "rawSql", "logql", "promql"}
		foundQueryField := false
		for _, field := range queryFields {
			value, exists := target[field]
			if !exists {
				continue
			}

			foundQueryField = true
			query := queryText(value)
			fieldPath := pathJoin(targetPath, field)
			if strings.TrimSpace(query) == "" {
				ctx.add(panelFinding(panel, string(model.RiskHigh), "empty-query", fieldPath, "Target query is empty", "", "A panel target has an empty query.", "Remove the empty target or add a valid query."))
				continue
			}

			lintQueryHeuristics(ctx, panel, target, field, query, fieldPath)
		}

		if !foundQueryField {
			ctx.add(panelFinding(panel, string(model.RiskHigh), "empty-query", targetPath, "Target query is missing", "no expr, query, rawSql, logql, or promql field", "A panel target has an empty query.", "Remove the empty target or add a valid query."))
		}
	}
}

func lintQueryHeuristics(ctx *lintContext, panel panelRef, target map[string]any, field string, query string, path string) {
	sourceType := datasourceType(panel.panel["datasource"], target["datasource"])

	if looksLikePrometheusQuery(sourceType, field) {
		if evidence, ok := prometheusHeavyEvidence(query); ok {
			ctx.add(panelFinding(panel, string(model.RiskMedium), "prometheus-heavy-query", path, "Prometheus query may be expensive", evidence, "Broad or overly frequent Prometheus queries may be expensive.", "Add more selective labels and review query range windows."))
		}
	}

	if looksLikeLokiQuery(sourceType, field) {
		if evidence, ok := lokiBroadEvidence(query); ok {
			ctx.add(panelFinding(panel, string(model.RiskMedium), "loki-broad-query", path, "Loki query is broad", evidence, "Broad Loki queries can be slow and expensive.", "Use low-cardinality labels such as cluster, namespace, app, service, job, and severity to narrow the query."))
		}
		if evidence, ok := lokiHighCardinalityEvidence(query); ok {
			ctx.add(panelFinding(panel, string(model.RiskMedium), "loki-high-cardinality-label-filter", path, "Loki selector uses high-cardinality labels", evidence, "High-cardinality labels in Loki selectors may indicate risky label design.", "Keep high-cardinality values in log body or structured metadata instead of labels."))
		}
	}
}

func queryText(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.TrimSpace(typed)
	case map[string]any:
		return stringField(typed, "query")
	default:
		return ""
	}
}

func isAggressiveInterval(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1s", "2s", "5s", "1000ms", "2000ms", "5000ms":
		return true
	default:
		return false
	}
}

func datasourceType(panelDatasource any, targetDatasource any) string {
	if target := datasourceTypeFromValue(targetDatasource); target != "" {
		return target
	}
	return datasourceTypeFromValue(panelDatasource)
}

func datasourceTypeFromValue(value any) string {
	switch typed := value.(type) {
	case string:
		return strings.ToLower(strings.TrimSpace(typed))
	case map[string]any:
		if text := stringField(typed, "type"); text != "" {
			return strings.ToLower(text)
		}
		if text := stringField(typed, "uid"); text != "" {
			return strings.ToLower(text)
		}
		if text := stringField(typed, "name"); text != "" {
			return strings.ToLower(text)
		}
	}
	return ""
}

func looksLikePrometheusQuery(sourceType string, field string) bool {
	if strings.Contains(sourceType, "prometheus") || strings.Contains(sourceType, "mimir") || strings.Contains(sourceType, "prom") {
		return true
	}
	return field == "promql"
}

func looksLikeLokiQuery(sourceType string, field string) bool {
	if strings.Contains(sourceType, "loki") {
		return true
	}
	return field == "logql"
}

func prometheusHeavyEvidence(query string) (string, bool) {
	if match := prometheusBroadSelectorRE.FindString(query); match != "" {
		return match, true
	}
	if match := prometheusShortRateRE.FindString(query); match != "" {
		return match, true
	}
	return "", false
}

func lokiBroadEvidence(query string) (string, bool) {
	if strings.Contains(query, "{}") {
		return "{}", true
	}
	if match := lokiBroadSelectorRE.FindString(query); match != "" {
		return match, true
	}
	return "", false
}

func lokiHighCardinalityEvidence(query string) (string, bool) {
	selectors := lokiSelectorRE.FindAllStringSubmatch(query, -1)
	for _, selector := range selectors {
		body := selector[1]
		for _, label := range highCardinalityLokiLabels {
			labelRE := regexp.MustCompile(`\b` + regexp.QuoteMeta(label) + `\s*(!=|!~|=~|=)`)
			if labelRE.MatchString(body) {
				return label, true
			}
		}
	}
	return "", false
}
