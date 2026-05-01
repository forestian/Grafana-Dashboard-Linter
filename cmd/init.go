package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"grafana-dashboard-linter/internal/demo"

	"github.com/spf13/cobra"
)

type initOptions struct {
	output string
	force  bool
}

func newInitCommand() *cobra.Command {
	opts := initOptions{
		output: "./dashboard-lint-demo",
	}

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Create an example dashboard linting project",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.output, "output", opts.output, "output directory")
	cmd.Flags().BoolVar(&opts.force, "force", false, "replace the output directory if it already exists")

	return cmd
}

func runInit(cmd *cobra.Command, opts initOptions) error {
	if opts.output == "" {
		return fmt.Errorf("--output cannot be empty")
	}

	output := filepath.Clean(opts.output)
	if _, err := os.Stat(output); err == nil {
		if !opts.force {
			return fmt.Errorf("output directory %q already exists; use --force to replace it", output)
		}
		if err := validateReplaceTarget(output); err != nil {
			return err
		}
		if err := os.RemoveAll(output); err != nil {
			return fmt.Errorf("replace output directory: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("check output directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Join(output, "dashboards"), 0o755); err != nil {
		return fmt.Errorf("create demo directories: %w", err)
	}

	files := map[string]string{
		filepath.Join(output, "README.md"):                         demo.ProjectREADME,
		filepath.Join(output, "dashboards", "good-dashboard.json"): demo.GoodDashboard,
		filepath.Join(output, "dashboards", "bad-dashboard.json"):  demo.BadDashboard,
	}

	for path, content := range files {
		if err := writeNewFile(path, content); err != nil {
			return err
		}
	}

	fmt.Fprintf(cmd.OutOrStdout(), "Created demo project at %s\n", filepath.ToSlash(output))
	return nil
}

func validateReplaceTarget(path string) error {
	absolute, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("resolve output directory: %w", err)
	}

	volume := filepath.VolumeName(absolute)
	root := volume + string(os.PathSeparator)
	if absolute == root {
		return fmt.Errorf("refusing to replace filesystem root %q", absolute)
	}

	workingDirectory, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}
	workingDirectory, err = filepath.Abs(workingDirectory)
	if err != nil {
		return fmt.Errorf("resolve working directory: %w", err)
	}
	if absolute == workingDirectory {
		return fmt.Errorf("refusing to replace current working directory %q", absolute)
	}

	return nil
}

func writeNewFile(path string, content string) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("create %s: %w", path, err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
