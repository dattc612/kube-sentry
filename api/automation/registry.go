package automation

import (
	"context"
	"fmt"
	"sync"

	"github.com/dattc/go-though-storm/api/models"
)

type StepExecFunc func(ctx context.Context, input map[string]interface{}, extraConfigs []models.AutomationJobExtraConfig) (map[string]interface{}, string, error)

type StepDefinition struct {
	Name string
	Exec StepExecFunc
}

type JobProvider func() []StepDefinition

type Registry struct {
	mu   sync.RWMutex
	jobs map[string]JobProvider
}

var DefaultRegistry = NewRegistry()

func NewRegistry() *Registry {
	return &Registry{
		jobs: make(map[string]JobProvider),
	}
}

func (r *Registry) Register(automationCode string, provider JobProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.jobs[automationCode] = provider
}

func (r *Registry) GetSteps(automationCode string) ([]StepDefinition, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.jobs[automationCode]
	if !exists {
		return nil, fmt.Errorf("No automation workflow registered for code: %s", automationCode)
	}
	return provider(), nil
}
