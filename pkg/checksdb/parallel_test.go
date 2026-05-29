package checksdb

import (
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/redhat-best-practices-for-k8s/certsuite/pkg/testhelper"
	"github.com/stretchr/testify/assert"
)

func TestParallelResultThreadSafety(t *testing.T) {
	r := &ParallelResult{}

	const n = 100
	done := make(chan struct{})

	for range n {
		go func() {
			r.AddCompliantObject(testhelper.NewPodReportObject("ns", "pod", "ok", true))
			r.AddNonCompliantObject(testhelper.NewPodReportObject("ns", "pod", "fail", false))
			done <- struct{}{}
		}()
	}

	for range n {
		<-done
	}

	compliant, nonCompliant := r.Results()
	assert.Len(t, compliant, n)
	assert.Len(t, nonCompliant, n)
}

func TestForEachParallel(t *testing.T) {
	check := NewCheck("test-parallel", []string{"test"})

	items := []string{"a", "b", "c", "d", "e"}

	ForEachParallel(check, items, 0, func(c *Check, item string, r *ParallelResult) {
		if item == "c" {
			c.LogError("item %s is non-compliant", item)
			r.AddNonCompliantObject(testhelper.NewPodReportObject("ns", item, "bad", false))
		} else {
			r.AddCompliantObject(testhelper.NewPodReportObject("ns", item, "ok", true))
		}
	})

	assert.Equal(t, "failed", check.Result.String())
}

func TestForEachParallelAllCompliant(t *testing.T) {
	check := NewCheck("test-parallel-pass", []string{"test"})

	items := []int{1, 2, 3}

	ForEachParallel(check, items, 2, func(c *Check, item int, r *ParallelResult) {
		r.AddCompliantObject(testhelper.NewPodReportObject("ns", "pod", "ok", true))
	})

	assert.Equal(t, "passed", check.Result.String())
}

func TestForEachParallelEmpty(t *testing.T) {
	check := NewCheck("test-parallel-empty", []string{"test"})

	ForEachParallel(check, []string{}, 0, func(c *Check, item string, r *ParallelResult) {
		t.Fatal("should not be called")
	})

	assert.Equal(t, "skipped", check.Result.String())
}

func TestForEachParallelRespectsLimit(t *testing.T) {
	check := NewCheck("test-parallel-limit", []string{"test"})

	var active atomic.Int32
	var maxActive atomic.Int32
	items := make([]int, 50)
	for i := range items {
		items[i] = i
	}

	ForEachParallel(check, items, 5, func(c *Check, _ int, r *ParallelResult) {
		cur := active.Add(1)
		for {
			old := maxActive.Load()
			if cur <= old || maxActive.CompareAndSwap(old, cur) {
				break
			}
		}
		r.AddCompliantObject(testhelper.NewPodReportObject("ns", "pod", "ok", true))
		active.Add(-1)
	})

	assert.LessOrEqual(t, int(maxActive.Load()), 5)
}

// TestForEachParallelPerNodeMutex verifies the per-node mutex pattern used by
// probe-pod-heavy tests. Items are assigned to "nodes"; a per-node mutex
// serializes execution within each node while allowing cross-node parallelism.
// The test asserts that per-node concurrency never exceeds 1, while total
// concurrency across nodes does exceed 1 (proving parallelism isn't lost).
func TestForEachParallelPerNodeMutex(t *testing.T) {
	type item struct {
		name string
		node string
	}

	// 3 nodes, 10 items per node = 30 items total.
	nodes := []string{"node-a", "node-b", "node-c"}
	var items []item
	for _, node := range nodes {
		for i := range 10 {
			items = append(items, item{name: fmt.Sprintf("%s-pod-%d", node, i), node: node})
		}
	}

	// Build per-node mutex map (same pattern as production code).
	mutexPerNode := make(map[string]*sync.Mutex, len(nodes))
	for _, node := range nodes {
		mutexPerNode[node] = &sync.Mutex{}
	}

	// Track per-node and total concurrency.
	perNodeActive := make(map[string]*atomic.Int32, len(nodes))
	perNodeMaxActive := make(map[string]*atomic.Int32, len(nodes))
	for _, node := range nodes {
		perNodeActive[node] = &atomic.Int32{}
		perNodeMaxActive[node] = &atomic.Int32{}
	}
	var totalMaxActive atomic.Int32
	var totalActive atomic.Int32

	check := NewCheck("test-per-node-mutex", []string{"test"})

	ForEachParallel(check, items, len(nodes), func(c *Check, it item, r *ParallelResult) {
		if nodeMutex, ok := mutexPerNode[it.node]; ok {
			nodeMutex.Lock()
			defer nodeMutex.Unlock()
		}

		// Record per-node concurrency.
		nodeCount := perNodeActive[it.node].Add(1)
		for {
			old := perNodeMaxActive[it.node].Load()
			if nodeCount <= old || perNodeMaxActive[it.node].CompareAndSwap(old, nodeCount) {
				break
			}
		}

		// Record total concurrency.
		totalCount := totalActive.Add(1)
		for {
			old := totalMaxActive.Load()
			if totalCount <= old || totalMaxActive.CompareAndSwap(old, totalCount) {
				break
			}
		}

		// Simulate probe pod work so goroutines overlap.
		time.Sleep(5 * time.Millisecond)

		r.AddCompliantObject(testhelper.NewPodReportObject("ns", it.name, "ok", true))
		perNodeActive[it.node].Add(-1)
		totalActive.Add(-1)
	})

	// Per-node concurrency must never exceed 1.
	for _, node := range nodes {
		assert.Equal(t, int32(1), perNodeMaxActive[node].Load(),
			"node %s had concurrent execution — per-node mutex failed", node)
	}

	// Total concurrency should exceed 1, proving cross-node parallelism works.
	assert.Greater(t, int(totalMaxActive.Load()), 1,
		"cross-node parallelism was not achieved")

	assert.Equal(t, "passed", check.Result.String())
}
