package inspectionchecklist

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var ErrNotFound = errors.New("inspection checklist not found")

type Store interface {
	EnsureDefaultTemplate(Template) error
	GetActiveTemplate() (Template, error)
	CreateTemplate(items []TemplateItem, actorID, reason string, now time.Time) (Template, error)
	GetChecklist(tripID string) (Checklist, []Item, error)
	CreateChecklist(Checklist, []Item) error
	SaveItem(Item) error
}

type MemoryStore struct {
	mu         sync.RWMutex
	templates  map[string]Template
	activeID   string
	checklists map[string]Checklist
	items      map[string]map[string]Item
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		templates:  make(map[string]Template),
		checklists: make(map[string]Checklist),
		items:      make(map[string]map[string]Item),
	}
}

func (s *MemoryStore) EnsureDefaultTemplate(template Template) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeID != "" {
		return nil
	}
	s.templates[template.ID] = cloneTemplate(template)
	s.activeID = template.ID
	return nil
}

func (s *MemoryStore) GetActiveTemplate() (Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.activeID == "" {
		return Template{}, ErrNotFound
	}
	return cloneTemplate(s.templates[s.activeID]), nil
}

func (s *MemoryStore) CreateTemplate(items []TemplateItem, actorID, reason string, now time.Time) (Template, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var maxVersion int64
	for id, template := range s.templates {
		if template.Version > maxVersion {
			maxVersion = template.Version
		}
		if template.Active {
			template.Active = false
			s.templates[id] = template
		}
	}
	version := maxVersion + 1
	template := Template{
		ID:        templateID(version),
		Version:   version,
		Active:    true,
		Items:     cloneTemplateItems(items),
		CreatedBy: actorID,
		Reason:    reason,
		CreatedAt: now,
	}
	s.templates[template.ID] = template
	s.activeID = template.ID
	return cloneTemplate(template), nil
}

func (s *MemoryStore) GetChecklist(tripID string) (Checklist, []Item, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	checklist, ok := s.checklists[tripID]
	if !ok {
		return Checklist{}, nil, ErrNotFound
	}
	byKey := s.items[tripID]
	items := make([]Item, 0, len(byKey))
	for _, item := range byKey {
		items = append(items, item)
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].SortOrder == items[j].SortOrder {
			return items[i].Key < items[j].Key
		}
		return items[i].SortOrder < items[j].SortOrder
	})
	return checklist, items, nil
}

func (s *MemoryStore) CreateChecklist(checklist Checklist, items []Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.checklists[checklist.TripID]; exists {
		return nil
	}
	s.checklists[checklist.TripID] = checklist
	byKey := make(map[string]Item, len(items))
	for _, item := range items {
		byKey[item.Key] = item
	}
	s.items[checklist.TripID] = byKey
	return nil
}

func (s *MemoryStore) SaveItem(item Item) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	byKey, ok := s.items[item.TripID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := byKey[item.Key]; !ok {
		return ErrNotFound
	}
	byKey[item.Key] = item
	checklist := s.checklists[item.TripID]
	checklist.UpdatedAt = latestTime(checklist.UpdatedAt, item.CustomerUpdatedAt, item.DriverUpdatedAt)
	s.checklists[item.TripID] = checklist
	return nil
}

func templateID(version int64) string { return "inspection_checklist_v" + itoa(version) }

func itoa(value int64) string {
	if value == 0 {
		return "0"
	}
	buf := [20]byte{}
	index := len(buf)
	for value > 0 {
		index--
		buf[index] = byte('0' + value%10)
		value /= 10
	}
	return string(buf[index:])
}

func cloneTemplate(template Template) Template {
	template.Items = cloneTemplateItems(template.Items)
	return template
}

func cloneTemplateItems(items []TemplateItem) []TemplateItem {
	return append([]TemplateItem(nil), items...)
}

func latestTime(base time.Time, values ...*time.Time) time.Time {
	latest := base
	for _, value := range values {
		if value != nil && value.After(latest) {
			latest = *value
		}
	}
	return latest
}
