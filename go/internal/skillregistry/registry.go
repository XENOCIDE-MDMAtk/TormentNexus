package skillregistry

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// SkillInfo describes a registered skill.
type SkillInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Content     string   `json:"content,omitempty"`
	Category    string   `json:"category,omitempty"`
	AlwaysOn    bool     `json:"alwaysOn"`
	Tags        []string `json:"tags,omitempty"`
	Path        string   `json:"path,omitempty"`
}

// SkillRegistry manages the global skill inventory.
type SkillRegistry struct {
	mu     sync.RWMutex
	skills map[string]*SkillInfo
}

// NewSkillRegistry creates a new empty registry.
func NewSkillRegistry() *SkillRegistry {
	return &SkillRegistry{
		skills: make(map[string]*SkillInfo),
	}
}

// Register adds or updates a skill in the registry.
func (sr *SkillRegistry) Register(skill SkillInfo) error {
	if skill.ID == "" {
		return fmt.Errorf("skill ID cannot be empty")
	}
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.skills[strings.ToLower(skill.ID)] = &skill
	return nil
}

// Get returns a skill by ID.
func (sr *SkillRegistry) Get(id string) (*SkillInfo, bool) {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	s, ok := sr.skills[strings.ToLower(id)]
	if !ok {
		return nil, false
	}
	copy := *s
	return &copy, true
}

// List returns all registered skills.
func (sr *SkillRegistry) List() []SkillInfo {
	sr.mu.RLock()
	defer sr.mu.RUnlock()

	result := make([]SkillInfo, 0, len(sr.skills))
	for _, s := range sr.skills {
		result = append(result, *s)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

// Search performs a fuzzy search across skill names and descriptions.
func (sr *SkillRegistry) Search(query string, limit int) []SkillInfo {
	if limit <= 0 {
		limit = 10
	}
	query = strings.ToLower(query)

	sr.mu.RLock()
	defer sr.mu.RUnlock()

	type scored struct {
		skill SkillInfo
		score float64
	}

	var results []scored
	for _, s := range sr.skills {
		score := 0.0

		if strings.ToLower(s.ID) == query || strings.ToLower(s.Name) == query {
			score += 10.0
		} else if strings.Contains(strings.ToLower(s.Name), query) {
			score += 5.0
		}

		if strings.Contains(strings.ToLower(s.Description), query) {
			score += 3.0
		}

		for _, tag := range s.Tags {
			if strings.ToLower(tag) == query {
				score += 4.0
			} else if strings.Contains(strings.ToLower(tag), query) {
				score += 1.0
			}
		}

		if score > 0 {
			results = append(results, scored{skill: *s, score: score})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	if len(results) > limit {
		results = results[:limit]
	}

	skills := make([]SkillInfo, len(results))
	for i, r := range results {
		skills[i] = r.skill
	}
	return skills
}
