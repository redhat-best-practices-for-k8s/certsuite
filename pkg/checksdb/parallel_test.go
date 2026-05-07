package checksdb

import (
	"sync/atomic"
	"testing"

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
