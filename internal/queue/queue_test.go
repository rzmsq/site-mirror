package queue

import (
	"errors"
	"net/url"
	"sync"
	"testing"
	"time"
)

func TestNewQueue(t *testing.T) {
	t.Parallel()

	capacity := 100
	domain := "example.com"

	q := NewQueue(capacity, domain)

	if q == nil {
		t.Fatal("NewQueue returned nil")
	}
	if q.domain != domain {
		t.Errorf("domain: got %q, want %q", q.domain, domain)
	}
	if cap(q.tasks) != capacity {
		t.Errorf("tasks channel capacity: got %d, want %d", cap(q.tasks), capacity)
	}
	if q.visited == nil {
		t.Error("visited map is nil")
	}
}

func TestEnqueue_ValidTask(t *testing.T) {
	t.Parallel()

	q := NewQueue(10, "example.com")
	u, _ := url.Parse("https://example.com/page1")
	task := Task{URL: u, Depth: 1, Type: "page"}

	err := q.Enqueue(task, 5)

	if err != nil {
		t.Errorf("Enqueue returned error for valid task: %v", err)
	}

	select {
	case received := <-q.tasks:
		if received.URL.String() != task.URL.String() {
			t.Errorf("received task URL: got %q, want %q", received.URL.String(), task.URL.String())
		}
		q.Done()
	case <-time.After(100 * time.Millisecond):
		t.Error("task was not added to channel")
	}
}

func TestEnqueue_ExternalDomain(t *testing.T) {
	t.Parallel()

	q := NewQueue(10, "example.com")
	u, _ := url.Parse("https://other.com/page1")
	task := Task{URL: u, Depth: 1, Type: "page"}

	err := q.Enqueue(task, 5)

	if !errors.Is(err, ErrExternalDomain) {
		t.Errorf("expected ErrExternalDomain, got %v", err)
	}

	select {
	case <-q.tasks:
		t.Error("task was added to channel despite wrong domain")
	default:
	}
}

func TestEnqueue_DepthLimit(t *testing.T) {
	t.Parallel()

	q := NewQueue(10, "example.com")
	u, _ := url.Parse("https://example.com/page1")
	task := Task{URL: u, Depth: 6, Type: "page"}

	err := q.Enqueue(task, 5)

	if !errors.Is(err, ErrDepthLimit) {
		t.Errorf("expected ErrDepthLimit, got %v", err)
	}
}

func TestEnqueue_Duplicate(t *testing.T) {
	t.Parallel()

	q := NewQueue(10, "example.com")
	u, _ := url.Parse("https://example.com/page1")
	task := Task{URL: u, Depth: 1, Type: "page"}

	err1 := q.Enqueue(task, 5)
	err2 := q.Enqueue(task, 5)

	if err1 != nil {
		t.Errorf("first Enqueue returned error: %v", err1)
	}
	if !errors.Is(err2, ErrURLisVisited) {
		t.Errorf("expected ErrURLisVisited, got %v", err2)
	}
}

func TestEnqueue_FullQueue(t *testing.T) {
	t.Parallel()

	q := NewQueue(2, "example.com")

	for i := 1; i <= 2; i++ {
		u, _ := url.Parse("https://example.com/page" + string(rune('0'+i)))
		task := Task{URL: u, Depth: 1, Type: "page"}
		if err := q.Enqueue(task, 5); err != nil {
			t.Errorf("Enqueue %d failed: %v", i, err)
		}
	}

	u, _ := url.Parse("https://example.com/page3")
	task := Task{URL: u, Depth: 1, Type: "page"}

	err := q.Enqueue(task, 5)

	if !errors.Is(err, ErrQueueFull) {
		t.Errorf("expected ErrQueueFull, got %v", err)
	}
}

func TestDequeue(t *testing.T) {
	t.Parallel()

	q := NewQueue(10, "example.com")
	u, _ := url.Parse("https://example.com/page1")
	task := Task{URL: u, Depth: 1, Type: "page"}

	if err := q.Enqueue(task, 5); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	ch := q.Dequeue()
	select {
	case received := <-ch:
		if received.URL.String() != task.URL.String() {
			t.Errorf("dequeued task: got %q, want %q", received.URL.String(), task.URL.String())
		}
		q.Done()
	case <-time.After(100 * time.Millisecond):
		t.Error("Dequeue timed out")
	}
}

func TestDone(t *testing.T) {
	t.Parallel()

	q := NewQueue(10, "example.com")
	u, _ := url.Parse("https://example.com/page1")
	task := Task{URL: u, Depth: 1, Type: "page"}

	if err := q.Enqueue(task, 5); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	<-q.tasks
	q.Done()

	done := make(chan bool)
	go func() {
		q.activeTasks.Wait()
		done <- true
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Error("Done did not decrement activeTasks counter")
	}
}

func TestClose(t *testing.T) {
	q := NewQueue(10, "example.com")

	q.Close()

	select {
	case _, ok := <-q.tasks:
		if ok {
			t.Error("channel should be closed")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("channel was not closed")
	}
}

func TestWaitAndClose(t *testing.T) {
	q := NewQueue(10, "example.com")
	u, _ := url.Parse("https://example.com/page1")
	task := Task{URL: u, Depth: 1, Type: "page"}

	if err := q.Enqueue(task, 5); err != nil {
		t.Fatalf("Enqueue failed: %v", err)
	}

	go func() {
		<-q.tasks
		time.Sleep(50 * time.Millisecond)
		q.Done()
	}()

	done := make(chan bool)
	go func() {
		q.WaitAndClose()
		done <- true
	}()

	select {
	case <-done:
		_, ok := <-q.tasks
		if ok {
			t.Error("channel should be closed after WaitAndClose")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("WaitAndClose did not complete in time")
	}
}

func TestConcurrentEnqueue(t *testing.T) {
	t.Parallel()

	q := NewQueue(100, "example.com")
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			u, _ := url.Parse("https://example.com/page" + string(rune('0'+i)))
			task := Task{URL: u, Depth: 1, Type: "page"}
			_ = q.Enqueue(task, 5)
		}(i)
	}

	wg.Wait()

	count := 0
	timeout := time.After(1 * time.Second)

	for count < numGoroutines {
		select {
		case <-q.tasks:
			count++
			q.Done()
		case <-timeout:
			t.Fatalf("received %d tasks, expected %d", count, numGoroutines)
		}
	}
}

func TestVisited_ThreadSafety(t *testing.T) {
	t.Parallel()

	q := NewQueue(100, "example.com")
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			u, _ := url.Parse("https://example.com/page1")
			task := Task{URL: u, Depth: 1, Type: "page"}
			_ = q.Enqueue(task, 5)
		}()
	}

	wg.Wait()

	count := 0
	timeout := time.After(100 * time.Millisecond)

	for {
		select {
		case <-q.tasks:
			count++
			q.Done()
		case <-timeout:
			if count != 1 {
				t.Errorf("expected exactly 1 task in queue, got %d", count)
			}
			return
		}
	}
}
