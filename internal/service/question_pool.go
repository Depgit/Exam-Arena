package service

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand/v2"
	"sync"
	"time"

	appCache "github.com/exam-arena/internal/cache"
	"github.com/exam-arena/internal/questiongen"
	"github.com/exam-arena/internal/repository"
)

// QuestionPool keeps a rotating stock of generated questions in the
// questions table, next to the hand-written ones.
//
// Matches, practice and the daily challenge read questions exactly as
// before; this only decides how many generated rows exist. Each category
// holds PoolSize published generated questions, split evenly across the
// three difficulties. Every RotateEvery, RotateFraction of them are
// archived (never deleted while anything refers to them) and replaced
// with fresh ones, so players keep seeing new questions.
type QuestionPool struct {
	repo *repository.QuestionRepo
	bank *appCache.QuestionBank
	cfg  QuestionPoolConfig
	mu   sync.Mutex // one fill/rotate at a time per instance
}

type QuestionPoolConfig struct {
	Enabled        bool
	PoolSize       int           // per category, across all difficulties
	RotateEvery    time.Duration // how often to refresh part of the pool
	RotateFraction float64       // share of each pool replaced per rotation
	// PerUser caps all published questions per category at
	// players × PerUser ÷ categories. Beyond that the oldest are archived,
	// hand-written or generated alike. 0 disables the cap.
	PerUser int
}

func NewQuestionPool(repo *repository.QuestionRepo, bank *appCache.QuestionBank, cfg QuestionPoolConfig) *QuestionPool {
	if cfg.PoolSize < len(questiongen.Difficulties) {
		cfg.PoolSize = 600
	}
	if cfg.RotateEvery <= 0 {
		cfg.RotateEvery = 6 * time.Hour
	}
	if cfg.RotateFraction <= 0 || cfg.RotateFraction > 1 {
		cfg.RotateFraction = 0.25
	}
	return &QuestionPool{repo: repo, bank: bank, cfg: cfg}
}

func (p *QuestionPool) perLevel() int { return p.cfg.PoolSize / len(questiongen.Difficulties) }

// budget is the current size plan: how many generated questions to keep
// per difficulty, and how many published questions each category may hold
// in total (0 = no cap).
func (p *QuestionPool) budget(ctx context.Context, activeCategories int) (levelTarget, categoryCap int, err error) {
	if p.cfg.PerUser <= 0 || activeCategories == 0 {
		return p.perLevel(), 0, nil
	}
	players, err := p.repo.CountPlayers(ctx)
	if err != nil {
		return 0, 0, err
	}
	levelTarget, categoryCap = capBudget(p.perLevel(), players, p.cfg.PerUser, activeCategories, len(questiongen.Difficulties))
	slog.Info("question budget", "players", players, "per_user", p.cfg.PerUser,
		"total_cap", players*p.cfg.PerUser, "per_category", categoryCap, "generated_per_level", levelTarget)
	return levelTarget, categoryCap, nil
}

// capBudget splits players × perUser evenly across categories, and keeps
// the generated pool within each category's share so the generator never
// inserts questions the cap would immediately archive.
func capBudget(perLevel, players, perUser, categories, levels int) (levelTarget, categoryCap int) {
	if perUser <= 0 || categories <= 0 || levels <= 0 {
		return perLevel, 0
	}
	categoryCap = players * perUser / categories
	levelTarget = perLevel
	if capped := categoryCap / levels; capped < levelTarget {
		levelTarget = capped
	}
	return levelTarget, categoryCap
}

// activeGeneratedCategories maps each generator category code to its id,
// skipping missing or deactivated ones.
func (p *QuestionPool) activeGeneratedCategories(ctx context.Context) (map[string]string, error) {
	out := map[string]string{}
	for _, code := range questiongen.Categories() {
		id, err := p.repo.ActiveCategoryIDByCode(ctx, code)
		if err != nil {
			return nil, err
		}
		if id != "" {
			out[code] = id
		}
	}
	return out, nil
}

// Start fills the pool in the background, then rotates it on a timer.
// Startup is not delayed: matches use the existing bank until the first
// fill lands and refreshes it.
func (p *QuestionPool) Start(ctx context.Context) {
	if !p.cfg.Enabled {
		slog.Info("question generator disabled (QUESTION_GEN_ENABLED=false)")
		return
	}
	go func() {
		if err := p.Fill(ctx); err != nil {
			slog.Error("question pool initial fill failed", "error", err)
		}
		ticker := time.NewTicker(p.cfg.RotateEvery)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := p.Rotate(ctx); err != nil {
					slog.Error("question pool rotation failed", "error", err)
				}
			}
		}
	}()
}

// FillResult reports what one fill or rotation did, per category code.
type FillResult struct {
	Inserted map[string]int `json:"inserted"`
	Retired  map[string]int `json:"retired,omitempty"`
	Pruned   int64          `json:"pruned"`
}

// Fill tops every generated category up to its target.
func (p *QuestionPool) Fill(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	_, err := p.fillLocked(ctx)
	return err
}

func (p *QuestionPool) fillLocked(ctx context.Context) (map[string]int, error) {
	inserted := map[string]int{}
	cats, err := p.activeGeneratedCategories(ctx)
	if err != nil {
		return inserted, err
	}
	target, categoryCap, err := p.budget(ctx, len(cats))
	if err != nil {
		return inserted, err
	}
	for _, code := range questiongen.Categories() {
		catID, ok := cats[code]
		if !ok {
			continue // category missing or deactivated
		}
		n, trimmed, err := p.fillCategory(ctx, code, catID, target)
		inserted[code] = n
		if err != nil {
			return inserted, fmt.Errorf("%s: %w", code, err)
		}
		// Enforce the cap after topping up: keep the newest questions,
		// archive the oldest of any kind.
		if categoryCap > 0 {
			archived, err := p.repo.ArchiveOldestBeyond(ctx, catID, categoryCap)
			if err != nil {
				return inserted, fmt.Errorf("%s: %w", code, err)
			}
			if archived > 0 {
				slog.Info("question cap: archived oldest questions", "category", code, "archived", archived, "cap", categoryCap)
				trimmed += int(archived)
			}
		}
		if n > 0 || trimmed > 0 {
			if err := p.bank.RefreshCategory(ctx, catID); err != nil {
				slog.Warn("question bank refresh failed", "category", code, "error", err)
			}
		}
	}
	slog.Info("question pool filled", "inserted", inserted)
	return inserted, nil
}

// fillCategory brings each difficulty's generated pool to target: tops it
// up, or retires the oldest extras if the cap has shrunk below what's there.
func (p *QuestionPool) fillCategory(ctx context.Context, code, catID string, target int) (inserted, trimmed int, err error) {
	counts, bodies, err := p.repo.GeneratedPoolState(ctx, catID)
	if err != nil {
		return 0, 0, err
	}
	for _, d := range questiongen.Difficulties {
		if extra := counts[string(d)] - target; extra > 0 {
			n, err := p.repo.RetireOldestGenerated(ctx, catID, string(d), extra)
			if err != nil {
				return 0, trimmed, err
			}
			trimmed += int(n)
			counts[string(d)] -= int(n)
		}
	}
	if trimmed > 0 {
		slog.Info("question pool trimmed to budget", "category", code, "retired", trimmed)
	}
	topicIDs := map[string]string{}
	for _, name := range questiongen.Topics(code) {
		id, err := p.repo.EnsureTopic(ctx, catID, name)
		if err != nil {
			return 0, trimmed, err
		}
		topicIDs[name] = id
	}

	total := 0
	for _, d := range questiongen.Difficulties {
		need := target - counts[string(d)]
		if need <= 0 {
			continue
		}
		batch := make([]repository.GeneratedQuestion, 0, need)
		// Fresh random seeds; each stored question keeps its own seed so it
		// can be rebuilt. Kept within int63 to fit the BIGINT column.
		for attempts := 0; len(batch) < need && attempts < need*20; attempts++ {
			seed := rand.Uint64() & math.MaxInt64
			q, err := questiongen.Generate(code, d, seed)
			if err != nil {
				continue
			}
			if bodies[q.Body] {
				continue // already in the pool
			}
			bodies[q.Body] = true
			batch = append(batch, repository.GeneratedQuestion{
				CategoryID:       catID,
				TopicID:          topicIDs[q.Topic],
				Difficulty:       string(d),
				Body:             q.Body,
				Explanation:      q.Explanation,
				EstimatedSeconds: q.EstimatedSeconds,
				GeneratorKey:     q.TemplateKey,
				GeneratorSeed:    int64(q.Seed),
				Options:          q.Options,
				Correct:          q.Correct,
			})
		}
		n, err := p.repo.InsertGenerated(ctx, catID, string(d), target, batch)
		total += n
		if err != nil {
			return total, trimmed, err
		}
	}
	return total, trimmed, nil
}

// Rotate archives the oldest RotateFraction of each generated pool, tops
// the pools back up with fresh questions, and prunes archived generated
// questions that nothing refers to any more.
func (p *QuestionPool) Rotate(ctx context.Context) error {
	_, err := p.RotateNow(ctx)
	return err
}

// RotateNow is Rotate with a report, for the admin endpoint.
func (p *QuestionPool) RotateNow(ctx context.Context) (*FillResult, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	res := &FillResult{Retired: map[string]int{}}
	cats, err := p.activeGeneratedCategories(ctx)
	if err != nil {
		return res, err
	}
	target, _, err := p.budget(ctx, len(cats))
	if err != nil {
		return res, err
	}
	retire := int(math.Round(float64(target) * p.cfg.RotateFraction))
	for _, code := range questiongen.Categories() {
		catID, ok := cats[code]
		if !ok {
			continue
		}
		for _, d := range questiongen.Difficulties {
			n, err := p.repo.RetireOldestGenerated(ctx, catID, string(d), retire)
			if err != nil {
				return res, err
			}
			res.Retired[code] += int(n)
		}
	}

	inserted, err := p.fillLocked(ctx)
	res.Inserted = inserted
	if err != nil {
		return res, err
	}

	pruned, err := p.repo.PruneArchivedGenerated(ctx)
	if err != nil {
		slog.Warn("pruning archived generated questions failed", "error", err)
	}
	res.Pruned = pruned
	slog.Info("question pool rotated", "retired", res.Retired, "inserted", res.Inserted, "pruned", pruned)
	return res, nil
}

// PreviewQuestion is a generated question with its answer visible, for
// admins checking quality. Nothing is stored.
type PreviewQuestion struct {
	Template    string   `json:"template"`
	Topic       string   `json:"topic"`
	Difficulty  string   `json:"difficulty"`
	Seed        uint64   `json:"seed"`
	Body        string   `json:"body"`
	Options     []string `json:"options"`
	Correct     int      `json:"correct"`
	Explanation string   `json:"explanation"`
}

// Preview generates n questions for a category without storing them.
// difficulty "" mixes all three levels.
func (p *QuestionPool) Preview(code, difficulty string, n int) ([]PreviewQuestion, error) {
	if questiongen.Templates(code) == nil {
		return nil, fmt.Errorf("no generator for category %q (have %v)", code, questiongen.Categories())
	}
	if n <= 0 || n > 100 {
		n = 20
	}
	out := make([]PreviewQuestion, 0, n)
	for i := 0; i < n; i++ {
		d := questiongen.Difficulty(difficulty)
		if difficulty == "" {
			d = questiongen.Difficulties[i%len(questiongen.Difficulties)]
		}
		q, err := questiongen.Generate(code, d, rand.Uint64()&math.MaxInt64)
		if err != nil {
			return nil, err
		}
		out = append(out, PreviewQuestion{
			Template: q.TemplateKey, Topic: q.Topic, Difficulty: string(q.Difficulty), Seed: q.Seed,
			Body: q.Body, Options: q.Options, Correct: q.Correct, Explanation: q.Explanation,
		})
	}
	return out, nil
}
