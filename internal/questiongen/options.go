package questiongen

import (
	"math/rand/v2"
	"strconv"
	"strings"
)

// ── Number formatting ───────────────────────────────────────────────────

// fmtInt formats with thousands separators: 1232 → "1,232", -4500 → "-4,500".
func fmtInt(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.Itoa(n)
	if len(s) >= 4 {
		var b strings.Builder
		pre := len(s) % 3
		if pre > 0 {
			b.WriteString(s[:pre])
		}
		for i := pre; i < len(s); i += 3 {
			if b.Len() > 0 {
				b.WriteByte(',')
			}
			b.WriteString(s[i : i+3])
		}
		s = b.String()
	}
	if neg {
		return "-" + s
	}
	return s
}

// fmtPlain formats without separators. Used wherever numbers appear in a
// comma-separated list ("512, 1000, 983"), where "1,000" would read as two
// numbers — and for those questions' options, so both look the same.
func fmtPlain(n int) string { return strconv.Itoa(n) }

// fmtHalf formats a value stored in halves: 165 → "82.5", 60 → "30".
func fmtHalf(halves int) string {
	if halves%2 == 0 {
		return fmtInt(halves / 2)
	}
	return fmtInt(halves/2) + ".5"
}

// ── Numeric distractors ─────────────────────────────────────────────────

// numDistractors returns candidate wrong answers for a numeric answer.
//
// mistakes are template-specific errors (the result of a classic slip, e.g.
// evaluating left to right) and come first, shuffled. Generic near-misses
// follow, also shuffled, and are spread on both sides of the answer so the
// right option is not simply "the middle one". Values are formatted with
// format; negative values are dropped when the answer is non-negative.
func numDistractors(r *rand.Rand, ans int, format func(int) string, mistakes ...int) []string {
	keep := func(v int) bool { return v != ans && (ans < 0 || v >= 0) }

	var first []int
	for _, m := range mistakes {
		if keep(m) {
			first = append(first, m)
		}
	}
	r.Shuffle(len(first), func(i, j int) { first[i], first[j] = first[j], first[i] })

	abs := ans
	if abs < 0 {
		abs = -abs
	}
	step := 1
	switch {
	case abs >= 2000:
		step = 100
	case abs >= 200:
		step = 10
	case abs >= 30:
		step = between(r, 1, 3)
	}
	generic := []int{ans + step, ans - step, ans + 2*step, ans - 2*step, ans + 3*step}
	if abs >= 20 {
		generic = append(generic, ans+10, ans-10)
	}
	if sw, ok := swapLastDigits(ans); ok {
		generic = append(generic, sw)
	}
	if abs >= 50 {
		generic = append(generic, ans+abs/10, ans-abs/10)
	}
	r.Shuffle(len(generic), func(i, j int) { generic[i], generic[j] = generic[j], generic[i] })

	out := make([]string, 0, len(first)+len(generic))
	for _, v := range append(first, generic...) {
		if keep(v) {
			out = append(out, format(v))
		}
	}
	return out
}

// swapLastDigits turns 53 into 35, 1247 into 1274 — a common slip.
func swapLastDigits(n int) (int, bool) {
	neg := n < 0
	if neg {
		n = -n
	}
	if n < 10 {
		return 0, false
	}
	last, prev := n%10, (n/10)%10
	if last == prev {
		return 0, false
	}
	out := n - last - prev*10 + prev + last*10
	if neg {
		out = -out
	}
	return out, true
}

func gcd(a, b int) int {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

func lcm(a, b int) int { return a / gcd(a, b) * b }

func isPrime(n int) bool {
	if n < 2 {
		return false
	}
	for i := 2; i*i <= n; i++ {
		if n%i == 0 {
			return false
		}
	}
	return true
}

func isSquare(n int) bool {
	if n < 0 {
		return false
	}
	for i := 0; i*i <= n; i++ {
		if i*i == n {
			return true
		}
	}
	return false
}

func isCube(n int) bool {
	for i := 0; i*i*i <= n; i++ {
		if i*i*i == n {
			return true
		}
	}
	return false
}

// joinNums renders a series: "2, 5, 8, 11" (no thousands separators).
func joinNums(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = fmtPlain(x)
	}
	return strings.Join(parts, ", ")
}

// listAnd renders "12, 18 and 27".
func listAnd(xs []int) string {
	if len(xs) == 1 {
		return fmtInt(xs[0])
	}
	return joinNums(xs[:len(xs)-1]) + " and " + fmtPlain(xs[len(xs)-1])
}
