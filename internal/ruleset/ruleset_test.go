package ruleset

import (
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

func load(t *testing.T, o Options) *Ruleset {
	t.Helper()
	rs, err := Load(o)
	if err != nil {
		t.Fatal(err)
	}
	return rs
}

func loadErr(t *testing.T, o Options, want string) {
	t.Helper()
	_, err := Load(o)
	if err == nil || !strings.Contains(err.Error(), want) {
		t.Fatalf("Load(%+v) error = %v, want %q", o, err, want)
	}
}

func ste(files ...string) Options {
	return Options{Rulesets: append([]string{"ste-100"}, files...)}
}

func TestBuiltin(t *testing.T) {
	rs := load(t, ste())
	if rs.Name != "ste-100" || rs.Description == "" {
		t.Fatalf("name = %q, description = %q", rs.Name, rs.Description)
	}
	var ids []string
	for _, r := range rs.Active() {
		ids = append(ids, r.ID+" "+r.Severity)
	}
	want := []string{
		"STE-100-01 error", "STE-100-02 error", "STE-100-03 warning", "STE-100-04 error",
		"STE-100-05 warning", "STE-100-06 warning", "STE-100-07 warning", "STE-100-08 warning",
		"STE-100-09 warning", "STE-100-10 warning", "STE-100-11 warning",
	}
	if !slices.Equal(ids, want) {
		t.Fatalf("active rules = %q", ids)
	}
	r1 := rs.Rules["STE-100-01"]
	if r1.Params.MaxWords != 20 || !slices.Equal(r1.Params.ProceduralMarkers, []string{"numbered"}) || r1.Scope != "procedural_sentence" {
		t.Fatalf("STE-100-01 = %+v", r1)
	}
	if got := rs.Rules["STE-100-02"].Params.MaxWords; got != 25 {
		t.Fatalf("STE-100-02 max_words = %d", got)
	}
}

// TestLexiconProjection checks that STE-100-05 and STE-100-06 come from the
// lexicon and nothing else.
func TestLexiconProjection(t *testing.T) {
	rs := load(t, ste())
	swap, noun := map[string]string{}, map[string]string{}
	for _, c := range rs.Lexicon {
		for _, u := range c.Unapproved {
			swap[u] = c.Approved
		}
		for _, n := range c.Nominalizations {
			noun[n] = c.Approved
		}
	}
	if got := rs.Rules["STE-100-05"].Params.Swap; !maps.Equal(got, swap) || got["utilize"] != "use" {
		t.Fatalf("STE-100-05 swap = %v, want %v", got, swap)
	}
	if got := rs.Rules["STE-100-06"].Params.Swap; !maps.Equal(got, noun) || got["verification"] != "verify" {
		t.Fatalf("STE-100-06 swap = %v, want %v", got, noun)
	}
	// A user concept changes both rules.
	rs = load(t, ste("testdata/lexicon.cue"))
	if rs.Rules["STE-100-05"].Params.Swap["apprise"] != "tell" || rs.Rules["STE-100-06"].Params.Swap["apprisal"] != "tell" {
		t.Fatal("a lexicon concept of a user file does not reach STE-100-05 and STE-100-06")
	}
}

// TestCoreMatchesOracle reads the limits, the banned words and the passive
// pattern of scripts/ste-check.py and compares them with rules 01 to 04.
func TestCoreMatchesOracle(t *testing.T) {
	src, err := os.ReadFile("../../scripts/ste-check.py")
	if err != nil {
		t.Fatal(err)
	}
	text := string(src)
	num := func(name string) int {
		m := regexp.MustCompile(`(?m)^` + name + ` = (\d+)$`).FindStringSubmatch(text)
		if m == nil {
			t.Fatalf("ste-check.py has no %s", name)
		}
		n, _ := strconv.Atoi(m[1])
		return n
	}
	rs := load(t, ste())
	if num("MAX_PROCEDURAL") != rs.Rules["STE-100-01"].Params.MaxWords || num("MAX_DESCRIPTIVE") != rs.Rules["STE-100-02"].Params.MaxWords {
		t.Fatal("the word limits differ from ste-check.py")
	}
	block := regexp.MustCompile(`(?s)BANNED = \[(.*?)\]`).FindStringSubmatch(text)
	if block == nil {
		t.Fatal("ste-check.py has no BANNED list")
	}
	var banned []string
	for _, m := range regexp.MustCompile(`"([^"]+)"`).FindAllStringSubmatch(block[1], -1) {
		banned = append(banned, m[1])
	}
	tokens := slices.Sorted(maps.Keys(rs.Rules["STE-100-04"].Params.Tokens))
	slices.Sort(banned)
	if len(banned) != 13 || !slices.Equal(tokens, banned) {
		t.Fatalf("STE-100-04 tokens = %q, ste-check.py BANNED = %q", tokens, banned)
	}
	passive := regexp.MustCompile(`PASSIVE_RE = re.compile\(r"\\b\(([a-z|]+)\)\\s\+\\w\+(\w+)\\b"`).FindStringSubmatch(text)
	if passive == nil {
		t.Fatal("ste-check.py has no PASSIVE_RE of the known form")
	}
	r3 := rs.Rules["STE-100-03"].Params
	aux := strings.Split(passive[1], "|")
	slices.Sort(aux)
	if got := slices.Sorted(maps.Keys(r3.Auxiliaries)); !slices.Equal(got, aux) {
		t.Fatalf("STE-100-03 auxiliaries = %q, ste-check.py = %q", got, aux)
	}
	if got := slices.Sorted(maps.Keys(r3.Suffixes)); !slices.Equal(got, []string{passive[2]}) {
		t.Fatalf("STE-100-03 suffixes = %q, ste-check.py = %q", got, passive[2])
	}
	if rs.Rules["STE-100-03"].Severity != "warning" || rs.Rules["STE-100-04"].Severity != "error" {
		t.Fatal("the core severities differ from ste-check.py")
	}
}

func TestUserFiles(t *testing.T) {
	rs := load(t, ste("testdata/add_rule.cue"))
	if r := rs.Rules["DOC-CUSTOM-01"]; !r.Enabled || r.Params.Pattern != `\bTODO\b` || r.ID != "DOC-CUSTOM-01" {
		t.Fatalf("DOC-CUSTOM-01 = %+v", r)
	}
	rs = load(t, ste("testdata/omit_rule.cue"))
	if rs.Rules["STE-100-07"].Enabled || len(rs.Active()) != 10 {
		t.Fatal("omit_rule.cue does not turn off STE-100-07")
	}
	rs = load(t, ste("testdata/max_words.cue"))
	if rs.Rules["STE-100-01"].Params.MaxWords != 15 {
		t.Fatal("max_words.cue does not change STE-100-01")
	}
	rs = load(t, ste("testdata/vocabulary.cue"))
	if !rs.Vocabulary["okfw"] || !rs.Vocabulary["zmake"] {
		t.Fatalf("vocabulary = %v", rs.Vocabulary)
	}
	rs = load(t, ste("testdata/severity.cue"))
	if rs.Rules["STE-100-05"].Severity != "error" {
		t.Fatal("severity.cue does not change STE-100-05")
	}
	rs = load(t, Options{Rulesets: []string{"testdata/custom"}})
	if rs.Name != "house-style" || len(rs.Rules) != 2 || rs.Rules["HS-02"].Params.Tokens["etc."] != true {
		t.Fatalf("custom ruleset = %+v", rs)
	}
}

func TestOrderDoesNotMatter(t *testing.T) {
	files := []string{"ste-100", "testdata/add_rule.cue", "testdata/omit_rule.cue", "testdata/max_words.cue", "testdata/lexicon.cue"}
	a := load(t, Options{Rulesets: files})
	rev := slices.Clone(files)
	slices.Reverse(rev)
	b := load(t, Options{Rulesets: rev})
	if !reflect.DeepEqual(a, b) {
		t.Fatal("the order of the files changes the result")
	}
}

func TestFlags(t *testing.T) {
	o := ste("testdata/omit_rule.cue")
	o.Omit = []string{"STE-100-09"}
	o.Overrides = []Override{
		{RuleID: "STE-100-01", Param: "max_words", Value: "18"},
		{RuleID: "STE-100-01", Param: "procedural_markers", Value: `["numbered", "bullet"]`},
		{RuleID: "STE-100-04", Param: "tokens", Value: `{"utilise": true}`},
		{RuleID: "STE-100-09", Param: "pattern", Value: `x+`},
	}
	o.AllowTerms = []string{"okfw"}
	rs := load(t, o)
	r1 := rs.Rules["STE-100-01"].Params
	if r1.MaxWords != 18 || !slices.Equal(r1.ProceduralMarkers, []string{"numbered", "bullet"}) {
		t.Fatalf("STE-100-01 params = %+v", r1)
	}
	if !rs.Rules["STE-100-04"].Params.Tokens["utilise"] || len(rs.Rules["STE-100-04"].Params.Tokens) != 14 {
		t.Fatal("--set does not add a banned term")
	}
	if rs.Rules["STE-100-09"].Enabled || rs.Rules["STE-100-09"].Params.Pattern != "x+" || !rs.Vocabulary["okfw"] {
		t.Fatal("--omit-rule, a string --set or --allow-term has no effect")
	}
	o = ste()
	o.Enable = []string{"STE-100-07"}
	if !load(t, o).Rules["STE-100-07"].Enabled {
		t.Fatal("--enable-rule has no effect")
	}
}

func TestErrors(t *testing.T) {
	cases := []struct {
		o    Options
		want string
	}{
		{ste("testdata/bad_conflict.cue"), "conflict.utilize"},
		{ste("testdata/bad_banned.cue"), `approved and also banned by STE-100-04`},
		{ste("testdata/bad_typo.cue"), "max_wrods: field not allowed"},
		{ste("testdata/bad_check.cue"), `rules."STE-100-01".check`},
		{ste("testdata/bad_required.cue"), "pattern"},
		{ste("testdata/bad_id.cue"), "bad id"},
		{ste("testdata/bad_syntax.cue"), "testdata/bad_syntax.cue"},
		{ste("testdata/max_words.cue", "testdata/max_words_16.cue"), "conflicting values 16 and 15"},
		{ste("testdata/notcue.txt"), "does not end in .cue"},
		{ste("testdata/missing.cue"), "not a built-in ruleset (ste-100)"},
		{Options{Rulesets: []string{t.TempDir()}}, "the directory has 0 CUE files"},
		{Options{Rulesets: []string{"testdata/custom"}, Omit: []string{"STE-100-01"}}, "no rule STE-100-01"},
		{Options{Rulesets: []string{"ste-100"}, Overrides: []Override{{"STE-100-01", "max_words", "many"}}}, "conflicting values"},
		{Options{Rulesets: []string{"ste-100"}, Overrides: []Override{{"STE-100-01", "words", "3"}}}, "field not allowed"},
		{Options{Rulesets: []string{"ste-100"}, AllowTerms: []string{"should"}}, "conflict.should"},
		{Options{Rulesets: []string{"ste-100", "testdata/omit_rule.cue"}, Enable: []string{"STE-100-07"}}, "conflicting values"},
	}
	for _, c := range cases {
		loadErr(t, c.o, c.want)
	}
}

func TestFileSizeLimit(t *testing.T) {
	p := filepath.Join(t.TempDir(), "big.cue")
	if err := os.WriteFile(p, []byte("// "+strings.Repeat("x", MaxFileBytes)+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	loadErr(t, Options{Rulesets: []string{p}}, "more than")
}

func TestBuiltins(t *testing.T) {
	if got := Builtins(); !slices.Equal(got, []string{"ste-100"}) {
		t.Fatalf("Builtins() = %v", got)
	}
}
