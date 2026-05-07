package checksdb

import (
	"fmt"
	"sync"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"golang.org/x/sync/errgroup"
)

const DefaultParallelLimit = 10

type ParallelResult struct {
	mu                  sync.Mutex
	compliantObjects    []*testhelper.ReportObject
	nonCompliantObjects []*testhelper.ReportObject
}

func (r *ParallelResult) AddCompliantObject(obj *testhelper.ReportObject) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.compliantObjects = append(r.compliantObjects, obj)
}

func (r *ParallelResult) AddNonCompliantObject(obj *testhelper.ReportObject) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nonCompliantObjects = append(r.nonCompliantObjects, obj)
}

func (r *ParallelResult) AddCompliantObjects(objs []*testhelper.ReportObject) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.compliantObjects = append(r.compliantObjects, objs...)
}

func (r *ParallelResult) AddNonCompliantObjects(objs []*testhelper.ReportObject) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nonCompliantObjects = append(r.nonCompliantObjects, objs...)
}

func (r *ParallelResult) Results() (compliant, nonCompliant []*testhelper.ReportObject) {
	return r.compliantObjects, r.nonCompliantObjects
}

// ForEachParallel iterates over items concurrently, calling fn for each.
// Concurrency is bounded by limit (0 uses DefaultParallelLimit).
// Results are collected via the thread-safe ParallelResult and passed to check.SetResult after completion.
func ForEachParallel[T any](check *Check, items []T, limit int, fn func(*Check, T, *ParallelResult)) {
	if limit <= 0 {
		limit = DefaultParallelLimit
	}

	result := &ParallelResult{}
	g := new(errgroup.Group)
	g.SetLimit(limit)

	for _, item := range items {
		g.Go(func() error {
			defer func() {
				if r := recover(); r != nil {
					check.LogError("Panic during parallel check execution: %v", r)
					check.SetResultError(fmt.Sprintf("panic: %v", r))
				}
			}()
			fn(check, item, result)
			return nil
		})
	}

	_ = g.Wait()
	check.SetResult(result.Results())
}
