package cmd

import (
	"fmt"
	"os"
	"strings"

	"grafana-dashboard-linter/internal/report"
	"grafana-dashboard-linter/internal/risk"
	"grafana-dashboard-linter/internal/scanner"

	"github.com/spf13/cobra"
)

type checkOptions struct {
	file       string
	dir        string
	format     string
	output     string
	failOnRisk string
	strict     bool
	ignoreRule []string
}

func newCheckCommand() *cobra.Command {
	opts := checkOptions{
		format:     "text",
		failOnRisk: "none",
	}

	cmd := &cobra.Command{
		Use:   "check",
		Short: "Check Grafana dashboard JSON files",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheck(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.file, "file", "", "dashboard JSON file to check")
	cmd.Flags().StringVar(&opts.dir, "dir", "", "directory of dashboard JSON files to check recursively")
	cmd.Flags().StringVar(&opts.format, "format", opts.format, "output format: text, json, or markdown")
	cmd.Flags().StringVar(&opts.output, "output", "", "write report to this path instead of stdout")
	cmd.Flags().StringVar(&opts.failOnRisk, "fail-on-risk", opts.failOnRisk, "exit non-zero for findings at or above this risk: none, low, medium, or high")
	cmd.Flags().BoolVar(&opts.strict, "strict", false, "enable optional documentation and maintainability checks")
	cmd.Flags().StringArrayVar(&opts.ignoreRule, "ignore-rule", nil, "rule ID to ignore; can be provided multiple times")

	return cmd
}

func runCheck(cmd *cobra.Command, opts checkOptions) error {
	opts.format = strings.ToLower(strings.TrimSpace(opts.format))
	opts.failOnRisk = strings.ToLower(strings.TrimSpace(opts.failOnRisk))

	if err := validateCheckOptions(opts); err != nil {
		return err
	}

	scanOpts := scanner.Options{
		Strict:      opts.strict,
		IgnoreRules: opts.ignoreRule,
	}

	var rep scanner.Report
	var err error
	if opts.file != "" {
		rep, err = scanner.ScanFile(opts.file, scanOpts)
	} else {
		rep, err = scanner.ScanDir(opts.dir, scanOpts)
	}
	if err != nil {
		return err
	}

	rendered, err := report.Render(rep, opts.format)
	if err != nil {
		return err
	}

	if opts.output != "" {
		if err := writeReportFile(opts.output, rendered); err != nil {
			return err
		}
	} else {
		fmt.Fprint(cmd.OutOrStdout(), rendered)
	}

	if risk.ThresholdMet(rep.Findings, opts.failOnRisk) {
		return fmt.Errorf("findings meet fail-on-risk threshold %q", opts.failOnRisk)
	}

	return nil
}

func validateCheckOptions(opts checkOptions) error {
	hasFile := strings.TrimSpace(opts.file) != ""
	hasDir := strings.TrimSpace(opts.dir) != ""
	if hasFile == hasDir {
		return fmt.Errorf("exactly one input source is required: --file or --dir")
	}

	if hasFile {
		info, err := os.Stat(opts.file)
		if err != nil {
			return fmt.Errorf("--file %q does not exist: %w", opts.file, err)
		}
		if info.IsDir() {
			return fmt.Errorf("--file %q is a directory", opts.file)
		}
	}

	if hasDir {
		info, err := os.Stat(opts.dir)
		if err != nil {
			return fmt.Errorf("--dir %q does not exist: %w", opts.dir, err)
		}
		if !info.IsDir() {
			return fmt.Errorf("--dir %q is not a directory", opts.dir)
		}
	}

	if !report.ValidFormat(opts.format) {
		return fmt.Errorf("--format must be text, json, or markdown")
	}

	if !risk.ValidThreshold(opts.failOnRisk) {
		return fmt.Errorf("--fail-on-risk must be none, low, medium, or high")
	}

	if opts.output != "" {
		if _, err := os.Stat(opts.output); err == nil {
			return fmt.Errorf("--output %q already exists; refusing to overwrite", opts.output)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check --output path: %w", err)
		}
	}

	return nil
}

func writeReportFile(path string, content string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create report %s: %w", path, err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("write report %s: %w", path, err)
	}
	return nil
}
