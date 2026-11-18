// Package report writes the findings of a scan as text or JSON. Phase 5 of
// PLAN.md adds SARIF.
package report

import (
	"encoding/json"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/engine"
	"github.com/infrashift/lexis/internal/output"
)

// Summary counts the files and the findings of a scan.
type Summary struct {
	Ruleset  string `json:"ruleset"`
	Files    int    `json:"files"`
	Errors   int    `json:"errors"`
	Warnings int    `json:"warnings"`
	Notes    int    `json:"notes"`
}

// Summarize counts the findings by severity.
func Summarize(ruleset string, files int, findings []engine.Finding) Summary {
	assert.True(ruleset != "", "the ruleset has a name")
	assert.True(files >= 0, "a count is not negative")
	s := Summary{Ruleset: ruleset, Files: files}
	for _, f := range findings {
		switch f.Severity {
		case "error":
			s.Errors++
		case "warning":
			s.Warnings++
		default:
			s.Notes++
		}
	}
	return s
}

// Text writes one line for each finding, then the summary:
//
//	path:line:column: severity: RULE-ID: message
func Text(out *output.Writer, findings []engine.Finding, s Summary) {
	assert.True(out != nil, "the writer exists")
	assert.True(s.Errors+s.Warnings+s.Notes == len(findings), "the summary counts each finding")
	for _, f := range findings {
		out.Printf("%s:%d:%d: %s: %s: %s\n", f.Path, f.Line, f.Column, f.Severity, f.RuleID, f.Message)
	}
	out.Printf("lexis: %d error(s), %d warning(s), %d note(s) in %d file(s), ruleset %s\n",
		s.Errors, s.Warnings, s.Notes, s.Files, s.Ruleset)
}

// jsonReport is the document of the JSON format.
type jsonReport struct {
	Summary  Summary          `json:"summary"`
	Findings []engine.Finding `json:"findings"`
}

// JSON writes the findings and the summary as one JSON document.
func JSON(out *output.Writer, findings []engine.Finding, s Summary) error {
	assert.True(out != nil, "the writer exists")
	assert.True(s.Errors+s.Warnings+s.Notes == len(findings), "the summary counts each finding")
	if findings == nil {
		findings = []engine.Finding{}
	}
	data, err := json.MarshalIndent(jsonReport{Summary: s, Findings: findings}, "", "  ")
	if err != nil {
		return err
	}
	out.Printf("%s\n", data)
	return nil
}
