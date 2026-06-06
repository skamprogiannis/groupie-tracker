package groupie

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

// blockingCatalog makes every search wait until the test releases it, which
// lets the timeout and cancellation tests run deterministically.
type blockingCatalog struct {
	released chan struct{}
}

func (c *blockingCatalog) Search(string) []SearchResult {
	<-c.released
	return nil
}

// newBlockingWorker returns a worker whose searches block until cleanup. The
// cleanup releases any in-flight search before closing the worker so the
// goroutine can exit; without that ordering Close would deadlock.
func newBlockingWorker(t *testing.T) *SearchWorker {
	t.Helper()
	catalog := &blockingCatalog{released: make(chan struct{})}
	worker := NewSearchWorker(catalog)
	t.Cleanup(func() {
		close(catalog.released)
		worker.Close()
	})
	return worker
}

func TestSearchWorkerReturnsResults(t *testing.T) {
	catalog := newTestCatalog(t)
	worker := NewSearchWorker(catalog)
	defer worker.Close()

	results, err := worker.Search(context.Background(), "freddie")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results) != 1 || results[0].Name != "Queen" {
		t.Fatalf("results = %#v, want single Queen result", results)
	}
}

func TestSearchWorkerEmptyQueryReturnsAll(t *testing.T) {
	catalog := newTestCatalog(t)
	worker := NewSearchWorker(catalog)
	defer worker.Close()

	results, err := worker.Search(context.Background(), "")
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(results) != 4 {
		t.Fatalf("len(results) = %d, want 4 for empty query", len(results))
	}
}

func TestSearchWorkerTimeout(t *testing.T) {
	// The underlying search blocks until cleanup, so the request can only
	// complete by hitting the context deadline.
	worker := newBlockingWorker(t)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	results, err := worker.Search(ctx, "anything")
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("err = %v, want context.DeadlineExceeded", err)
	}
	if results != nil {
		t.Fatalf("results = %#v, want nil on timeout", results)
	}
}

func TestSearchWorkerCancellation(t *testing.T) {
	worker := newBlockingWorker(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // already canceled before the call

	_, err := worker.Search(ctx, "anything")
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestSearchWorkerCloseIsSafeAndRepeatable(t *testing.T) {
	worker := NewSearchWorker(newTestCatalog(t))

	worker.Close()
	worker.Close() // second Close must not panic

	_, err := worker.Search(context.Background(), "queen")
	if !errors.Is(err, ErrWorkerClosed) {
		t.Fatalf("err = %v, want ErrWorkerClosed after Close", err)
	}
}

func TestSearchWorkerNilReceiver(t *testing.T) {
	var worker *SearchWorker
	if _, err := worker.Search(context.Background(), "queen"); !errors.Is(err, ErrWorkerClosed) {
		t.Fatalf("err = %v, want ErrWorkerClosed for nil worker", err)
	}
	worker.Close() // must not panic
}

func TestSearchWorkerHandlesConcurrentRequests(t *testing.T) {
	catalog := newTestCatalog(t)
	worker := NewSearchWorker(catalog)
	defer worker.Close()

	const callers = 20
	var wg sync.WaitGroup
	wg.Add(callers)
	for i := 0; i < callers; i++ {
		go func() {
			defer wg.Done()
			if _, err := worker.Search(context.Background(), "queen"); err != nil {
				t.Errorf("concurrent Search error: %v", err)
			}
		}()
	}
	wg.Wait()
}
