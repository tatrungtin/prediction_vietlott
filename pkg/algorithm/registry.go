package algorithm

import (
	"fmt"
	"sync"

	"github.com/tool_predict/internal/domain/valueobject"
)

// Registry manages algorithm registration and weights
type Registry struct {
	mu         sync.RWMutex
	algorithms map[string]Algorithm
	weights    map[string]float64 // For ensemble voting
}

// NewRegistry creates a new algorithm registry
func NewRegistry() *Registry {
	return &Registry{
		algorithms: make(map[string]Algorithm),
		weights:    make(map[string]float64),
	}
}

// Register registers an algorithm with a given weight
func (r *Registry) Register(algo Algorithm, weight float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := algo.Name()
	if _, exists := r.algorithms[name]; exists {
		return fmt.Errorf("algorithm %s already registered", name)
	}

	if weight < 0 {
		return fmt.Errorf("weight cannot be negative, got %f", weight)
	}

	r.algorithms[name] = algo
	r.weights[name] = weight

	return nil
}

// RegisterOrUpdate registers an algorithm or updates its weight if already registered
func (r *Registry) RegisterOrUpdate(algo Algorithm, weight float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := algo.Name()

	if weight < 0 {
		return fmt.Errorf("weight cannot be negative, got %f", weight)
	}

	if _, exists := r.algorithms[name]; exists {
		// Update existing algorithm
		r.algorithms[name] = algo
		r.weights[name] = weight
	} else {
		// Register new algorithm
		r.algorithms[name] = algo
		r.weights[name] = weight
	}

	return nil
}

// Get retrieves an algorithm by name
func (r *Registry) Get(name string) (Algorithm, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	algo, exists := r.algorithms[name]
	if !exists {
		return nil, fmt.Errorf("algorithm %s not found", name)
	}

	return algo, nil
}

// GetAll returns all registered algorithms
func (r *Registry) GetAll() []Algorithm {
	r.mu.RLock()
	defer r.mu.RUnlock()

	algos := make([]Algorithm, 0, len(r.algorithms))
	for _, algo := range r.algorithms {
		algos = append(algos, algo)
	}

	return algos
}

// GetNames returns all registered algorithm names
func (r *Registry) GetNames() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.algorithms))
	for name := range r.algorithms {
		names = append(names, name)
	}

	return names
}

// GetWeight returns the weight for an algorithm
func (r *Registry) GetWeight(name string) float64 {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.weights[name]
}

// UpdateWeight updates the weight for an algorithm
func (r *Registry) UpdateWeight(name string, weight float64) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if weight < 0 {
		return fmt.Errorf("weight cannot be negative, got %f", weight)
	}

	if _, exists := r.algorithms[name]; !exists {
		return fmt.Errorf("algorithm %s not found", name)
	}

	r.weights[name] = weight
	return nil
}

// GetAlgorithmsForGameType returns algorithms that can predict for a specific game type
func (r *Registry) GetAlgorithmsForGameType(gameType valueobject.GameType) []Algorithm {
	r.mu.RLock()
	defer r.mu.RUnlock()

	algos := make([]Algorithm, 0)
	for _, algo := range r.algorithms {
		// All algorithms support both game types in our implementation
		// But we keep this method for future extensibility
		algos = append(algos, algo)
	}

	return algos
}

// Count returns the number of registered algorithms
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.algorithms)
}

// Unregister removes an algorithm from the registry
func (r *Registry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.algorithms[name]; !exists {
		return fmt.Errorf("algorithm %s not found", name)
	}

	delete(r.algorithms, name)
	delete(r.weights, name)

	return nil
}

// Clear removes all algorithms from the registry
func (r *Registry) Clear() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.algorithms = make(map[string]Algorithm)
	r.weights = make(map[string]float64)
}
