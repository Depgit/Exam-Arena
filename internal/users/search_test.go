package users

import "testing"

func TestLikeEscaperTreatsWildcardsLiterally(t *testing.T) {
	for in, want := range map[string]string{
		"chunnu": "chunnu",
		"50%":    `50\%`,
		"a_b":    `a\_b`,
		`back\`:  `back\\`,
	} {
		if got := likeEscaper.Replace(in); got != want {
			t.Errorf("%q: got %q, want %q", in, got, want)
		}
	}
}
