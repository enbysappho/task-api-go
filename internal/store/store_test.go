package store

import (
	"errors"
	"strings"
	"sync"
	"testing"
)

func TestCreateDefaultsAndTrims(t *testing.T) {
	s := New()
	task, err := s.Create("  write tests  ", "")
	if err != nil {
		t.Fatal(err)
	}
	if task.ID != 1 || task.Title != "write tests" || task.Status != StatusTodo {
		t.Fatalf("unexpected task: %+v", task)
	}
}

func TestValidationTable(t *testing.T) {
	cases := []struct {
		name, title, status string
		want                error
	}{
		{"empty title", "", StatusTodo, ErrTitleRequired},
		{"blank title", "   ", StatusTodo, ErrTitleRequired},
		{"title at limit", strings.Repeat("a", 100), StatusTodo, nil},
		{"title over limit", strings.Repeat("a", 101), StatusTodo, ErrTitleTooLong},
		{"bad status", "x", "waiting", ErrBadStatus},
		{"status doing", "x", StatusDoing, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := New().Create(c.title, c.status)
			if !errors.Is(err, c.want) {
				t.Fatalf("got %v, want %v", err, c.want)
			}
		})
	}
}

func TestGetUpdateDelete(t *testing.T) {
	s := New()
	a, _ := s.Create("a", "")
	if _, err := s.Get(99); !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
	u, err := s.Update(a.ID, "a2", StatusDone)
	if err != nil || u.Title != "a2" || u.Status != StatusDone {
		t.Fatalf("update failed: %+v %v", u, err)
	}
	if err := s.Delete(a.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(a.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second delete: want ErrNotFound, got %v", err)
	}
}

func TestListFilterAndOrder(t *testing.T) {
	s := New()
	s.Create("one", StatusTodo)
	s.Create("two", StatusDone)
	s.Create("three", StatusTodo)
	if got := s.List(""); len(got) != 3 || got[0].ID != 1 || got[2].ID != 3 {
		t.Fatalf("unexpected order: %+v", got)
	}
	if got := s.List(StatusTodo); len(got) != 2 {
		t.Fatalf("filter failed: %+v", got)
	}
}

func TestConcurrentCreates(t *testing.T) {
	s := New()
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); s.Create("x", "") }()
	}
	wg.Wait()
	if n := len(s.List("")); n != 50 {
		t.Fatalf("want 50 tasks, got %d", n)
	}
}
