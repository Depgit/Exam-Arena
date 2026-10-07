package questiongen

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"
)

// Logical reasoning templates. Puzzles that could be read two ways are the
// main risk here, so odd-one-out and analogy questions are checked against
// every rule in their family and rejected if a second rule also fits.

var reasoningTemplates = []*Template{
	{Key: "reason.series_ap", Topic: "Number Series", Weight: 3, Levels: all, build: buildSeriesAP},
	{Key: "reason.series_gp", Topic: "Number Series", Weight: 2, Levels: all, build: buildSeriesGP},
	{Key: "reason.series_powers", Topic: "Number Series", Weight: 2, Levels: all, build: buildSeriesPowers},
	{Key: "reason.series_fib", Topic: "Number Series", Weight: 1, Levels: all, build: buildSeriesFib},
	{Key: "reason.series_diff", Topic: "Number Series", Weight: 2, Levels: []Difficulty{Medium, Hard}, build: buildSeriesDiff},
	{Key: "reason.series_alt", Topic: "Number Series", Weight: 2, Levels: []Difficulty{Medium, Hard}, build: buildSeriesAlt},
	{Key: "reason.letters", Topic: "Letter Series", Weight: 2, Levels: all, build: buildLetterSeries},
	{Key: "reason.coding", Topic: "Coding-Decoding", Weight: 3, Levels: all, build: buildCoding},
	{Key: "reason.odd_one_out", Topic: "Odd One Out", Weight: 3, Levels: all, build: buildOddOneOut},
	{Key: "reason.analogy", Topic: "Analogies", Weight: 2, Levels: all, build: buildAnalogy},
	{Key: "reason.direction", Topic: "Direction Sense", Weight: 3, Levels: all, build: buildDirection},
	{Key: "reason.calendar", Topic: "Calendar", Weight: 2, Levels: all, build: buildCalendar},
	{Key: "reason.clock", Topic: "Clocks", Weight: 2, Levels: all, build: buildClock},
}

// seriesDraft asks for the next term after shown.
func seriesDraft(r *rand.Rand, shown []int, ans int, explanation string, mistakes ...int) draft {
	return draft{
		body:        fmt.Sprintf("What comes next in the series: %s, ?", joinNums(shown)),
		answer:      fmtPlain(ans),
		distractors: numDistractors(r, ans, fmtPlain, mistakes...),
		explanation: explanation,
	}
}

func buildSeriesAP(r *rand.Rand, d Difficulty) (draft, error) {
	start := between(r, 1, level(d, 20, 50, 90))
	step := level(d, between(r, 2, 9), between(r, 11, 25), between(r, 13, 37))
	if d != Easy && r.IntN(3) == 0 {
		step = -step
		start += 6 * -step // stay positive
	}
	terms := make([]int, 6)
	for i := range terms {
		terms[i] = start + i*step
	}
	if d == Hard { // missing middle term instead of the next one
		gap := between(r, 1, 3) // never the first or last shown term
		shown := make([]string, len(terms)-1)
		for i := 0; i < len(terms)-1; i++ {
			shown[i] = fmtPlain(terms[i])
		}
		shown[gap] = "?"
		ans := terms[gap]
		return draft{
			body:        fmt.Sprintf("Find the missing number: %s", strings.Join(shown, ", ")),
			answer:      fmtPlain(ans),
			distractors: numDistractors(r, ans, fmtPlain, ans+1, ans-1, ans+step/2),
			explanation: fmt.Sprintf("Each term changes by %d.", step),
		}, nil
	}
	ans := terms[5]
	return seriesDraft(r, terms[:5], ans, fmt.Sprintf("Each term changes by %d.", step),
		ans+step, ans-1, ans+1), nil
}

func buildSeriesGP(r *rand.Rand, d Difficulty) (draft, error) {
	switch d {
	case Easy, Medium:
		ratio := level(d, 2, between(r, 3, 4), 0)
		a := between(r, 1, level(d, 6, 4, 0))
		terms := []int{a}
		for len(terms) < 6 {
			terms = append(terms, terms[len(terms)-1]*ratio)
		}
		ans := terms[5]
		last := terms[4]
		return seriesDraft(r, terms[:5], ans, fmt.Sprintf("Each term is multiplied by %d.", ratio),
			last+(last-terms[3]), ans+ratio, ans-ratio), nil
	default: // ×k + c
		k, c := between(r, 2, 3), pick(r, []int{-1, 1, 2, 3})
		terms := []int{between(r, 1, 5)}
		for len(terms) < 6 {
			terms = append(terms, terms[len(terms)-1]*k+c)
		}
		if terms[1] <= terms[0] {
			return draft{}, errReject
		}
		ans := terms[5]
		last := terms[4]
		op := fmt.Sprintf("+ %d", c)
		if c < 0 {
			op = fmt.Sprintf("− %d", -c)
		}
		return seriesDraft(r, terms[:5], ans, fmt.Sprintf("Each term is (previous × %d) %s.", k, op),
			last*k, last*k+c+1, last*k-c), nil
	}
}

func buildSeriesPowers(r *rand.Rand, d Difficulty) (draft, error) {
	n := between(r, 1, 8)
	var f func(int) int
	var rule string
	switch d {
	case Easy:
		f, rule = func(x int) int { return x * x }, "squares"
	case Medium:
		switch r.IntN(3) {
		case 0:
			f, rule = func(x int) int { return x * x * x }, "cubes"
		case 1:
			f, rule = func(x int) int { return x*x + 1 }, "n² + 1"
		default:
			f, rule = func(x int) int { return x*x - 1 }, "n² − 1"
		}
		n = between(r, 2, 6)
	default:
		if r.IntN(2) == 0 { // consecutive primes
			var primes []int
			for p := between(r, 2, 40); len(primes) < 6; p++ {
				if isPrime(p) {
					primes = append(primes, p)
				}
			}
			ans := primes[5]
			return seriesDraft(r, primes[:5], ans, "These are consecutive prime numbers.",
				ans+1, ans-1, ans+2, primes[4]+(primes[4]-primes[3])), nil
		}
		f, rule = func(x int) int { return x*x*x + 1 }, "n³ + 1"
		n = between(r, 1, 5)
	}
	terms := make([]int, 6)
	for i := range terms {
		terms[i] = f(n + i)
	}
	ans := terms[5]
	return seriesDraft(r, terms[:5], ans, fmt.Sprintf("The terms follow %s for consecutive numbers.", rule),
		terms[4]+(terms[4]-terms[3]), ans+1, ans-1), nil
}

func buildSeriesFib(r *rand.Rand, d Difficulty) (draft, error) {
	if d == Hard { // each term is the sum of the previous three
		terms := []int{between(r, 1, 3), between(r, 1, 4), between(r, 2, 6)}
		for len(terms) < 7 {
			n := len(terms)
			terms = append(terms, terms[n-1]+terms[n-2]+terms[n-3])
		}
		ans := terms[6]
		return seriesDraft(r, terms[:6], ans, "Each term is the sum of the previous three.",
			terms[5]+terms[4], ans+1, ans-1), nil
	}
	a := between(r, 1, level(d, 3, 10, 0))
	b := a + between(r, 0, level(d, 2, 8, 0))
	terms := []int{a, b}
	for len(terms) < 7 {
		n := len(terms)
		terms = append(terms, terms[n-1]+terms[n-2])
	}
	ans := terms[6]
	return seriesDraft(r, terms[:6], ans, "Each term is the sum of the previous two.",
		terms[5]+(terms[5]-terms[4]), ans+1, ans-1), nil
}

func buildSeriesDiff(r *rand.Rand, d Difficulty) (draft, error) {
	start := between(r, 1, 30)
	terms := []int{start}
	diff := between(r, 1, 5)
	inc := between(r, 1, 3)
	rule := fmt.Sprintf("The differences grow by %d each time.", inc)
	if d == Hard {
		rule = "The differences double each time."
	}
	for len(terms) < 6 {
		terms = append(terms, terms[len(terms)-1]+diff)
		if d == Hard {
			diff *= 2
		} else {
			diff += inc
		}
	}
	ans := terms[5]
	lastDiff := terms[4] - terms[3]
	return seriesDraft(r, terms[:5], ans, rule, terms[4]+lastDiff, ans+1, ans-1), nil
}

func buildSeriesAlt(r *rand.Rand, d Difficulty) (draft, error) {
	a, sa := between(r, 1, 20), between(r, 2, 9)
	b, sb := between(r, 20, 60), between(r, 2, 9)
	geo := d == Hard
	var terms []int
	x, y := a, b
	for len(terms) < 8 {
		terms = append(terms, x, y)
		x += sa
		if geo {
			y *= 2
		} else {
			y -= sb
		}
	}
	if !geo && terms[7] <= 0 {
		return draft{}, errReject
	}
	// Show 7 terms; the 8th belongs to the second interleaved series.
	ans := terms[7]
	rule := fmt.Sprintf("Two series alternate: one adds %d, the other subtracts %d.", sa, sb)
	if geo {
		rule = fmt.Sprintf("Two series alternate: one adds %d, the other doubles.", sa)
	}
	return seriesDraft(r, terms[:7], ans, rule, terms[6]+sa, terms[6]+1, ans+sb), nil
}

// ── Letters ─────────────────────────────────────────────────────────────

func letter(i int) string { return string(rune('A' + ((i%26)+26)%26)) }

func buildLetterSeries(r *rand.Rand, d Difficulty) (draft, error) {
	switch d {
	case Easy, Medium:
		step := level(d, between(r, 1, 3), between(r, 2, 5), 0)
		if d == Medium && r.IntN(2) == 0 {
			step = -step
		}
		start := between(r, 0, 25)
		end := start + 5*step
		if end < 0 || end > 25 { // no wrap-around: keep it fair
			return draft{}, errReject
		}
		var shown []string
		for i := 0; i < 5; i++ {
			shown = append(shown, letter(start+i*step))
		}
		ans := letter(end)
		var dis []string
		for _, off := range []int{1, -1, 2, -2, step} {
			if v := end + off; v >= 0 && v <= 25 {
				dis = append(dis, letter(v))
			}
		}
		r.Shuffle(len(dis), func(i, j int) { dis[i], dis[j] = dis[j], dis[i] })
		return draft{
			body:        fmt.Sprintf("What comes next: %s, ?", strings.Join(shown, ", ")),
			answer:      ans,
			distractors: dis,
			explanation: letterStepExplanation(step),
		}, nil
	default: // mirrored pairs: AZ, BY, CX, ...
		start := between(r, 0, 8)
		var shown []string
		for i := 0; i < 4; i++ {
			shown = append(shown, letter(start+i)+letter(25-start-i))
		}
		n := start + 4
		ans := letter(n) + letter(25-n)
		dis := []string{
			letter(25-n) + letter(n),     // swapped
			letter(n) + letter(25-n+1),   // second letter not advanced
			letter(n+1) + letter(25-n-1), // skipped one
			letter(n-1) + letter(25-n),   // first letter not advanced
		}
		r.Shuffle(len(dis), func(i, j int) { dis[i], dis[j] = dis[j], dis[i] })
		return draft{
			body:        fmt.Sprintf("What comes next: %s, ?", strings.Join(shown, ", ")),
			answer:      ans,
			distractors: dis,
			explanation: "The first letter moves forward from A, the second moves backward from Z.",
		}, nil
	}
}

func letterStepExplanation(step int) string {
	if step < 0 {
		return fmt.Sprintf("Each letter moves back %d place(s) in the alphabet.", -step)
	}
	return fmt.Sprintf("Each letter moves forward %d place(s) in the alphabet.", step)
}

var codeWords = []string{
	"CAT", "DOG", "SUN", "PEN", "CUP", "MAP", "BOOK", "LAMP", "TREE", "FISH", "MILK", "ROAD",
	"CAKE", "STAR", "RAIN", "GOLD", "SALT", "WIND", "FROG", "KITE", "NEST", "PARK", "DESK", "SHIP",
	"PLANT", "CHAIR", "RIVER", "TIGER", "BREAD", "STONE", "CLOUD", "HOUSE",
}

func shiftWord(w string, k int) string {
	var b strings.Builder
	for _, c := range w {
		b.WriteString(letter(int(c-'A') + k))
	}
	return b.String()
}

// nonZero avoids a shift of 0, which would offer the uncoded word itself.
func nonZero(k, fallback int) int {
	if k == 0 {
		return fallback
	}
	return k
}

func reverse(w string) string {
	rs := []rune(w)
	for i, j := 0, len(rs)-1; i < j; i, j = i+1, j-1 {
		rs[i], rs[j] = rs[j], rs[i]
	}
	return string(rs)
}

func buildCoding(r *rand.Rand, d Difficulty) (draft, error) {
	ex := pick(r, codeWords)
	target := pick(r, codeWords)
	if ex == target || len(target) < 3 {
		return draft{}, errReject
	}
	switch d {
	case Easy, Medium:
		k := level(d, pick(r, []int{1, -1}), pick(r, []int{2, 3, 4, -2, -3}), 0)
		ans := shiftWord(target, k)
		return draft{
			body: fmt.Sprintf("In a certain code, %s is written as %s. How is %s written in that code?",
				ex, shiftWord(ex, k), target),
			answer:      ans,
			distractors: []string{shiftWord(target, k+1), shiftWord(target, nonZero(k-1, k+2)), shiftWord(target, -k), reverse(ans)},
			explanation: fmt.Sprintf("Each letter is shifted %d place(s) in the alphabet.", k),
		}, nil
	default: // reverse, then shift
		k := pick(r, []int{1, 2, -1})
		code := func(w string) string { return shiftWord(reverse(w), k) }
		ans := code(target)
		return draft{
			body: fmt.Sprintf("In a certain code, %s is written as %s. How is %s written in that code?",
				ex, code(ex), target),
			answer:      ans,
			distractors: []string{shiftWord(target, k), reverse(target), shiftWord(reverse(target), -k), shiftWord(reverse(target), k+1)},
			explanation: fmt.Sprintf("The word is reversed and each letter is shifted %d place(s).", k),
		}, nil
	}
}

// ── Odd one out ─────────────────────────────────────────────────────────

type property struct {
	name string
	has  func(int) bool
}

var oddProperties = []property{
	{"even", func(n int) bool { return n%2 == 0 }},
	{"a multiple of 3", func(n int) bool { return n%3 == 0 }},
	{"a multiple of 4", func(n int) bool { return n%4 == 0 }},
	{"a multiple of 5", func(n int) bool { return n%5 == 0 }},
	{"a multiple of 6", func(n int) bool { return n%6 == 0 }},
	{"a multiple of 7", func(n int) bool { return n%7 == 0 }},
	{"a multiple of 8", func(n int) bool { return n%8 == 0 }},
	{"a multiple of 9", func(n int) bool { return n%9 == 0 }},
	{"a multiple of 11", func(n int) bool { return n%11 == 0 }},
	{"a perfect square", isSquare},
	{"a perfect cube", isCube},
	{"prime", isPrime},
}

// singledOut returns the index of the only option that differs from the
// rest under p (one has it and three don't, or the reverse), or -1.
func singledOut(nums []int, p property) int {
	var with, without []int
	for i, n := range nums {
		if p.has(n) {
			with = append(with, i)
		} else {
			without = append(without, i)
		}
	}
	switch {
	case len(with) == 1:
		return with[0]
	case len(without) == 1:
		return without[0]
	}
	return -1
}

func buildOddOneOut(r *rand.Rand, d Difficulty) (draft, error) {
	var rule property
	switch d {
	case Easy:
		rule = oddProperties[0] // even
	case Medium:
		rule = pick(r, []property{oddProperties[1], oddProperties[2], oddProperties[5], oddProperties[7], oddProperties[9]})
	default:
		rule = pick(r, []property{oddProperties[8], oddProperties[10], oddProperties[11]})
	}
	lo, hi := level(d, 10, 10, 10), level(d, 99, 150, 200)
	if rule.name == "a perfect cube" {
		hi = 1000 // 27…1000, otherwise only three cubes fit
	}

	// Most random sets are ambiguous (some other property also singles a
	// number out), so search for a clean set rather than rejecting the seed.
	for attempt := 0; attempt < 300; attempt++ {
		nums, odd, ok := oddOneOutCandidate(r, rule, lo, hi)
		if !ok {
			continue
		}
		ansIdx := indexOf(nums, odd)
		ambiguous := false
		for _, p := range oddProperties {
			if s := singledOut(nums, p); s >= 0 && s != ansIdx {
				ambiguous = true
				break
			}
		}
		if ambiguous {
			continue
		}
		var dis []string
		for _, n := range nums {
			if n != odd {
				dis = append(dis, fmtPlain(n))
			}
		}
		return draft{
			body:        fmt.Sprintf("Which number is the odd one out: %s?", joinNums(nums)),
			answer:      fmtPlain(odd),
			distractors: dis,
			explanation: fmt.Sprintf("All the others are %s; %d is not.", rule.name, odd),
		}, nil
	}
	return draft{}, errReject
}

// oddOneOutCandidate draws three numbers with the rule and one without.
func oddOneOutCandidate(r *rand.Rand, rule property, lo, hi int) ([]int, int, bool) {
	var group []int
	for tries := 0; len(group) < 3; tries++ {
		if tries > 2000 {
			return nil, 0, false
		}
		n := between(r, lo, hi)
		if rule.has(n) && !containsInt(group, n) {
			group = append(group, n)
		}
	}
	for tries := 0; tries < 2000; tries++ {
		n := between(r, lo, hi)
		// Primes vs "looks prime": make the odd one an odd composite.
		if rule.has(n) || (rule.name == "prime" && (n%2 == 0 || n%3 == 0 || n%5 == 0)) {
			continue
		}
		nums := append(append([]int{}, group...), n)
		r.Shuffle(len(nums), func(i, j int) { nums[i], nums[j] = nums[j], nums[i] })
		return nums, n, true
	}
	return nil, 0, false
}

func containsInt(xs []int, v int) bool { return indexOf(xs, v) >= 0 }

func indexOf(xs []int, v int) int {
	for i, x := range xs {
		if x == v {
			return i
		}
	}
	return -1
}

// ── Analogies ───────────────────────────────────────────────────────────

type relation struct {
	name string
	f    func(int) int
}

func analogyFamily() []relation {
	var fam []relation
	for k := 2; k <= 9; k++ {
		k := k
		fam = append(fam, relation{fmt.Sprintf("multiply by %d", k), func(x int) int { return x * k }})
		fam = append(fam, relation{fmt.Sprintf("add %d", k), func(x int) int { return x + k }})
	}
	fam = append(fam,
		relation{"square", func(x int) int { return x * x }},
		relation{"cube", func(x int) int { return x * x * x }},
		relation{"square plus 1", func(x int) int { return x*x + 1 }},
		relation{"square minus 1", func(x int) int { return x*x - 1 }},
		relation{"double plus 1", func(x int) int { return 2*x + 1 }},
		relation{"triple minus 1", func(x int) int { return 3*x - 1 }},
		relation{"n × (n + 1)", func(x int) int { return x * (x + 1) }},
		relation{"cube plus 1", func(x int) int { return x*x*x + 1 }},
	)
	return fam
}

func buildAnalogy(r *rand.Rand, d Difficulty) (draft, error) {
	fam := analogyFamily()
	var pool []relation
	for _, rel := range fam {
		simple := strings.HasPrefix(rel.name, "multiply") || strings.HasPrefix(rel.name, "add")
		hard := strings.Contains(rel.name, "cube") || strings.Contains(rel.name, "×")
		switch {
		case d == Easy && simple, d == Medium && !simple && !hard, d == Hard && hard:
			pool = append(pool, rel)
		}
	}
	rel := pick(r, pool)
	x1 := between(r, 2, 6)
	x2 := x1 + between(r, 1, 3)
	q := x2 + between(r, 1, 4)
	ans := rel.f(q)

	// Unambiguous: no other relation may fit both examples yet differ on q.
	var others []int
	for _, g := range fam {
		if g.name == rel.name {
			continue
		}
		if g.f(x1) == rel.f(x1) && g.f(x2) == rel.f(x2) && g.f(q) != ans {
			return draft{}, errReject
		}
		others = append(others, g.f(q))
	}
	// The best distractors are what nearby rules would give.
	r.Shuffle(len(others), func(i, j int) { others[i], others[j] = others[j], others[i] })
	var near []int
	for _, v := range others {
		if v > 0 && v != ans && v < ans*3 && v > ans/3 {
			near = append(near, v)
		}
	}
	return draft{
		body: fmt.Sprintf("Complete the analogy: %d : %d :: %d : %d :: %d : ?",
			x1, rel.f(x1), x2, rel.f(x2), q),
		answer:      fmtPlain(ans),
		distractors: numDistractors(r, ans, fmtPlain, near...),
		explanation: fmt.Sprintf("Rule: %s. %d → %d.", rel.name, q, ans),
	}, nil
}

// ── Direction sense ─────────────────────────────────────────────────────

var compass = []string{"North", "East", "South", "West"}

var walkers = []string{"Ravi", "Priya", "Arjun", "Meera", "Kabir", "Ananya", "Rohan", "Isha", "Vikram", "Sara"}

func buildDirection(r *rand.Rand, d Difficulty) (draft, error) {
	name := pick(r, walkers)
	start := r.IntN(4)
	switch d {
	case Easy: // which way is he facing after some turns
		turns := between(r, 2, 4)
		facing := start
		var steps []string
		for i := 0; i < turns; i++ {
			if r.IntN(2) == 0 {
				facing = (facing + 1) % 4
				steps = append(steps, "right")
			} else {
				facing = (facing + 3) % 4
				steps = append(steps, "left")
			}
		}
		ans := compass[facing]
		var dis []string
		for _, c := range compass {
			if c != ans {
				dis = append(dis, c)
			}
		}
		return draft{
			body: fmt.Sprintf("%s is facing %s and turns %s. Which direction is %s facing now?",
				name, compass[start], joinTurns(steps), name),
			answer:      ans,
			distractors: dis,
			explanation: fmt.Sprintf("Starting %s, turning %s ends facing %s.", compass[start], joinTurns(steps), ans),
		}, nil
	case Medium: // walk a, right b, right c, right b → |a−c| along the start axis
		a := between(r, 4, 15)
		c := between(r, 1, a-1)
		b := between(r, 2, 10)
		net := a - c
		ans := fmt.Sprintf("%d km %s", net, compass[start])
		return draft{
			body: fmt.Sprintf("%s walks %d km %s, turns right and walks %d km, turns right and walks %d km, then turns right and walks %d km. Where is %s now, relative to the start?",
				name, a, compass[start], b, c, b, name),
			answer: ans,
			distractors: []string{
				fmt.Sprintf("%d km %s", net, compass[(start+2)%4]),
				fmt.Sprintf("%d km %s", a+c, compass[start]),
				fmt.Sprintf("%d km %s", b, compass[(start+1)%4]),
				fmt.Sprintf("%d km %s", net+b, compass[start]),
			},
			explanation: fmt.Sprintf("The two %d km legs cancel out. Along %s: %d − %d = %d km.", b, compass[start], a, c, net),
		}, nil
	default: // right-angle walk → straight-line distance (Pythagorean triple)
		tri := pick(r, [][2]int{{3, 4}, {4, 3}, {5, 12}, {12, 5}, {6, 8}, {8, 6}})
		k := between(r, 1, 3)
		along, across := tri[0]*k, tri[1]*k
		c := between(r, 1, 8)
		a := along + c
		dist := 0
		for h := 1; ; h++ {
			if h*h == along*along+across*across {
				dist = h
				break
			}
		}
		km := func(v int) string { return fmt.Sprintf("%d km", v) }
		return draft{
			body: fmt.Sprintf("%s walks %d km %s, turns right and walks %d km, then turns right and walks %d km. How far is %s from the starting point?",
				name, a, compass[start], across, c, name),
			answer:      km(dist),
			distractors: numDistractors(r, dist, km, a+across+c, along+across, along, across),
			explanation: fmt.Sprintf("Net displacement is %d km and %d km at right angles, so √(%d² + %d²) = %d km.", along, across, along, across, dist),
		}, nil
	}
}

func joinTurns(steps []string) string {
	if len(steps) == 1 {
		return steps[0]
	}
	return strings.Join(steps[:len(steps)-1], ", then ") + ", then " + steps[len(steps)-1]
}

// ── Calendar ────────────────────────────────────────────────────────────

func buildCalendar(r *rand.Rand, d Difficulty) (draft, error) {
	year := between(r, 2015, 2030)
	base := time.Date(year, time.Month(between(r, 1, 12)), between(r, 1, 28), 12, 0, 0, 0, time.UTC)
	var target time.Time
	switch d {
	case Easy:
		target = base.AddDate(0, 0, between(r, 2, 27))
	case Medium:
		target = base.AddDate(0, 0, between(r, 30, 120))
	default:
		target = base.AddDate(between(r, 1, 3), 0, 0)
	}
	ans := target.Weekday().String()
	var dis []string
	for _, off := range []int{1, -1, 2, -2, 3} {
		dis = append(dis, time.Weekday((int(target.Weekday())+off+7)%7).String())
	}
	r.Shuffle(len(dis), func(i, j int) { dis[i], dis[j] = dis[j], dis[i] })
	days := int(target.Sub(base).Hours() / 24)
	return draft{
		// Tense-free wording: these dates are past or future depending on when it's read.
		body: fmt.Sprintf("%s falls on a %s. On which day of the week does %s fall?",
			base.Format("2 January 2006"), base.Weekday(), target.Format("2 January 2006")),
		answer:      ans,
		distractors: dis,
		explanation: fmt.Sprintf("%d days later: %d mod 7 = %d, so move %d day(s) forward from %s to %s.",
			days, days, days%7, days%7, base.Weekday(), ans),
	}, nil
}

// ── Clocks ──────────────────────────────────────────────────────────────

func buildClock(r *rand.Rand, d Difficulty) (draft, error) {
	h := between(r, 1, 12)
	m := level(d, 0, pick(r, []int{15, 30, 45}), 5*between(r, 1, 11))
	// Work in half-degrees: hour hand = 60h + m, minute hand = 12m.
	angle := (60*h + m) - 12*m
	if angle < 0 {
		angle = -angle
	}
	angle %= 720
	if angle > 360 {
		angle = 720 - angle
	}
	if angle == 0 || angle == 360 {
		return draft{}, errReject // hands overlap or are straight: too trivial
	}
	deg := func(halves int) string { return fmtHalf(halves) + "°" }
	naive := 60*h - 12*m // forgot the hour hand moves
	if naive < 0 {
		naive = -naive
	}
	naive %= 720
	if naive > 360 {
		naive = 720 - naive
	}
	return draft{
		body:   fmt.Sprintf("What is the smaller angle between the hands of a clock at %d:%02d?", h, m),
		answer: deg(angle),
		// Easy answers are whole multiples of 30°, so keep the wrong ones on that grid too.
		distractors: numDistractors(r, angle, deg, level(d,
			[]int{720 - angle, angle + 60, angle - 60, angle + 120},
			[]int{naive, 720 - angle, angle + 60, angle - 60, angle + 15},
			[]int{naive, 720 - angle, angle + 60, angle - 60, angle + 11})...),
		explanation: fmt.Sprintf("Angle = |30 × %d − 5.5 × %d| = %s.", h, m, deg(angle)),
	}, nil
}
