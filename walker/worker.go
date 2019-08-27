package walker

/**
walker package is not completed!!

The goal is walk the Terraform files and verify the value of each attribue.
**/

import (
	"fmt"
	"reflect"

	"github.com/fatih/color"
	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	log "github.com/sirupsen/logrus"
	"github.com/zclconf/go-cty/cty"
)

type Worker struct {
	Path     string
	Resource string
	Attribute string
	Errors   []string
	Scores   map[string]bool
}

func NewWorker(res string, attributes map[string][]string, path string) *Worker {
	score := make(map[string]bool)
	var att string

	for attr, keys := range attributes {
		score[attr] = false
		att = attr
		for _, i := range keys {
			score[i] = false
		}
	}

	return &Worker{
		Path:     path,
		Attribute: att,
		Resource: res,
		Scores:   score,
	}
}

// ExpressionWalk to get the key of the attribute
func (w *Worker) ExpressionWalk(ex hcl.Expression) {
	switch t := ex.(type) {
	case *hclsyntax.TemplateExpr:
		for _, p := range t.Parts {
			w.ExpressionWalk(p)
		}
	case *hclsyntax.TemplateWrapExpr:
		w.ExpressionWalk(t.Wrapped)
	case *hclsyntax.LiteralValueExpr:
		valueTypeWalk(t.Val)
	case *hclsyntax.TupleConsExpr:
		for _, v := range t.Exprs {
			w.ExpressionWalk(v)
		}
	case *hclsyntax.ScopeTraversalExpr:
		for _, v := range t.Traversal {
			w.traverseTypeWalk(v)
		}
	case *hclsyntax.ObjectConsExpr:
		for _, v := range t.Items {
			w.ExpressionWalk(v.KeyExpr)
		}
	case *hclsyntax.ObjectConsKeyExpr:
		w.ExpressionWalk(t.Wrapped)
	case *hclsyntax.FunctionCallExpr:
		for _, e := range t.Args {
			w.ExpressionWalk(e)
		}
	default:
		log.Errorf("Unknown expression type %v \n", reflect.TypeOf(t))
	}
}

func (w *Worker) ValidateScore() {
	var err []string
	log.Infof("Score!: %+v\n", w.Scores)
	for key, value := range w.Scores {
		if !value {
			err = append(err, fmt.Sprintf("<%v %v> %v %v is missing ", w.Path,
				w.Resource, key, w.Attribute))
		}
	}

	w.Errors = err
}

func (w *Worker) traverseTypeWalk(v hcl.Traverser) {
	switch t := v.(type) {
	case hcl.TraverseRoot:
		if _, ok := w.Scores[t.Name]; ok {
			log.Infof("> Found Key %v", t.Name)
			w.Scores[t.Name] = true
		}
	default:
		log.Errorf("Unknown tarverser type %v \n", reflect.TypeOf(t))
	}
}

func valueTypeWalk(t cty.Value) {
	switch t.Type() {
	case cty.String:
		fmt.Printf("string type %v \n", t.AsString())
	case cty.Number:
		fmt.Printf("number type %v \n", t.AsBigFloat())
	case cty.Bool:
		fmt.Printf("boolean type %v \n", t.True())
	case cty.EmptyObject:
		fmt.Printf("empty object type %v \n", t)
	case cty.DynamicPseudoType:
		fmt.Printf("empty dynamic pseudo type %v \n", t)
	case cty.EmptyTuple:
		fmt.Printf("empty tuple type %v \n", t)
	default:
		fmt.Printf(color.RedString("unknown value type %v\n"), reflect.TypeOf(t))
	}
}

/**
 Keep below for future reference
**/
//func (w *Worker) SecondVersionExpressionWalk(ex hcl.Expression) {
//	switch t := ex.(type) {
//	case *hclsyntax.TemplateExpr:
//		for _, p := range t.Parts {
//			w.ExpressionWalk(p)
//		}
//	case *hclsyntax.TemplateWrapExpr:
//		w.ExpressionWalk(t.Wrapped)
//
//	case *hclsyntax.LiteralValueExpr:
//		valueTypeWalk(t.Val)
//	case *hclsyntax.TupleConsExpr:
//		for _, v := range t.Exprs {
//			w.ExpressionWalk(v)
//		}
//	case *hclsyntax.ScopeTraversalExpr:
//		for _, v := range t.Traversal {
//			traverseTypeWalk(v)
//		}
//	case *hclsyntax.ObjectConsExpr:
//		for _, v := range t.Items {
//			w.ExpressionWalk(v.KeyExpr)
//			w.ExpressionWalk(v.ValueExpr)
//		}
//	case *hclsyntax.ObjectConsKeyExpr:
//		w.ExpressionWalk(t.Wrapped)
//	case *hclsyntax.FunctionCallExpr:
//		for _, e := range t.Args {
//			w.ExpressionWalk(e)
//		}
//	case *hclsyntax.ConditionalExpr:
//		w.ExpressionWalk(t.Condition)
//		w.ExpressionWalk(t.TrueResult)
//		w.ExpressionWalk(t.FalseResult)
//	case *hclsyntax.BinaryOpExpr:
//		w.ExpressionWalk(t.LHS)
//		w.ExpressionWalk(t.RHS)
//	case *hclsyntax.SplatExpr:
//		w.ExpressionWalk(t.Source)
//		w.ExpressionWalk(t.Each)
//	case *hclsyntax.RelativeTraversalExpr:
//		w.ExpressionWalk(t.Source)
//		for _, v := range t.Traversal {
//			traverseTypeWalk(v)
//		}
//	case *hclsyntax.AnonSymbolExpr:
//		fmt.Printf("%v \n", color.YellowString("don't know how to handle anon symbol exp"))
//	default:
//		fmt.Printf(color.RedString("unknown expression%v\n"), reflect.TypeOf(t))
//	}
//}

//func traverseTypeWalk(v hcl.Traverser) {
//	switch t := v.(type) {
//	case hcl.TraverseRoot:
//		fmt.Printf("traverse root %v\n", t.Name)
//	case hcl.TraverseAttr:
//		fmt.Printf("traverse attr %v\n", t.Name)
//	case hcl.TraverseIndex:
//		valueTypeWalk(t.Key)
//	default:
//		fmt.Printf(color.RedString("unknow tarverser type %v \n"), reflect.TypeOf(t))
//	}
//}
