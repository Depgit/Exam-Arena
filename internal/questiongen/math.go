package questiongen

import (
	"fmt"
	"math/rand/v2"
)

// Math templates. Difficulty scales number size and the number of steps.
// Wherever an answer must be a whole number (division, equations, averages)
// the answer is chosen first and the question is built around it.

var all = []Difficulty{Easy, Medium, Hard}

var mathTemplates = []*Template{
	{Key: "math.add", Topic: "Arithmetic", Weight: 3, Levels: all, build: buildAdd},
	{Key: "math.sub", Topic: "Arithmetic", Weight: 3, Levels: all, build: buildSub},
	{Key: "math.mul", Topic: "Arithmetic", Weight: 4, Levels: all, build: buildMul},
	{Key: "math.div", Topic: "Arithmetic", Weight: 3, Levels: all, build: buildDiv},
	{Key: "math.bodmas", Topic: "Order of Operations", Weight: 3, Levels: all, build: buildBodmas},
	{Key: "math.percent", Topic: "Percentages", Weight: 3, Levels: all, build: buildPercent},
	{Key: "math.linear", Topic: "Equations", Weight: 3, Levels: all, build: buildLinear},
	{Key: "math.powers", Topic: "Squares & Roots", Weight: 2, Levels: all, build: buildPowers},
	{Key: "math.hcf_lcm", Topic: "HCF & LCM", Weight: 2, Levels: all, build: buildHcfLcm},
	{Key: "math.average", Topic: "Averages", Weight: 2, Levels: all, build: buildAverage},
	{Key: "math.speed", Topic: "Speed, Distance & Time", Weight: 2, Levels: all, build: buildSpeed},
	{Key: "math.interest", Topic: "Simple Interest", Weight: 2, Levels: []Difficulty{Medium, Hard}, build: buildInterest},
}

func buildAdd(r *rand.Rand, d Difficulty) (draft, error) {
	var a, b int
	switch d {
	case Easy:
		a, b = between(r, 10, 99), between(r, 10, 99)
	case Medium:
		a, b = between(r, 100, 999), between(r, 100, 999)
	default:
		a, b = between(r, 1000, 9999), between(r, 100, 9999)
	}
	ans := a + b
	return draft{
		body:        fmt.Sprintf("What is %s + %s?", fmtInt(a), fmtInt(b)),
		answer:      fmtInt(ans),
		distractors: numDistractors(r, ans, fmtInt, carrySlips(ans)...),
		explanation: fmt.Sprintf("%s + %s = %s.", fmtInt(a), fmtInt(b), fmtInt(ans)),
	}, nil
}

func buildSub(r *rand.Rand, d Difficulty) (draft, error) {
	var a, b int
	switch d {
	case Easy:
		a, b = between(r, 20, 99), between(r, 10, 19)
	case Medium:
		a, b = between(r, 300, 999), between(r, 100, 299)
	default:
		a, b = between(r, 3000, 9999), between(r, 1000, 2999)
	}
	ans := a - b
	return draft{
		body:        fmt.Sprintf("What is %s − %s?", fmtInt(a), fmtInt(b)),
		answer:      fmtInt(ans),
		distractors: numDistractors(r, ans, fmtInt, carrySlips(ans)...),
		explanation: fmt.Sprintf("%s − %s = %s.", fmtInt(a), fmtInt(b), fmtInt(ans)),
	}, nil
}

// carrySlips are dropped or extra carries/borrows. ±100 only reads as a
// believable slip once the answer is in the hundreds.
func carrySlips(ans int) []int {
	out := []int{ans + 10, ans - 10}
	if ans >= 200 {
		out = append(out, ans+100, ans-100)
	}
	return out
}

func buildMul(r *rand.Rand, d Difficulty) (draft, error) {
	var a, b int
	switch d {
	case Easy:
		a, b = between(r, 2, 12), between(r, 2, 20)
	case Medium:
		a, b = between(r, 11, 99), between(r, 11, 99)
	default:
		a, b = between(r, 101, 999), between(r, 11, 99)
	}
	ans := a * b
	return draft{
		body:   fmt.Sprintf("What is %s × %s?", fmtInt(a), fmtInt(b)),
		answer: fmtInt(ans),
		// One row too many / too few, and a carry slip that keeps the last digit.
		distractors: numDistractors(r, ans, fmtInt, ans+a, ans-a, ans+b, ans-b, ans+10, ans-10),
		explanation: fmt.Sprintf("%s × %s = %s.", fmtInt(a), fmtInt(b), fmtInt(ans)),
	}, nil
}

func buildDiv(r *rand.Rand, d Difficulty) (draft, error) {
	var div, q int
	switch d {
	case Easy:
		div, q = between(r, 2, 12), between(r, 2, 12)
	case Medium:
		div, q = between(r, 3, 25), between(r, 11, 99)
	default:
		div, q = between(r, 11, 99), between(r, 101, 999)
	}
	n := div * q
	return draft{
		body:        fmt.Sprintf("What is %s ÷ %s?", fmtInt(n), fmtInt(div)),
		answer:      fmtInt(q),
		distractors: numDistractors(r, q, fmtInt, q+1, q-1, q+10, q-10),
		explanation: fmt.Sprintf("%s × %s = %s, so %s ÷ %s = %s.", fmtInt(div), fmtInt(q), fmtInt(n), fmtInt(n), fmtInt(div), fmtInt(q)),
	}, nil
}

func buildBodmas(r *rand.Rand, d Difficulty) (draft, error) {
	switch d {
	case Easy: // a + b × c — the classic trap is (a + b) × c
		a, b, c := between(r, 2, 20), between(r, 2, 9), between(r, 2, 9)
		ans := a + b*c
		return draft{
			body:        fmt.Sprintf("What is %d + %d × %d?", a, b, c),
			answer:      fmtInt(ans),
			distractors: numDistractors(r, ans, fmtInt, (a+b)*c),
			explanation: fmt.Sprintf("Multiply first: %d × %d = %d, then %d + %d = %d.", b, c, b*c, a, b*c, ans),
		}, nil
	case Medium: // a × b − c × d
		a, b, c, e := between(r, 3, 15), between(r, 3, 15), between(r, 2, 9), between(r, 2, 9)
		ans := a*b - c*e
		if ans <= 0 {
			return draft{}, errReject
		}
		return draft{
			body:        fmt.Sprintf("What is %d × %d − %d × %d?", a, b, c, e),
			answer:      fmtInt(ans),
			distractors: numDistractors(r, ans, fmtInt, (a*b-c)*e, a*(b-c)*e),
			explanation: fmt.Sprintf("%d × %d = %d and %d × %d = %d, so %d − %d = %d.", a, b, a*b, c, e, c*e, a*b, c*e, ans),
		}, nil
	default: // (a + b) × c − d ÷ e, with d ÷ e exact
		a, b, c := between(r, 2, 20), between(r, 2, 20), between(r, 2, 9)
		e, q := between(r, 2, 9), between(r, 2, 12)
		n := e * q
		ans := (a+b)*c - q
		if ans <= 0 {
			return draft{}, errReject
		}
		mistakes := []int{a + b*c - q}
		if ((a+b)*c-n)%e == 0 {
			mistakes = append(mistakes, ((a+b)*c-n)/e) // left to right
		}
		return draft{
			body:        fmt.Sprintf("What is (%d + %d) × %d − %d ÷ %d?", a, b, c, n, e),
			answer:      fmtInt(ans),
			distractors: numDistractors(r, ans, fmtInt, mistakes...),
			explanation: fmt.Sprintf("Brackets: %d + %d = %d. Then %d × %d = %d and %d ÷ %d = %d. Finally %d − %d = %d.",
				a, b, a+b, a+b, c, (a+b)*c, n, e, q, (a+b)*c, q, ans),
		}, nil
	}
}

func buildPercent(r *rand.Rand, d Difficulty) (draft, error) {
	p := pick(r, level(d,
		[]int{10, 20, 25, 50},
		[]int{5, 15, 30, 40, 60, 75},
		[]int{12, 35, 45, 65, 85, 120, 150}))
	var y int
	for tries := 0; ; tries++ {
		y = level(d, 20*between(r, 1, 25), 20*between(r, 5, 50), 50*between(r, 4, 60))
		if p*y%100 == 0 {
			break
		}
		if tries > 20 {
			return draft{}, errReject
		}
	}
	ans := p * y / 100
	mistakes := []int{(100 - p) * y / 100}
	if p*y%10 == 0 {
		mistakes = append(mistakes, p*y/10) // decimal point slip
	}
	return draft{
		body:        fmt.Sprintf("What is %d%% of %s?", p, fmtInt(y)),
		answer:      fmtInt(ans),
		distractors: numDistractors(r, ans, fmtInt, mistakes...),
		explanation: fmt.Sprintf("%d%% of %s = %d × %s ÷ 100 = %s.", p, fmtInt(y), p, fmtInt(y), fmtInt(ans)),
	}, nil
}

// term renders "3x", "x", "-x" for a coefficient.
func term(a int) string {
	switch a {
	case 1:
		return "x"
	case -1:
		return "-x"
	}
	return fmt.Sprintf("%dx", a)
}

// withConst renders "3x + 5" / "3x − 5".
func withConst(lhs string, b int) string {
	switch {
	case b > 0:
		return fmt.Sprintf("%s + %d", lhs, b)
	case b < 0:
		return fmt.Sprintf("%s − %d", lhs, -b)
	}
	return lhs
}

func buildLinear(r *rand.Rand, d Difficulty) (draft, error) {
	switch d {
	case Easy, Medium: // ax + b = c
		var a, x, b int
		if d == Easy {
			a, x, b = between(r, 2, 5), between(r, 1, 12), between(r, 1, 20)
		} else {
			a, x, b = between(r, 3, 12), between(r, 2, 25), between(r, -30, 30)
			if b == 0 {
				b = 7
			}
		}
		c := a*x + b
		mistakes := []int{c - b}
		if (c+b)%a == 0 {
			mistakes = append(mistakes, (c+b)/a) // moved b with the wrong sign
		}
		return draft{
			body:        fmt.Sprintf("Solve for x: %s = %d", withConst(term(a), b), c),
			answer:      fmtInt(x),
			distractors: numDistractors(r, x, fmtInt, mistakes...),
			explanation: fmt.Sprintf("%s = %d − (%d) = %d, so x = %d ÷ %d = %d.", term(a), c, b, c-b, c-b, a, x),
		}, nil
	default: // ax + b = cx + e
		a := between(r, 4, 12)
		c := between(r, 1, a-1)
		x, b := between(r, 2, 20), between(r, 1, 40)
		e := (a-c)*x + b
		var mistakes []int
		if (e+b)%(a-c) == 0 {
			mistakes = append(mistakes, (e+b)/(a-c))
		}
		if (e-b)%(a+c) == 0 {
			mistakes = append(mistakes, (e-b)/(a+c))
		}
		return draft{
			body:        fmt.Sprintf("Solve for x: %s = %s", withConst(term(a), b), withConst(term(c), e)),
			answer:      fmtInt(x),
			distractors: numDistractors(r, x, fmtInt, mistakes...),
			explanation: fmt.Sprintf("Collect terms: %s = %d − %d = %d, so x = %d.", term(a-c), e, b, e-b, x),
		}, nil
	}
}

func buildPowers(r *rand.Rand, d Difficulty) (draft, error) {
	kind := r.IntN(2) // 0: power, 1: root
	switch d {
	case Easy, Medium:
		n := level(d, between(r, 4, 15), between(r, 16, 35), 0)
		sq := n * n
		if kind == 0 {
			return draft{
				body:        fmt.Sprintf("What is %d²?", n),
				answer:      fmtInt(sq),
				distractors: numDistractors(r, sq, fmtInt, (n+1)*(n+1), (n-1)*(n-1), 2*n, sq+10),
				explanation: fmt.Sprintf("%d² = %d × %d = %s.", n, n, n, fmtInt(sq)),
			}, nil
		}
		return draft{
			body:        fmt.Sprintf("What is √%s?", fmtInt(sq)),
			answer:      fmtInt(n),
			distractors: numDistractors(r, n, fmtInt, n+1, n-1, n+2, n-2),
			explanation: fmt.Sprintf("%d × %d = %s, so √%s = %d.", n, n, fmtInt(sq), fmtInt(sq), n),
		}, nil
	default:
		if kind == 0 { // cube
			n := between(r, 4, 15)
			cube := n * n * n
			return draft{
				body:        fmt.Sprintf("What is %d³?", n),
				answer:      fmtInt(cube),
				distractors: numDistractors(r, cube, fmtInt, (n+1)*(n+1)*(n+1), (n-1)*(n-1)*(n-1), n*n*3, cube+n*n),
				explanation: fmt.Sprintf("%d³ = %d × %d × %d = %s.", n, n, n, n, fmtInt(cube)),
			}, nil
		}
		n := between(r, 36, 99)
		sq := n * n
		return draft{
			body:        fmt.Sprintf("What is √%s?", fmtInt(sq)),
			answer:      fmtInt(n),
			distractors: numDistractors(r, n, fmtInt, n+1, n-1, n+10, n-10),
			explanation: fmt.Sprintf("%d × %d = %s, so √%s = %d.", n, n, fmtInt(sq), fmtInt(sq), n),
		}, nil
	}
}

// coprimePair returns distinct p, q in [lo, hi] with gcd 1.
func coprimePair(r *rand.Rand, lo, hi int) (int, int) {
	for {
		p, q := between(r, lo, hi), between(r, lo, hi)
		if p != q && gcd(p, q) == 1 {
			return p, q
		}
	}
}

func buildHcfLcm(r *rand.Rand, d Difficulty) (draft, error) {
	switch d {
	case Easy: // HCF of g·p and g·q
		g := between(r, 2, 9)
		p, q := coprimePair(r, 2, 9)
		a, b := g*p, g*q
		return draft{
			body:        fmt.Sprintf("What is the HCF of %d and %d?", a, b),
			answer:      fmtInt(g),
			distractors: numDistractors(r, g, fmtInt, lcm(a, b), g*2, g*p, g*q),
			explanation: fmt.Sprintf("%d = %d × %d and %d = %d × %d, so the HCF is %d.", a, g, p, b, g, q, g),
		}, nil
	case Medium: // LCM of two
		a, b := between(r, 4, 30), between(r, 4, 30)
		if a == b || gcd(a, b) == 1 && a*b > 400 {
			return draft{}, errReject
		}
		ans := lcm(a, b)
		return draft{
			body:        fmt.Sprintf("What is the LCM of %d and %d?", a, b),
			answer:      fmtInt(ans),
			distractors: numDistractors(r, ans, fmtInt, a*b, gcd(a, b), ans*2, ans+a, ans+b),
			explanation: fmt.Sprintf("HCF(%d, %d) = %d, so LCM = %d × %d ÷ %d = %d.", a, b, gcd(a, b), a, b, gcd(a, b), ans),
		}, nil
	default: // LCM of three
		a, b, c := between(r, 4, 20), between(r, 4, 20), between(r, 4, 20)
		if a == b || b == c || a == c {
			return draft{}, errReject
		}
		ans := lcm(lcm(a, b), c)
		if ans > 2000 {
			return draft{}, errReject
		}
		return draft{
			body:        fmt.Sprintf("What is the LCM of %d, %d and %d?", a, b, c),
			answer:      fmtInt(ans),
			distractors: numDistractors(r, ans, fmtInt, lcm(a, b), lcm(b, c), lcm(a, c), ans*2),
			explanation: fmt.Sprintf("LCM(%d, %d) = %d, and LCM(%d, %d) = %d.", a, b, lcm(a, b), lcm(a, b), c, ans),
		}, nil
	}
}

func buildAverage(r *rand.Rand, d Difficulty) (draft, error) {
	switch d {
	case Easy, Medium:
		n := level(d, 3, 5, 0)
		lo, hi := level(d, 10, 20, 0), level(d, 50, 99, 0)
		nums := make([]int, n)
		sum := 0
		for i := range nums {
			nums[i] = between(r, lo, hi)
			sum += nums[i]
		}
		// Nudge the last number so the average is whole.
		nums[n-1] += (n - sum%n) % n
		sum += (n - sum%n) % n
		avg := sum / n
		return draft{
			body:        fmt.Sprintf("What is the average of %s?", listAnd(nums)),
			answer:      fmtInt(avg),
			distractors: numDistractors(r, avg, fmtInt, sum, sum/(n-1), avg+n),
			explanation: fmt.Sprintf("Sum = %d, and %d ÷ %d = %d.", sum, sum, n, avg),
		}, nil
	default: // find the missing number
		avg := between(r, 20, 80)
		known := make([]int, 4)
		sumKnown := 0
		for i := range known {
			known[i] = between(r, avg-15, avg+15)
			sumKnown += known[i]
		}
		fifth := 5*avg - sumKnown
		if fifth <= 0 {
			return draft{}, errReject
		}
		return draft{
			body: fmt.Sprintf("The average of five numbers is %d. Four of them are %s. What is the fifth?",
				avg, listAnd(known)),
			answer:      fmtInt(fifth),
			distractors: numDistractors(r, fifth, fmtInt, avg, 4*avg-sumKnown+avg, sumKnown/4),
			explanation: fmt.Sprintf("Total = 5 × %d = %d. Known sum = %d. Fifth = %d − %d = %d.",
				avg, 5*avg, sumKnown, 5*avg, sumKnown, fifth),
		}, nil
	}
}

func buildSpeed(r *rand.Rand, d Difficulty) (draft, error) {
	switch d {
	case Easy: // distance = speed × time
		s, t := 5*between(r, 4, 16), between(r, 2, 6)
		ans := s * t
		return draft{
			body:        fmt.Sprintf("A car travels at %d km/h for %d hours. How far does it go (in km)?", s, t),
			answer:      fmtInt(ans),
			distractors: numDistractors(r, ans, fmtInt, s+t, s*(t+1), s*(t-1)),
			explanation: fmt.Sprintf("Distance = speed × time = %d × %d = %d km.", s, t, ans),
		}, nil
	case Medium: // speed = distance ÷ time
		s, t := 5*between(r, 6, 24), between(r, 2, 8)
		dist := s * t
		return draft{
			body:        fmt.Sprintf("A train covers %s km in %d hours. What is its speed (in km/h)?", fmtInt(dist), t),
			answer:      fmtInt(s),
			distractors: numDistractors(r, s, fmtInt, dist-t, s+5, s-5, s+10),
			explanation: fmt.Sprintf("Speed = distance ÷ time = %s ÷ %d = %d km/h.", fmtInt(dist), t, s),
		}, nil
	default: // train passing a pole, convert m/s → km/h
		kmh := 18 * between(r, 2, 6) // multiples of 18 convert to whole m/s
		mps := kmh * 5 / 18
		t := between(r, 6, 20)
		length := mps * t
		return draft{
			body: fmt.Sprintf("A train %d m long passes a pole in %d seconds. What is its speed (in km/h)?",
				length, t),
			answer:      fmtInt(kmh),
			distractors: numDistractors(r, kmh, fmtInt, mps, mps*3, kmh+18, kmh-18),
			explanation: fmt.Sprintf("Speed = %d ÷ %d = %d m/s, and %d × 18/5 = %d km/h.", length, t, mps, mps, kmh),
		}, nil
	}
}

func buildInterest(r *rand.Rand, d Difficulty) (draft, error) {
	p := 100 * between(r, 5, 50)
	rate := between(r, 2, 15)
	t := between(r, 1, 5)
	si := p * rate * t / 100
	if d == Medium {
		return draft{
			body: fmt.Sprintf("What is the simple interest on ₹%s at %d%% per year for %d years?",
				fmtInt(p), rate, t),
			answer:      "₹" + fmtInt(si),
			distractors: numDistractors(r, si, func(v int) string { return "₹" + fmtInt(v) }, p*rate/100, p+si, si*t),
			explanation: fmt.Sprintf("SI = P × R × T ÷ 100 = %s × %d × %d ÷ 100 = ₹%s.", fmtInt(p), rate, t, fmtInt(si)),
		}, nil
	}
	// Hard: find the rate.
	pct := func(v int) string { return fmt.Sprintf("%d%%", v) }
	return draft{
		body: fmt.Sprintf("₹%s earns ₹%s simple interest in %d years. What is the rate per year?",
			fmtInt(p), fmtInt(si), t),
		answer:      pct(rate),
		distractors: numDistractors(r, rate, pct, rate*t, rate+1, rate-1, rate+2),
		explanation: fmt.Sprintf("R = SI × 100 ÷ (P × T) = %s × 100 ÷ (%s × %d) = %d%%.", fmtInt(si), fmtInt(p), t, rate),
	}, nil
}
