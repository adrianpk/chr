package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"forge.adrianpk.com/chc/internal/core"
)

const longGenerationTimeout = 5 * time.Minute

var activeGenerations = generationGuard{active: map[string]struct{}{}}

type generationGuard struct {
	mu     sync.Mutex
	active map[string]struct{}
}

func (g *generationGuard) acquire(key string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()

	if _, ok := g.active[key]; ok {
		return false
	}

	g.active[key] = struct{}{}

	return true
}

func (g *generationGuard) release(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	delete(g.active, key)
}

func (s service) withGenerationRuntime(
	parent context.Context,
	kind string,
	scope core.QuizScope,
	run func(context.Context) error,
) error {
	key := s.generationRuntimeKey(kind, scope)
	if !activeGenerations.acquire(key) {
		s.appendPlanGenerationLog("generation_guard_blocked", map[string]string{
			"kind":      kind,
			"case_slug": scope.CaseSlug,
		})

		return fmt.Errorf("%s generation already running for scope %s", kind, specScopeID(scope))
	}
	defer activeGenerations.release(key)

	ctx, cancel := context.WithTimeout(parent, longGenerationTimeout)
	defer cancel()

	return run(ctx)
}

func (s service) generationRuntimeKey(kind string, scope core.QuizScope) string {
	return s.projectContext.ProjectDir + ":" + kind + ":" + specScopeID(scope)
}
