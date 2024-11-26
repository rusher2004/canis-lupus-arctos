package store

import (
	"errors"

	"github.com/google/uuid"
)

type MemoryStore struct {
	riskMap map[string]Risk
}

type Risk struct {
	ID          uuid.UUID
	State       string
	Title       string
	Description string
}

var (
	ErrNotFound      = errors.New("risk not found")
	ErrAlreadyExists = errors.New("risk already exists")
)

// NewMemoryStore creates a new MemoryStore with a riskmap initialized with no records.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		riskMap: make(map[string]Risk),
	}
}

// CreateRisk stores a new risk in the memory store, creating a random UUID.
func (ms *MemoryStore) CreateRisk(state, title, desc string) (Risk, error) {
	r := Risk{
		ID:          uuid.New(),
		State:       state,
		Title:       title,
		Description: desc,
	}

	idStr := r.ID.String()

	if _, ok := ms.riskMap[idStr]; ok {
		return Risk{}, ErrAlreadyExists
	}

	ms.riskMap[idStr] = r
	return r, nil
}

// GetRisk retrieves a risk from the memory store by its ID.
func (ms *MemoryStore) GetRisk(id string) (Risk, error) {
	r, ok := ms.riskMap[id]
	if !ok {
		return Risk{}, ErrNotFound
	}

	return r, nil
}

// GetRiskList retrieves all risks from the memory store.
func (ms *MemoryStore) GetRiskList() ([]Risk, error) {
	var risks []Risk
	for _, r := range ms.riskMap {
		risks = append(risks, r)
	}

	return risks, nil
}
