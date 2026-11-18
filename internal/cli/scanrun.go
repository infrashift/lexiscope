package cli

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/engine"
	"github.com/infrashift/lexis/internal/output"
	"github.com/infrashift/lexis/internal/parser"
	"github.com/infrashift/lexis/internal/report"
	"github.com/infrashift/lexis/internal/ruleset"
)

// errFindings marks a scan with error findings: exit status 1, and the
// report holds the details.
var errFindings = errors.New("the scan found errors")

func (f *scanFlags) run(cmd *cobra.Command, args []string) error {
	assert.True(cmd != nil, "cobra gives the command")
	assert.True(f != nil, "the flags exist")
	opts, err := f.options(cmd.Flags().Changed("sarif-version"), args)
	if err != nil {
		return err
	}
	if opts.Format == "sarif" {
		return fmt.Errorf("--format sarif: %w: SARIF arrives in Phase 5 of PLAN.md", errNotAvailable)
	}
	rs, err := ruleset.Load(opts.Ruleset)
	if err != nil {
		return err
	}
	e, err := engine.New(rs)
	if err != nil {
		return err
	}
	files, err := engine.Collect(opts.Paths, opts.Excludes)
	if err != nil {
		return err
	}
	findings, err := checkFiles(e, files)
	if err != nil {
		return err
	}
	s := report.Summarize(rs.Name, len(files), findings)
	if err := writeReport(cmd.OutOrStdout(), opts, findings, s); err != nil {
		return err
	}
	if s.Errors > 0 {
		return errFindings
	}
	return nil
}

// checkFiles reads and checks each file.
func checkFiles(e *engine.Engine, files []string) ([]engine.Finding, error) {
	assert.True(e != nil, "the engine exists")
	assert.True(len(files) <= engine.MaxFiles, "the file count is bounded")
	var all []engine.Finding
	for _, p := range files {
		info, err := os.Stat(p)
		if err != nil {
			return nil, err
		}
		if info.Size() > parser.MaxFileBytes {
			return nil, fmt.Errorf("%s: %d bytes, more than %d", p, info.Size(), parser.MaxFileBytes)
		}
		src, err := os.ReadFile(p)
		if err != nil {
			return nil, err
		}
		fs, err := e.Check(p, src)
		if err != nil {
			return nil, err
		}
		all = append(all, fs...)
	}
	return all, nil
}

// writeReport writes the report to stdout, or to the --output file.
func writeReport(stdout io.Writer, opts ScanOptions, findings []engine.Finding, s report.Summary) error {
	assert.True(stdout != nil, "stdout exists")
	assert.True(opts.Output != "", "the output has a default")
	if opts.Output == "-" {
		return render(output.New(stdout), opts.Format, findings, s)
	}
	file, err := os.Create(opts.Output)
	if err != nil {
		return err
	}
	werr := render(output.New(file), opts.Format, findings, s)
	return errors.Join(werr, file.Close())
}

// render writes the report in one format.
func render(out *output.Writer, format string, findings []engine.Finding, s report.Summary) error {
	assert.True(out != nil, "the writer exists")
	assert.True(format == "text" || format == "json", "the options checked the format")
	if format == "json" {
		if err := report.JSON(out, findings, s); err != nil {
			return err
		}
		return out.Err()
	}
	report.Text(out, findings, s)
	return out.Err()
}
