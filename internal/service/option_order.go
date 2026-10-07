package service

import (
	"hash/fnv"
	randv2 "math/rand/v2"
	"sort"

	"github.com/exam-arena/internal/models"
)

// Answer options are shuffled every time a question is handed to a player,
// so a question seen before doesn't keep its answer in the same position.
//
// The shuffle is seeded from a key (match+player, practice session, daily
// date+player), never from the clock: reloading the page mid-match
// reproduces the same order instead of making options jump around, while
// every new session gets a fresh one. Grading uses option IDs, so the order
// shown never affects scoring.

// playerOptions returns q's options in the order this key should see them,
// numbered 1..n in that order (clients sort by order_index).
func playerOptions(q models.Question, key string) []models.OptionForPlayer {
	opts := make([]models.OptionForPlayer, len(q.Options))
	for i, o := range q.Options {
		opts[i] = models.OptionForPlayer{ID: o.ID, OptionText: o.OptionText, OrderIndex: o.OrderIndex}
	}
	// Start from the stored order so the result doesn't depend on the
	// order the database happened to return rows in.
	sort.SliceStable(opts, func(i, j int) bool { return opts[i].OrderIndex < opts[j].OrderIndex })

	if key != "" {
		h := fnv.New64a()
		h.Write([]byte(key))
		h.Write([]byte{0})
		h.Write([]byte(q.ID))
		seed := h.Sum64()
		r := randv2.New(randv2.NewPCG(seed, seed^0x9E3779B97F4A7C15))
		r.Shuffle(len(opts), func(i, j int) { opts[i], opts[j] = opts[j], opts[i] })
	}
	for i := range opts {
		opts[i].OrderIndex = i + 1
	}
	return opts
}

// matchShuffleKey gives each player in a match their own option order.
func matchShuffleKey(matchID, userID string) string { return "match:" + matchID + ":" + userID }
