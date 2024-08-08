package labels

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/log"
)

type LabelsExprEvaluator interface {
	Eval(labels []string) bool
}

type labelsExprParser struct {
	astRootNode ast.Expr
}

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

// Evaluates the labels expression against the labels slice.
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
