package queue

import (
	"errors"
	"net/url"
	"sync"
)

var (
	ErrExternalDomain = errors.New("external domain")
	ErrDepthLimit     = errors.New("depth limit")
	ErrURLisVisited   = errors.New("URL is visited")
	ErrQueueFull      = errors.New("queue is full")
)

type Task struct {
	URL   *url.URL
	Depth int
	Type  string
}

type Queue struct {
	tasks   chan Task
	visited map[string]bool
	mu      sync.Mutex
	domain  string
}

func NewQueue(capacity int, domain string) *Queue {
	return &Queue{
		tasks:   make(chan Task, capacity),
		visited: make(map[string]bool),
		domain:  domain,
	}
}

func (q *Queue) Enqueue(t Task, maxDepth int) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if t.URL.Host != q.domain {
		return ErrExternalDomain
	}

	if t.Depth > maxDepth {
		return ErrDepthLimit
	}

	urlStr := t.URL.String()
	if _, exists := q.visited[urlStr]; exists {
		return ErrURLisVisited
	} else {
		q.visited[urlStr] = true
	}

	select {
	case q.tasks <- t:
		return nil
	default:
		return ErrQueueFull
	}
}

func (q *Queue) Dequeue() <-chan Task {
	return q.tasks
}

func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	close(q.tasks)
}
