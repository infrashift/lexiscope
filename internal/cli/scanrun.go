package cli

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"

	"github.com/spf13/cobra"

	"github.com/infrashift/lexiscope/internal/assert"
	"github.com/infrashift/lexiscope/internal/engine"
	"github.com/infrashift/lexiscope/internal/output"
	"github.com/infrashift/lexiscope/internal/parser"
	"github.com/infrashift/lexiscope/internal/report"
	"github.com/infrashift/lexiscope/internal/ruleset"
)

// errFindings marks a scan with a finding at the --fail-on severity or
// higher: exit status 1, and the report holds the details.
var errFindings = errors.New("the scan found a finding at the --fail-on severity")

func (f *scanFlags) run(cmd *cobra.Command, args []string) error {
	assert.True(cmd != nil, "cobra gives the command")
	assert.True(f != nil, "the flags exist")
	opts, err := f.options(cmd.Flags().Changed("sarif-version"), args)
	if err != nil {
		return err
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
	if err := writeReport(cmd.OutOrStdout(), opts, rs, findings, s); err != nil {
		return err
	}
	if failed(s, opts.FailOn) {
		return errFindings
	}
	return nil
}

// failed reports whether the summary counts a finding at the failOn
// severity or higher.
func failed(s report.Summary, failOn string) bool {
	assert.True(slices.Contains(severities, failOn), "the options checked --fail-on")
	assert.True(s.Errors >= 0 && s.Warnings >= 0 && s.Notes >= 0, "a count is not negative")
	switch failOn {
	case "error":
		return s.Errors > 0
	case "warning":
		return s.Errors+s.Warnings > 0
	}
	return s.Errors+s.Warnings+s.Notes > 0
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

// writeReport writes the report to stdout, or to the --output file. A
// SARIF report that fails validation writes no file.
func writeReport(stdout io.Writer, opts ScanOptions, rs *ruleset.Ruleset, findings []engine.Finding, s report.Summary) error {
	assert.True(stdout != nil, "stdout exists")
	assert.True(opts.Output != "", "the output has a default")
	var buf bytes.Buffer
	if err := render(output.New(&buf), opts, rs, findings, s); err != nil {
		return err
	}
	if opts.Output == "-" {
		out := output.New(stdout)
		out.Printf("%s", buf.Bytes())
		return out.Err()
	}
	return os.WriteFile(opts.Output, buf.Bytes(), 0o644)
}

// render writes the report in the format of the options.
func render(out *output.Writer, opts ScanOptions, rs *ruleset.Ruleset, findings []engine.Finding, s report.Summary) error {
	assert.True(out != nil, "the writer exists")
	assert.True(slices.Contains(formats, opts.Format), "the options checked the format")
	switch opts.Format {
	case "json":
		if err := report.JSON(out, findings, s); err != nil {
			return err
		}
	case "sarif":
		root, err := os.Getwd()
		if err != nil {
			return err
		}
		if err := report.SARIF(out, opts.SARIFVersion, root, rs, findings); err != nil {
			return err
		}
	default:
		report.Text(out, findings, s)
	}
	return out.Err()
}
