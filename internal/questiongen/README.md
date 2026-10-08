# questiongen

**Job:** invent multiple-choice Math and Logical Reasoning questions.

**Pure code:** no database, no clock, no global randomness. The same
(template, difficulty, seed) always produces the same question, so a stored
question can be rebuilt exactly.

| You call | You get back |
|---|---|
| `Generate("MATH", questiongen.Hard, seed)` | a `Question`: text, 4 options, which is correct, explanation |
| `GenerateWith(template, difficulty, seed)` | the same, from one specific template |
| `Categories()`, `Templates(cat)`, `Topics(cat)` | what can be generated |

**Files:** `math.go` (12 math templates), `reasoning.go` (13 reasoning
templates, with checks that reject ambiguous puzzles), `options.go` (number
formatting and believable wrong answers), `questiongen.go` (the engine).

**Tests** generate thousands of questions per template and check each has 4
distinct options, exactly one correct, and an independently re-computed answer.
