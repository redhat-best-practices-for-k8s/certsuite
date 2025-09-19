package labels

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

// LabelsExprEvaluator Evaluates label sets for compliance
//
// The evaluator takes an array of strings representing labels and returns true
// if they satisfy the underlying expression rules, otherwise false. It
// encapsulates the logic needed to determine whether a given set of labels
// matches the expected pattern or condition defined by the system.
type LabelsExprEvaluator interface {
	Eval(labels []string) bool
}

// labelsExprParser Parses and evaluates label expressions against a list of labels
//
// It walks the abstract syntax tree of an expression, checking identifiers
// against provided labels, handling parentheses, logical NOT, AND, OR
// operators, and reporting unexpected nodes. The result is true if the
// expression matches the label set, otherwise false.
type labelsExprParser struct {
	astRootNode ast.Expr
}

// NewLabelsExprEvaluator Creates an evaluator that checks label expressions
//
// The function transforms a comma-separated string of labels into a
// Go-compatible boolean expression, parses it into an abstract syntax tree, and
// returns an evaluator object. It replaces hyphens with underscores and commas
// with logical OR operators before parsing. If the input cannot be parsed, an
// error is returned.
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

// labelsExprParser.Eval Evaluates a logical expression against a set of labels
//
// This method builds a lookup map from the supplied label strings, normalizing
// dashes to underscores for matching. It then recursively traverses an abstract
// syntax tree representing the expression, evaluating identifiers, parentheses,
// unary NOT, and binary AND/OR operators using the lookup map. The result is a
// boolean indicating whether the labels satisfy the expression.
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
