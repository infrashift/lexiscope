package report

import (
	"encoding/json"
	"fmt"
	"net/url"
	"path/filepath"
	"slices"
	"strings"
	"unicode"

	"github.com/owenrumney/go-sarif/v3/pkg/report/convert"
	v210 "github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"

	"github.com/infrashift/lexis/internal/assert"
	"github.com/infrashift/lexis/internal/engine"
	"github.com/infrashift/lexis/internal/output"
	"github.com/infrashift/lexis/internal/ruleset"
	"github.com/infrashift/lexis/internal/version"
)

const (
	// InformationURI is tool.driver.informationUri.
	InformationURI = "https://github.com/infrashift/lexis"
	// RulesHelpURI is the helpUri of each rule: the rules reference.
	RulesHelpURI = "https://infrashift.github.io/lexis/reference/rules/"
)

// SARIF writes the findings as one SARIF run of the version 2.1.0 or 2.2.
// root is the directory of the relative artifact URIs. lexis builds a 2.1.0
// report, and go-sarif converts it to 2.2 without loss. Each report must
// pass `Validate` before lexis writes it (ADR-0006).
func SARIF(out *output.Writer, sarifVersion, root string, rs *ruleset.Ruleset, findings []engine.Finding) error {
	assert.True(out != nil, "the writer exists")
	assert.True(slices.Contains(version.SARIFVersions(), sarifVersion), "the options checked the SARIF version")
	r, err := build(root, rs, findings)
	if err != nil {
		return err
	}
	var doc any = r
	if sarifVersion == "2.2" {
		r22, err := convert.ToV22(r, convert.WithStrictConversion())
		if err != nil {
			return fmt.Errorf("SARIF 2.2: %w", err)
		}
		if err := r22.Validate(); err != nil {
			return fmt.Errorf("SARIF 2.2: %w", err)
		}
		doc = r22
	} else if err := r.Validate(); err != nil {
		return fmt.Errorf("SARIF 2.1.0: %w", err)
	}
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	out.Printf("%s\n", data)
	return nil
}

// build makes the 2.1.0 report: the active rules of rs, and one result for
// each finding.
func build(root string, rs *ruleset.Ruleset, findings []engine.Finding) (*v210.Report, error) {
	assert.True(rs != nil, "the ruleset is resolved")
	assert.True(root != "", "the root is set")
	run := v210.NewRunWithInformationURI("lexis", InformationURI)
	run.Tool.Driver.WithVersion(version.Get().Version)
	// lexis reads lines as Python's universal newlines do (internal/parser).
	run.WithColumnKind("utf16CodeUnits").NewlineSequences = []string{"\r\n", "\n", "\r"}
	for _, rule := range rs.Active() {
		run.AddRule(rule.ID).
			WithShortDescription(v210.NewMultiformatMessageString().WithText(rule.Title)).
			WithHelpURI(RulesHelpURI).
			WithDefaultConfiguration(v210.NewReportingConfiguration().WithLevel(rule.Severity))
	}
	for _, f := range findings {
		uri, err := artifactURI(root, f.Path)
		if err != nil {
			return nil, err
		}
		rule, ok := rs.Rules[f.RuleID]
		assert.True(ok && rule.Enabled, "each finding has an active rule")
		res := run.CreateResultForRule(f.RuleID).
			WithLevel(f.Severity).
			WithMessage(v210.NewTextMessage(f.Message)).
			AddLocation(v210.NewLocationWithPhysicalLocation(v210.NewPhysicalLocation().
				WithArtifactLocation(v210.NewSimpleArtifactLocation(uri)).
				WithRegion(region(f))))
		if rule.Check == "term_substitution" && f.Replacement != "" {
			res.AddFixe(fix(uri, f))
		}
	}
	return v210.NewReport().AddRun(run), nil
}

// region is the span of a finding. The end column is the column after the
// last character, as in SARIF.
func region(f engine.Finding) *v210.Region {
	assert.True(f.Line >= 1 && f.Column >= 1, "a finding has a 1-based position")
	assert.True(f.EndLine > f.Line || (f.EndLine == f.Line && f.EndColumn >= f.Column), "the span is in order")
	return v210.NewRegion().
		WithStartLine(f.Line).WithStartColumn(f.Column).
		WithEndLine(f.EndLine).WithEndColumn(f.EndColumn)
}

// fix replaces the term of a STE-100-05 finding with the approved term.
func fix(uri string, f engine.Finding) *v210.Fix {
	assert.True(uri != "", "the artifact has a URI")
	assert.True(f.Replacement != "" && f.Term != "", "the finding has a term and a replacement")
	text := matchCase(f.Term, f.Replacement)
	change := v210.NewArtifactChange().
		WithArtifactLocation(v210.NewSimpleArtifactLocation(uri)).
		WithReplacements([]*v210.Replacement{v210.NewReplacement().
			WithDeletedRegion(region(f)).
			WithInsertedContent(v210.NewArtifactContent().WithText(text))})
	return v210.NewFix().
		WithDescription(v210.NewTextMessage(fmt.Sprintf("Replace %q with %q.", f.Term, text))).
		WithArtifactChanges([]*v210.ArtifactChange{change})
}

// matchCase gives the replacement a capital first letter when the term has
// one, so `Utilize` becomes `Use`.
func matchCase(term, repl string) string {
	assert.True(term != "", "the term is set")
	assert.True(repl != "", "the replacement is set")
	if !unicode.IsUpper([]rune(term)[0]) {
		return repl
	}
	r := []rune(repl)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// artifactURI returns the path relative to root as a relative URI
// reference with slashes, for example "docs/a%20b.md".
func artifactURI(root, path string) (string, error) {
	assert.True(root != "", "the root is set")
	assert.True(path != "", "a finding has a path")
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return "", fmt.Errorf("%s: no path relative to %s: %w", path, root, err)
	}
	// url.URL adds "./" when the first segment has a colon, so a reader
	// does not take it as a scheme.
	uri := (&url.URL{Path: filepath.ToSlash(rel)}).String()
	assert.True(!strings.HasPrefix(uri, "/"), "the URI is relative")
	return uri, nil
}
