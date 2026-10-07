// Package questiongen procedurally generates multiple-choice questions.
//
// Every question is produced by a Template from a seed. The same
// (template key, difficulty, seed) always yields the same question, so a
// stored question can be rebuilt exactly when it is reported or disputed.
//
// The package is pure: no database, no clock, no global randomness. The
// pool keeper (internal/service/question_pool.go) decides how many to make
// and stores them as ordinary rows, which is what lets matches, practice,
// the daily challenge and flags use generated questions unchanged.
package questiongen

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"sort"
)

type Difficulty string

const (
	Easy   Difficulty = "easy"
	Medium Difficulty = "medium"
	Hard   Difficulty = "hard"
)

// Difficulties lists every level, in increasing order.
var Difficulties = []Difficulty{Easy, Medium, Hard}

// OptionCount is the number of choices every generated question has.
const OptionCount = 4

// Question is one generated multiple-choice question.
type Question struct {
	TemplateKey      string
	Topic            string
	Difficulty       Difficulty
	Seed             uint64
	Body             string
	Options          []string // display order
	Correct          int      // index into Options
	Explanation      string
	EstimatedSeconds int
}

// draft is what a template builds before options are deduplicated and shuffled.
type draft struct {
	body        string
	answer      string
	distractors []string // in priority order; the most instructive mistakes first
	explanation string
}

// Template generates one family of questions.
type Template struct {
	Key    string
	Topic  string
	Weight int          // relative frequency within its category
	Levels []Difficulty // levels this template supports
	build  func(r *rand.Rand, d Difficulty) (draft, error)
}

func (t *Template) supports(d Difficulty) bool {
	for _, l := range t.Levels {
		if l == d {
			return true
		}
	}
	return false
}

// errReject means "this seed produced an unusable question, try another".
// It is expected and never surfaces past Generate.
var errReject = errors.New("questiongen: rejected draft")

// Category codes, matching exam_categories.code.
const (
	CategoryMath      = "MATH"
	CategoryReasoning = "REASONING"
)

var registry = map[string][]*Template{
	CategoryMath:      mathTemplates,
	CategoryReasoning: reasoningTemplates,
}

// Categories returns the category codes the generator can produce.
func Categories() []string {
	out := make([]string, 0, len(registry))
	for c := range registry {
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// Templates returns the templates registered for a category.
func Templates(category string) []*Template {
	return registry[category]
}

// TemplateByKey finds a template in any category.
func TemplateByKey(key string) (*Template, bool) {
	for _, ts := range registry {
		for _, t := range ts {
			if t.Key == key {
				return t, true
			}
		}
	}
	return nil, false
}

// Topics returns the distinct topic names used by a category's templates.
func Topics(category string) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range registry[category] {
		if !seen[t.Topic] {
			seen[t.Topic] = true
			out = append(out, t.Topic)
		}
	}
	return out
}

// maxAttempts bounds retries when a seed yields an unusable draft.
const maxAttempts = 64

// Generate builds one question for the category and difficulty. The
// template is chosen by weight from the seed, so the result is fully
// determined by (category, d, seed).
func Generate(category string, d Difficulty, seed uint64) (Question, error) {
	var candidates []*Template
	total := 0
	for _, t := range registry[category] {
		if t.supports(d) {
			candidates = append(candidates, t)
			total += t.Weight
		}
	}
	if total == 0 {
		return Question{}, fmt.Errorf("questiongen: no templates for %s/%s", category, d)
	}

	for attempt := uint64(0); attempt < maxAttempts; attempt++ {
		s := mix(seed, attempt)
		pick := int(newRand(s).IntN(total))
		var t *Template
		for _, c := range candidates {
			if pick < c.Weight {
				t = c
				break
			}
			pick -= c.Weight
		}
		q, err := GenerateWith(t, d, s)
		if errors.Is(err, errReject) {
			continue
		}
		return q, err
	}
	return Question{}, fmt.Errorf("questiongen: no usable question for %s/%s after %d attempts", category, d, maxAttempts)
}

// GenerateWith builds a question from one specific template. This is how a
// stored generated question is rebuilt from its generator_key and seed.
func GenerateWith(t *Template, d Difficulty, seed uint64) (Question, error) {
	if !t.supports(d) {
		return Question{}, fmt.Errorf("questiongen: %s does not support %s", t.Key, d)
	}
	r := newRand(seed)
	dr, err := t.build(r, d)
	if err != nil {
		return Question{}, err
	}
	opts, correct, err := finalize(r, dr.answer, dr.distractors)
	if err != nil {
		return Question{}, err
	}
	return Question{
		TemplateKey:      t.Key,
		Topic:            t.Topic,
		Difficulty:       d,
		Seed:             seed,
		Body:             dr.body,
		Options:          opts,
		Correct:          correct,
		Explanation:      dr.explanation,
		EstimatedSeconds: estimatedSeconds[d],
	}, nil
}

var estimatedSeconds = map[Difficulty]int{Easy: 20, Medium: 40, Hard: 60}

// finalize picks the first OptionCount-1 usable distractors (distinct, not
// blank, not the answer) and shuffles them in with the answer.
func finalize(r *rand.Rand, answer string, distractors []string) ([]string, int, error) {
	if answer == "" {
		return nil, 0, errReject
	}
	seen := map[string]bool{answer: true}
	opts := make([]string, 0, OptionCount)
	for _, d := range distractors {
		if d == "" || seen[d] {
			continue
		}
		seen[d] = true
		opts = append(opts, d)
		if len(opts) == OptionCount-1 {
			break
		}
	}
	if len(opts) < OptionCount-1 {
		return nil, 0, errReject
	}
	opts = append(opts, answer)
	r.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
	for i, o := range opts {
		if o == answer {
			return opts, i, nil
		}
	}
	return nil, 0, errReject // unreachable
}

// newRand returns a deterministic generator for a seed.
func newRand(seed uint64) *rand.Rand {
	return rand.New(rand.NewPCG(seed, seed^0x9E3779B97F4A7C15))
}

// mix derives a well-spread child seed (splitmix64 finaliser), so attempt
// n of seed s never collides with attempt 0 of a nearby seed.
func mix(seed, n uint64) uint64 {
	z := seed + (n+1)*0x9E3779B97F4A7C15
	z = (z ^ (z >> 30)) * 0xBF58476D1CE4E5B9
	z = (z ^ (z >> 27)) * 0x94D049BB133111EB
	return z ^ (z >> 31)
}

// between returns a uniform int in [lo, hi].
func between(r *rand.Rand, lo, hi int) int {
	if hi <= lo {
		return lo
	}
	return lo + r.IntN(hi-lo+1)
}

func pick[T any](r *rand.Rand, xs []T) T {
	return xs[r.IntN(len(xs))]
}

// level picks the value for the difficulty from (easy, medium, hard).
func level[T any](d Difficulty, easy, medium, hard T) T {
	switch d {
	case Easy:
		return easy
	case Medium:
		return medium
	default:
		return hard
	}
}
