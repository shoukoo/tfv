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
	"github.com/zclconf/go-cty/cty"
)

func Walk(b *hclsyntax.Body) {

	if len(b.Blocks) > 0 {
		for _, block := range b.Blocks {
			fmt.Printf("%+v %+v\n", block.Type, block.Labels)
			Walk(block.Body)
		}

	}

	if len(b.Attributes) > 0 {
		for _, attr := range b.Attributes {
			fmt.Printf("%+v %+v\n", color.BlueString(attr.Name), reflect.TypeOf(attr.Expr))
			expressionWalk(attr.Expr)
		}
	}

}

func expressionWalk(ex hcl.Expression) {
	switch t := ex.(type) {
	case *hclsyntax.TemplateExpr:
		for _, p := range t.Parts {
			expressionWalk(p)
		}
	case *hclsyntax.TemplateWrapExpr:
		expressionWalk(t.Wrapped)

	case *hclsyntax.LiteralValueExpr:
		valueTypeWalk(t.Val)

	case *hclsyntax.TupleConsExpr:
		for _, v := range t.Exprs {
			expressionWalk(v)
		}
	case *hclsyntax.ScopeTraversalExpr:
		for _, v := range t.Traversal {
			traverseTypeWalk(v)
		}
	case *hclsyntax.ObjectConsExpr:
		for _, v := range t.Items {
			expressionWalk(v.KeyExpr)
			expressionWalk(v.ValueExpr)
		}
	case *hclsyntax.ObjectConsKeyExpr:
		expressionWalk(t.Wrapped)
	case *hclsyntax.FunctionCallExpr:
		for _, e := range t.Args {
			expressionWalk(e)
		}
	case *hclsyntax.ConditionalExpr:
		fmt.Printf("%v \n", color.YellowString("condition exp"))
		expressionWalk(t.Condition)
		expressionWalk(t.TrueResult)
		expressionWalk(t.FalseResult)
	case *hclsyntax.BinaryOpExpr:
		// Operations
		fmt.Printf("%v \n", color.YellowString("binary exp"))
		expressionWalk(t.LHS)
		expressionWalk(t.RHS)
	case *hclsyntax.SplatExpr:
		// Operations
		fmt.Printf("%v \n", color.YellowString("splat exp"))
		expressionWalk(t.Source)
		expressionWalk(t.Each)
	case *hclsyntax.RelativeTraversalExpr:
		fmt.Printf("%v \n", color.YellowString("relative traversal exp"))
		expressionWalk(t.Source)
		for _, v := range t.Traversal {
			traverseTypeWalk(v)
		}
	case *hclsyntax.AnonSymbolExpr:
		fmt.Printf("%v \n", color.YellowString("don't know how to handle anon symbol exp"))
	default:
		fmt.Printf(color.RedString("unknown expression%v\n"), reflect.TypeOf(t))
	}
}

func traverseTypeWalk(v hcl.Traverser) {
	switch t := v.(type) {
	case hcl.TraverseRoot:
		fmt.Printf("traverse root %v\n", t.Name)
	case hcl.TraverseAttr:
		fmt.Printf("traverse attr %v\n", t.Name)
	case hcl.TraverseIndex:
		valueTypeWalk(t.Key)
	default:
		fmt.Printf(color.RedString("unknow tarverser type %v \n"), reflect.TypeOf(t))
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
