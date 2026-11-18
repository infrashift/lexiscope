package report

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/owenrumney/go-sarif/v3/pkg/report/convert"
	v210 "github.com/owenrumney/go-sarif/v3/pkg/report/v210/sarif"
	v22 "github.com/owenrumney/go-sarif/v3/pkg/report/v22/sarif"

	"github.com/infrashift/lexis/internal/engine"
	"github.com/infrashift/lexis/internal/output"
	"github.com/infrashift/lexis/internal/ruleset"
)

// scan checks one Markdown text with ste-100 minus STE-100-11, and returns
// the ruleset, the findings and the root.
func scan(t *testing.T, text string) (*ruleset.Ruleset, []engine.Finding, string) {
	t.Helper()
	root := t.TempDir()
	p := filepath.Join(root, "docs", "a b.md")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	rs, err := ruleset.Load(ruleset.Options{Rulesets: []string{"ste-100"}, Omit: []string{"STE-100-11"}})
	if err != nil {
		t.Fatal(err)
	}
	e, err := engine.New(rs)
	if err != nil {
		t.Fatal(err)
	}
	fs, err := e.Check(p, []byte(text))
	if err != nil {
		t.Fatal(err)
	}
	return rs, fs, root
}

// sarifDoc is the part of a SARIF log that the tests read.
type sarifDoc struct {
	Version string `json:"version"`
	GUID    string `json:"guid"`
	Runs    []struct {
		ColumnKind string `json:"columnKind"`
		Tool       struct {
			Driver struct {
				Name, Version, InformationURI string
				Rules                         []struct {
					ID                   string
					HelpURI              string
					DefaultConfiguration struct{ Level string }
				}
			}
		}
		Results []struct {
			RuleID    string
			RuleIndex int
			Level     string
			Message   struct{ Text string }
			Locations []struct {
				PhysicalLocation struct {
					ArtifactLocation struct{ URI string }
					Region           struct{ StartLine, StartColumn, EndLine, EndColumn int }
				}
			}
			Fixes []struct {
				ArtifactChanges []struct {
					Replacements []struct {
						InsertedContent struct{ Text string }
					}
				}
			}
		}
	}
}

func TestSARIF(t *testing.T) {
	rs, fs, root := scan(t, "# Title\n\nIt will work. Utilize the tool.\n")
	if len(fs) != 2 {
		t.Fatalf("findings = %+v, want STE-100-04 and STE-100-05", fs)
	}
	for _, version := range []string{"2.1.0", "2.2"} {
		var buf bytes.Buffer
		out := output.New(&buf)
		if err := SARIF(out, version, root, rs, fs); err != nil || out.Err() != nil {
			t.Fatalf("SARIF %s: %v, %v", version, err, out.Err())
		}
		var doc sarifDoc
		if err := json.Unmarshal(buf.Bytes(), &doc); err != nil {
			t.Fatal(err)
		}
		if doc.Version != version || (version == "2.2") != (doc.GUID != "") || len(doc.Runs) != 1 {
			t.Fatalf("%s: version %q, guid %q, %d runs", version, doc.Version, doc.GUID, len(doc.Runs))
		}
		run := doc.Runs[0]
		d := run.Tool.Driver
		if d.Name != "lexis" || d.Version == "" || d.InformationURI != InformationURI || run.ColumnKind != "utf16CodeUnits" {
			t.Fatalf("%s: driver %+v, column kind %q", version, d, run.ColumnKind)
		}
		// The active rules only: the scan omits STE-100-11.
		if len(d.Rules) != 10 || d.Rules[0].ID != "STE-100-01" || d.Rules[9].ID != "STE-100-10" ||
			d.Rules[0].HelpURI != RulesHelpURI || d.Rules[0].DefaultConfiguration.Level != "error" {
			t.Fatalf("%s: rules %+v", version, d.Rules)
		}
		if len(run.Results) != 2 {
			t.Fatalf("%s: results %+v", version, run.Results)
		}
		for i, r := range run.Results {
			loc := r.Locations[0].PhysicalLocation
			if d.Rules[r.RuleIndex].ID != r.RuleID || r.Level != fs[i].Severity || r.Message.Text != fs[i].Message ||
				loc.ArtifactLocation.URI != "docs/a%20b.md" || loc.Region.StartLine != 3 {
				t.Fatalf("%s: result %d = %+v", version, i, r)
			}
		}
		r := run.Results[1]
		reg := r.Locations[0].PhysicalLocation.Region
		if r.RuleID != "STE-100-05" || reg.StartColumn != 15 || reg.EndLine != 3 || reg.EndColumn != 22 ||
			len(r.Fixes) != 1 || r.Fixes[0].ArtifactChanges[0].Replacements[0].InsertedContent.Text != "Use" {
			t.Fatalf("%s: STE-100-05 result = %+v", version, r)
		}
		if len(run.Results[0].Fixes) != 0 {
			t.Fatalf("%s: a STE-100-04 result has a fix", version)
		}
	}
}

func TestSARIFNoFindings(t *testing.T) {
	rs, fs, root := scan(t, "The tool reads the file.\n")
	for _, version := range []string{"2.1.0", "2.2"} {
		var buf bytes.Buffer
		if err := SARIF(output.New(&buf), version, root, rs, fs); err != nil {
			t.Fatal(err)
		}
		// SARIF: an empty array means "no results"; null means "not run".
		if !strings.Contains(buf.String(), `"results": []`) {
			t.Fatalf("%s: no empty results array:\n%s", version, buf.String())
		}
	}
}

// valid22 returns a 2.2 report of two findings that passes `Validate`.
func valid22(t *testing.T) *v22.Report {
	t.Helper()
	rs, fs, root := scan(t, "It will work. Utilize the tool.\n")
	r, err := build(root, rs, fs)
	if err != nil {
		t.Fatal(err)
	}
	r22, err := convert.ToV22(r, convert.WithStrictConversion())
	if err != nil {
		t.Fatal(err)
	}
	if err := r22.Validate(); err != nil {
		t.Fatalf("the valid report fails: %v", err)
	}
	return r22
}

// TestValidateRejectsBroken22 is the proof of ADR-0006: go-sarif validates a
// SARIF 2.2 report against the 2.2 schema. Each broken report breaks one
// constraint, at the top level and deep in the $defs of the schema.
func TestValidateRejectsBroken22(t *testing.T) {
	cases := []struct {
		name   string
		path   string // the start of the path in the error
		mutate func(r *v22.Report)
	}{
		{"version", "version", func(r *v22.Report) { r.Version = "2.1.0" }},
		{"guid", "guid", func(r *v22.Report) { r.Guid = "not-a-guid" }},
		{"level", "runs.0.results.0.level", func(r *v22.Report) { r.Runs[0].Results[0].Level = "fatal" }},
		{"kind", "runs.0.results.0.kind", func(r *v22.Report) { r.Runs[0].Results[0].Kind = "bad" }},
		{"start line", "runs.0.results.1.locations.0.physicalLocation.region.startLine", func(r *v22.Report) {
			r.Runs[0].Results[1].Locations[0].PhysicalLocation.Region.StartLine = new(0)
		}},
		{"rule default level", "runs.0.tool.driver.rules.2.defaultConfiguration.level", func(r *v22.Report) {
			r.Runs[0].Tool.Driver.Rules[2].DefaultConfiguration.Level = "high"
		}},
		{"driver name", "runs.0.tool.driver", func(r *v22.Report) { r.Runs[0].Tool.Driver.Name = nil }},
		{"column kind", "runs.0.columnKind", func(r *v22.Report) { r.Runs[0].ColumnKind = new("bytes") }},
	}
	for _, c := range cases {
		r := valid22(t)
		c.mutate(r)
		err := r.Validate()
		if err == nil || !strings.Contains(err.Error(), c.path) {
			t.Errorf("%s: Validate() = %v; want an error at %s", c.name, err, c.path)
		}
	}
}

// TestValidateRejectsBroken210 checks the same for SARIF 2.1.0.
func TestValidateRejectsBroken210(t *testing.T) {
	rs, fs, root := scan(t, "It will work.\n")
	r, err := build(root, rs, fs)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Validate(); err != nil {
		t.Fatalf("the valid report fails: %v", err)
	}
	r.Runs[0].Results[0].Level = "fatal"
	if err := r.Validate(); err == nil || !strings.Contains(err.Error(), "runs.0.results.0.level") {
		t.Fatalf("Validate() = %v; want an error at the level", err)
	}
	r.Runs[0].Results[0].Level = "error"
	r.Version = "2.2"
	if err := (*v210.Report).Validate(r); err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("Validate() = %v; want an error at the version", err)
	}
}

// TestSchema22NeedsNoDraft202012 is the second half of the proof. The
// validator of go-sarif, gojsonschema, reads JSON Schema drafts up to 07.
// The 2.2 schema declares draft 2020-12. `Validate` is sound for this
// schema for two reasons. The schema uses no keyword that only draft
// 2019-09 or 2020-12 has. No $ref has a sibling constraint, which draft 07
// ignores. A new go-sarif version with a changed schema fails this test.
func TestSchema22NeedsNoDraft202012(t *testing.T) {
	out, err := exec.Command("go", "list", "-m", "-f", "{{.Dir}}", "github.com/owenrumney/go-sarif/v3").Output()
	if err != nil {
		t.Fatalf("go list: %v", err)
	}
	src, err := os.ReadFile(filepath.Join(strings.TrimSpace(string(out)), "pkg", "report", "v22", "sarif", "schema.go"))
	if err != nil {
		t.Fatal(err)
	}
	s := string(src)
	var schema any
	if err := json.Unmarshal([]byte(s[strings.Index(s, "`")+1:strings.LastIndex(s, "`")]), &schema); err != nil {
		t.Fatal(err)
	}
	newer := map[string]bool{
		"$anchor": true, "$dynamicRef": true, "$dynamicAnchor": true, "$recursiveRef": true, "$recursiveAnchor": true,
		"$vocabulary": true, "prefixItems": true, "unevaluatedItems": true, "unevaluatedProperties": true,
		"dependentRequired": true, "dependentSchemas": true, "minContains": true, "maxContains": true,
	}
	notes := map[string]bool{"$ref": true, "description": true, "default": true, "title": true, "$comment": true}
	refs := 0
	stack := []any{schema}
	for len(stack) > 0 {
		node := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		switch n := node.(type) {
		case []any:
			stack = append(stack, n...)
		case map[string]any:
			_, hasRef := n["$ref"]
			for k, v := range n {
				if newer[k] {
					t.Errorf("the 2.2 schema uses %s, which draft 07 does not have", k)
				}
				if hasRef && !notes[k] {
					t.Errorf("a $ref has the sibling %s, which draft 07 ignores", k)
				}
				stack = append(stack, v)
			}
			if hasRef {
				refs++
			}
		}
	}
	if refs < 100 {
		t.Fatalf("found %d $ref; the schema did not load", refs)
	}
}

func TestArtifactURI(t *testing.T) {
	root := t.TempDir()
	cases := map[string]string{
		filepath.Join(root, "docs", "a b.md"):         "docs/a%20b.md",
		filepath.Join(root, "a:b.md"):                 "./a:b.md",
		filepath.Join(filepath.Dir(root), "other.md"): "../other.md",
		filepath.Join(root, "x", "..", "y", "%25.md"): "y/%2525.md",
		filepath.Join(root, "docs", "été.md"):         "docs/%C3%A9t%C3%A9.md",
	}
	for p, want := range cases {
		got, err := artifactURI(root, p)
		if err != nil || got != want {
			t.Errorf("artifactURI(%q) = %q, %v; want %q", p, got, err, want)
		}
	}
}

func TestMatchCase(t *testing.T) {
	cases := [][3]string{{"Utilize", "use", "Use"}, {"utilize", "use", "use"}, {"ÉTÉ", "été", "Été"}, {"Therefore", "so", "So"}}
	for _, c := range cases {
		if got := matchCase(c[0], c[1]); got != c[2] {
			t.Errorf("matchCase(%q, %q) = %q; want %q", c[0], c[1], got, c[2])
		}
	}
}
