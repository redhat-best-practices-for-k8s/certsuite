package labels

import (
	"testing"
)

// FuzzNewLabelsExprEvaluator feeds arbitrary text to the --label-filter parser.
// The expression comes straight from user input, so the parser and the AST
// walk behind Eval must never panic, whatever the caller passes in.
func FuzzNewLabelsExprEvaluator(f *testing.F) {
	for _, seed := range []string{
		"",
		"all",
		"networking && lifecycle",
		"networking,lifecycle",
		"!(performance) && access-control",
		"((a) && (b)) || !(c)",
		"a && b && c && d",
		"!!!a",
		"a &&",
		"&& a",
		"((",
		"a || b || (c && !d)",
		"a.b",
		"a[0]",
	} {
		f.Add(seed)
	}

	labelSets := [][]string{
		nil,
		{},
		{"all"},
		{"networking", "lifecycle"},
		{"a", "b", "c", "d"},
	}

	f.Fuzz(func(t *testing.T, expr string) {
		evaluator, err := NewLabelsExprEvaluator(expr)
		if err != nil {
			// A rejected expression is a valid outcome; the only requirement
			// is that the failure is reported, not a panic.
			return
		}

		for _, labels := range labelSets {
			// The result is deliberately not asserted: this target is looking
			// for panics and unbounded recursion, not for semantics.
			_ = evaluator.Eval(labels)
		}
	})
}
