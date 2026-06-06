package groupie

import (
	"context"
	"errors"
	"sync"
)

// ErrWorkerClosed is returned when a search is requested after the worker has
// been shut down.
var ErrWorkerClosed = errors.New("search worker closed")

// searchable is the minimal capability the worker needs from a catalog. The
// in-memory *Catalog satisfies it, and tests can supply slow or failing
// implementations to exercise timeout and cancellation paths.
type searchable interface {
	Search(query SearchQuery) []SearchResult
}

// searchRequest carries a single query plus the caller's context and a private
// reply channel through the worker.
type searchRequest struct {
	ctx   context.Context
	query SearchQuery
	reply chan []SearchResult
}

// SearchWorker runs catalog searches on a dedicated goroutine. Callers
// communicate with it exclusively through channels, which keeps the immutable
// catalog access serialized and gives every request independent timeout and
// cancellation handling. This is the project's asynchronous client-server
// event: the /api/search handler hands each query to the worker and waits for
// a reply or a deadline.
type SearchWorker struct {
	catalog   searchable
	requests  chan searchRequest
	quit      chan struct{}
	done      chan struct{}
	closeOnce sync.Once
}

// NewSearchWorker starts the worker goroutine and returns immediately. Callers
// must invoke Close to stop the goroutine.
func NewSearchWorker(catalog searchable) *SearchWorker {
	worker := &SearchWorker{
		catalog:  catalog,
		requests: make(chan searchRequest),
		quit:     make(chan struct{}),
		done:     make(chan struct{}),
	}
	go worker.run()
	return worker
}

func (w *SearchWorker) run() {
	defer close(w.done)
	for {
		select {
		case <-w.quit:
			return
		case req := <-w.requests:
			w.handle(req)
		}
	}
}

func (w *SearchWorker) handle(req searchRequest) {
	// Skip work the caller has already abandoned.
	if req.ctx.Err() != nil {
		return
	}

	var results []SearchResult
	if w.catalog != nil {
		results = w.catalog.Search(req.query)
	}

	// req.reply is buffered, so this never blocks the worker even if the
	// caller has already timed out and stopped listening.
	select {
	case req.reply <- results:
	case <-req.ctx.Done():
	}
}

// Search sends a query to the worker and waits for the reply, the context
// deadline, or worker shutdown, whichever happens first. It is safe for
// concurrent use.
func (w *SearchWorker) Search(ctx context.Context, query SearchQuery) ([]SearchResult, error) {
	if w == nil {
		return nil, ErrWorkerClosed
	}

	req := searchRequest{
		ctx:   ctx,
		query: query,
		reply: make(chan []SearchResult, 1),
	}

	select {
	case w.requests <- req:
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-w.quit:
		return nil, ErrWorkerClosed
	}

	select {
	case results := <-req.reply:
		return results, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-w.quit:
		return nil, ErrWorkerClosed
	}
}

// Close stops the worker goroutine and blocks until it has exited. It is safe
// to call multiple times.
func (w *SearchWorker) Close() {
	if w == nil {
		return
	}
	w.closeOnce.Do(func() {
		close(w.quit)
	})
	<-w.done
}
