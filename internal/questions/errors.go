package questions

import "errors"

// ErrNotEnoughQuestions means a category doesn't have enough published
// questions to fill a match or challenge.
var ErrNotEnoughQuestions = errors.New("not enough questions in this category yet — pick another category")
