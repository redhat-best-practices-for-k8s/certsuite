package labels

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

// LabelsExprEvaluator evaluates label expressions.
//
// It defines a single method Eval that processes a label expression
// (typically a string) and returns the resulting set of labels as a map.
// The returned map may be empty if no labels match, and an error is
// provided if the expression cannot be parsed or evaluated.
type LabelsExprEvaluator interface {
	Eval(labels []string) bool
}

// labelsExprParser parses a label expression into an AST and evaluates it against a set of labels.
//
// It holds the root node of the parsed abstract syntax tree. The Eval method runs the
// expression on a slice of strings, returning true if the labels satisfy the expression.
type labelsExprParser struct {
	astRootNode ast.Expr
}

// NewLabelsExprEvaluator creates a LabelsExprEvaluator from an expression string.
//
// It parses the supplied string as a label selector expression, replacing any
// placeholder variables with their actual values before parsing.
// If parsing succeeds, it returns a LabelsExprEvaluator that can evaluate
// labels against the expression. Otherwise it returns an error explaining why
// the expression could not be parsed.
func NewLabelsExprEvaluator(labelsExpr string) (LabelsExprEvaluator, error) {
	goLikeExpr := strings.ReplaceAll(labelsExpr, "-", "_")
	goLikeExpr = strings.ReplaceAll(goLikeExpr, ",", "||")

	node, err := parser.ParseExpr(goLikeExpr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse labels expression %s: %v", labelsExpr, err)
	}

	return labelsExprParser{
		astRootNode: node,
	}, nil
}

// Eval evaluates a labels expression against the provided slice of label strings and returns true if the expression matches.
//
// The function takes a single parameter, a slice of strings representing labels,
// and parses the stored expression in the receiver to determine whether
// the labels satisfy the expression logic. It returns a boolean indicating
// success (true) or failure (false). If parsing errors occur during evaluation,
// they are handled internally and result in a false return value.
func (exprParser labelsExprParser) Eval(labels []string) bool {
	// Define a map for fast name/ident checking when visiting nodes.
	labelsMap := make(map[string]bool)
	for _, label := range labels {
		labelsMap[strings.ReplaceAll(label, "-", "_")] = true
	}

	// Visit function to walk the labels expression's AST.
	var visit func(e ast.Expr) bool
	visit = func(e ast.Expr) bool {
		switch v := e.(type) {
		case *ast.Ident:
			// If the expression is an identifier, check if it exists in the wordMap.
			if _, ok := labelsMap[v.Name]; !ok {
				return false
			}
			return true
		case *ast.ParenExpr:
			return visit(v.X)
		case *ast.UnaryExpr:
			if v.Op == token.NOT {
				return !visit(v.X)
			}
		case *ast.BinaryExpr:
			// If the expression is a binary expression, evaluate both operands.
			left := visit(v.X)
			right := visit(v.Y)
			switch v.Op {
			case token.LAND:
				return left && right
			case token.LOR:
				return left || right
			default:
				return false
			}
		default:
			log.Error("Unexpected/not-implemented expr: %v", v)
			return false
		}
		return false
	}

	return visit(exprParser.astRootNode)
}
