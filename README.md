# grafana-dashboard-linter

`gdlint` is a local CLI tool for scanning Grafana dashboard JSON files before they are committed, shared, or imported into Grafana.

It reports common quality, maintainability, and operational issues such as empty panel titles, missing datasources, empty queries, duplicate panel IDs, aggressive refresh intervals, missing units, broad Loki or Prometheus queries, and weak templating variables.

## Why Dashboard Linting Matters

Grafana dashboards often become operational runbooks during incidents. Small dashboard problems can slow responders down, overload data sources, or make dashboards hard to reuse across environments. `gdlint` keeps those issues visible in local development and CI.

## Build And Install

```sh
go build -o gdlint .
```

To install into your Go bin directory:

```sh
go install .
```

## Install from GitHub Releases

Download a prebuilt binary from the GitHub Releases page.

Linux/macOS:

```sh
tar -xzf <archive>.tar.gz
chmod +x gdlint
./gdlint version
```

Windows:

Download the Windows archive, extract it, and run:

```powershell
gdlint.exe version
```

## Commands

```sh
gdlint version
gdlint init --output ./dashboard-lint-demo
gdlint check --file examples/bad-dashboard.json
gdlint check --dir ./dashboards
gdlint check --dir ./dashboards --format markdown --output report.md
gdlint check --dir ./dashboards --fail-on-risk high
gdlint check --dir ./dashboards --strict
gdlint check --dir ./dashboards --ignore-rule missing-tags
```

## Init

`gdlint init` creates a demo project:

```text
dashboard-lint-demo/
  README.md
  dashboards/
    good-dashboard.json
    bad-dashboard.json
```

If the output directory already exists, the command fails unless `--force` is set.

## Supported Inputs

- Direct Grafana dashboard JSON.
- Wrapped Grafana export JSON where the dashboard is under the top-level `dashboard` object.
- Recursive directory scans of `.json` files.

Non-JSON files are ignored during directory scans.

## Output Formats

- `text`: human-readable terminal output.
- `json`: full machine-readable report structure.
- `markdown`: GitHub PR comment friendly report.

Use `--output` to write a report to a new file. Existing output files are not overwritten.

## Fail-On-Risk

`--fail-on-risk` controls CI exit behavior:

- `none`: never fail due to findings.
- `low`: fail on low, medium, or high risk findings.
- `medium`: fail on medium or high risk findings.
- `high`: fail on high risk findings.

The report is still printed or written before the command exits non-zero.

## Ignore Rules

Use `--ignore-rule` one or more times to omit matching findings:

```sh
gdlint check --dir dashboards --ignore-rule missing-tags --ignore-rule hardcoded-datasource
```

## Strict Mode

`--strict` enables optional documentation and maintainability checks. In this MVP, strict mode reports missing dashboard and panel descriptions. Strict mode does not change parsing behavior.

## Limitations

- Local file scanning only.
- No Grafana API integration.
- No Prometheus or Loki live query validation.
- No dashboard upload, import, or authentication.
- No JSON schema validation against specific Grafana versions.
- Heuristic query checks are deterministic and conservative, not a replacement for production query review.

## Roadmap

- GitHub Action packaging.
- GitHub PR comment integration.
- Dashboard quality scoring.
- Dashboard score badges.
- Optional dashboard auto-fix suggestions.
- Version-aware Grafana schema validation.
