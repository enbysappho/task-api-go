// Package store holds tasks in memory and is safe for concurrent use.
package store

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

// Status values a task can have.
const (
	StatusTodo  = "todo"
	StatusDoing = "doing"
	StatusDone  = "done"
)

var (
	ErrNotFound      = errors.New("task not found")
	ErrTitleRequired = errors.New("title is required")
	ErrTitleTooLong  = errors.New("title must be at most 100 characters")
	ErrBadStatus     = errors.New("status must be one of: todo, doing, done")
)

// Task is a single to-do item.
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
}

// Store is an in-memory task repository.
type Store struct {
	mu     sync.RWMutex
	nextID int
	tasks  map[int]Task
}

func New() *Store { return &Store{nextID: 1, tasks: map[int]Task{}} }

func validate(title, status string) error {
	title = strings.TrimSpace(title)
	if title == "" {
		return ErrTitleRequired
	}
	if len([]rune(title)) > 100 {
		return ErrTitleTooLong
	}
	if status != StatusTodo && status != StatusDoing && status != StatusDone {
		return ErrBadStatus
	}
	return nil
}

// Create validates and stores a new task. An empty status defaults to "todo".
func (s *Store) Create(title, status string) (Task, error) {
	if status == "" {
		status = StatusTodo
	}
	if err := validate(title, status); err != nil {
		return Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t := Task{ID: s.nextID, Title: strings.TrimSpace(title), Status: status, CreatedAt: time.Now().UTC()}
	s.tasks[t.ID] = t
	s.nextID++
	return t, nil
}

func (s *Store) Get(id int) (Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	return t, nil
}

// List returns tasks ordered by ID; a non-empty status filters the result.
func (s *Store) List(status string) []Task {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if status == "" || t.Status == status {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) Update(id int, title, status string) (Task, error) {
	if err := validate(title, status); err != nil {
		return Task{}, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	if !ok {
		return Task{}, ErrNotFound
	}
	t.Title, t.Status = strings.TrimSpace(title), status
	s.tasks[id] = t
	return t, nil
}

func (s *Store) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[id]; !ok {
		return ErrNotFound
	}
	delete(s.tasks, id)
	return nil
}
