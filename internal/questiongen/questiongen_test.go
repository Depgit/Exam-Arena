package questiongen

import (
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

const perLevel = 3000

// Every template, every supported level, thousands of seeds: the structural
// guarantees must always hold.
func TestTemplatesProduceValidQuestions(t *testing.T) {
	for _, cat := range Categories() {
		for _, tpl := range Templates(cat) {
			for _, d := range tpl.Levels {
				t.Run(tpl.Key+"/"+string(d), func(t *testing.T) {
					ok, rejected := 0, 0
					for seed := uint64(1); seed <= perLevel; seed++ {
						q, err := GenerateWith(tpl, d, seed)
						if err == errReject {
							rejected++
							continue
						}
						if err != nil {
							t.Fatalf("seed %d: %v", seed, err)
						}
						checkQuestion(t, q)
						ok++
					}
					// A template that rejects most seeds would starve the pool.
					if rejected > perLevel*3/4 {
						t.Errorf("rejects too often: %d/%d", rejected, perLevel)
					}
				})
			}
		}
	}
}

func checkQuestion(t *testing.T, q Question) {
	t.Helper()
	if strings.TrimSpace(q.Body) == "" || strings.TrimSpace(q.Explanation) == "" {
		t.Fatalf("%s seed %d: blank body or explanation", q.TemplateKey, q.Seed)
	}
	if len(q.Options) != OptionCount {
		t.Fatalf("%s seed %d: %d options", q.TemplateKey, q.Seed, len(q.Options))
	}
	seen := map[string]bool{}
	for _, o := range q.Options {
		if strings.TrimSpace(o) == "" {
			t.Fatalf("%s seed %d: blank option in %q", q.TemplateKey, q.Seed, q.Options)
		}
		if seen[o] {
			t.Fatalf("%s seed %d: duplicate option %q in %q", q.TemplateKey, q.Seed, o, q.Options)
		}
		seen[o] = true
	}
	if q.Correct < 0 || q.Correct >= OptionCount {
		t.Fatalf("%s seed %d: correct index %d", q.TemplateKey, q.Seed, q.Correct)
	}
	if want, ok := independentAnswer(q.Body); ok && q.Options[q.Correct] != want {
		t.Fatalf("%s seed %d: %q marked %q, independently computed %q",
			q.TemplateKey, q.Seed, q.Body, q.Options[q.Correct], want)
	}
}

var (
	reBinary = regexp.MustCompile(`^What is ([\d,]+) ([+−×÷]) ([\d,]+)\?$`)
	reClock  = regexp.MustCompile(`at (\d+):(\d+)\?$`)
	reCal    = regexp.MustCompile(`^(.+) falls on a \w+\. On which day of the week does (.+) fall\?$`)
)

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.ReplaceAll(s, ",", ""))
	return n
}

// independentAnswer re-solves the question from its text alone, without
// the template's code, for the families where that is mechanical.
func independentAnswer(body string) (string, bool) {
	if m := reBinary.FindStringSubmatch(body); m != nil {
		a, b := atoi(m[1]), atoi(m[3])
		switch m[2] {
		case "+":
			return fmtInt(a + b), true
		case "−":
			return fmtInt(a - b), true
		case "×":
			return fmtInt(a * b), true
		case "÷":
			if a%b != 0 {
				return "not whole", true
			}
			return fmtInt(a / b), true
		}
	}
	if m := reClock.FindStringSubmatch(body); m != nil {
		h, mm := float64(atoi(m[1])), float64(atoi(m[2]))
		ang := 30*h + 0.5*mm - 6*mm
		for ang < 0 {
			ang += 360
		}
		for ang >= 360 {
			ang -= 360
		}
		if ang > 180 {
			ang = 360 - ang
		}
		return strings.TrimSuffix(strconv.FormatFloat(ang, 'f', 1, 64), ".0") + "°", true
	}
	if m := reCal.FindStringSubmatch(body); m != nil {
		base, err1 := time.Parse("2 January 2006", m[1])
		target, err2 := time.Parse("2 January 2006", m[2])
		if err1 != nil || err2 != nil {
			return "unparseable date", true
		}
		_ = base
		return target.Weekday().String(), true
	}
	return "", false
}

func TestGenerateIsDeterministic(t *testing.T) {
	for _, cat := range Categories() {
		for _, d := range Difficulties {
			for seed := uint64(1); seed <= 200; seed++ {
				a, err := Generate(cat, d, seed)
				if err != nil {
					t.Fatalf("%s/%s seed %d: %v", cat, d, seed, err)
				}
				b, _ := Generate(cat, d, seed)
				if a.Body != b.Body || strings.Join(a.Options, "|") != strings.Join(b.Options, "|") || a.Correct != b.Correct {
					t.Fatalf("%s/%s seed %d not deterministic", cat, d, seed)
				}
				// A stored question can be rebuilt from its template key and seed.
				tpl, ok := TemplateByKey(a.TemplateKey)
				if !ok {
					t.Fatalf("unknown template %s", a.TemplateKey)
				}
				c, err := GenerateWith(tpl, d, a.Seed)
				if err != nil || c.Body != a.Body || c.Options[c.Correct] != a.Options[a.Correct] {
					t.Fatalf("%s seed %d: rebuild mismatch", a.TemplateKey, a.Seed)
				}
			}
		}
	}
}

// Each level must offer several times more distinct questions than the
// pool keeps per level (QUESTION_GEN_POOL_SIZE / 3 = 200 by default).
func TestVariety(t *testing.T) {
	for _, cat := range Categories() {
		for _, d := range Difficulties {
			bodies := map[string]bool{}
			for seed := uint64(1); seed <= 2000; seed++ {
				q, err := Generate(cat, d, seed)
				if err != nil {
					t.Fatal(err)
				}
				bodies[q.Body] = true
			}
			if len(bodies) < 600 {
				t.Errorf("%s/%s: only %d distinct questions in 2000 seeds", cat, d, len(bodies))
			}
		}
	}
}

// The correct answer should land in each position about equally often.
func TestAnswerPositionIsBalanced(t *testing.T) {
	counts := make([]int, OptionCount)
	n := 0
	for _, cat := range Categories() {
		for _, d := range Difficulties {
			for seed := uint64(1); seed <= 2000; seed++ {
				q, _ := Generate(cat, d, seed)
				counts[q.Correct]++
				n++
			}
		}
	}
	for i, c := range counts {
		if share := float64(c) / float64(n); share < 0.20 || share > 0.30 {
			t.Errorf("answer at position %d %.1f%% of the time", i, share*100)
		}
	}
}

// A thousands separator inside a comma-separated list reads as two numbers.
func TestListsHaveNoThousandsSeparators(t *testing.T) {
	list := regexp.MustCompile(`\d+, \d{3}\b`)
	for _, d := range Difficulties {
		for seed := uint64(1); seed <= 3000; seed++ {
			q, _ := Generate(CategoryReasoning, d, seed)
			for _, s := range append([]string{q.Body}, q.Options...) {
				if strings.Contains(s, ",") && list.MatchString(s) && hasGroupedNumber(s) {
					t.Fatalf("%s seed %d: grouped number in a list: %q", q.TemplateKey, q.Seed, s)
				}
			}
		}
	}
}

// hasGroupedNumber reports "1,000"-style tokens (digits, comma, exactly 3 digits, no space).
func hasGroupedNumber(s string) bool {
	return regexp.MustCompile(`\d,\d{3}\b`).MatchString(s)
}

func TestFmtInt(t *testing.T) {
	for in, want := range map[int]string{0: "0", 999: "999", 1000: "1,000", 1232: "1,232", 123456: "123,456", -4500: "-4,500", 1234567: "1,234,567"} {
		if got := fmtInt(in); got != want {
			t.Errorf("fmtInt(%d) = %q, want %q", in, got, want)
		}
	}
}
