package parser

/**
walker package is not completed!!

The goal is walk the Terraform files and verify the value of each attribue.
**/

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/hashicorp/hcl2/hcl"
	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	log "github.com/sirupsen/logrus"
)

type Worker struct {
	Path      string
	Resource  string
	Attribute string
	Errors    []string
	Scorecard map[string]bool
	Body      *hclsyntax.Body // Worker needs to review this body
}

func NewWorker(task *Task, res string, path string) *Worker {
	return &Worker{
		Path:      path,
		Resource:  res,
		Attribute: task.Attribute,
		Scorecard: make(map[string]bool),
	}
}

func GenerateWorkers(body *hclsyntax.Body, tasks []*Task, path string) []*Worker {
	var workers []*Worker
	if len(body.Blocks) > 0 {
		for _, block := range body.Blocks {
			if block.Type == "resource" && len(block.Labels) > 0 {
				for _, task := range tasks {
					if block.Labels[0] == task.Resource {
						log.Infof("> Found %v %+v \n", task.Resource, strings.Join(block.Labels, " "))
						worker := NewWorker(
							task,
							strings.Join(block.Labels, " "),
							path,
						)
						worker.Scorecard[worker.Attribute] = false
						worker.generateScorecard(task.Body)
						workers = append(workers, worker)
					}
				}
			}
		}
	}

	return workers
}

func (w *Worker) generateScorecard(value interface{}) {
	switch v := reflect.ValueOf(value); v.Kind() {
	case reflect.Slice:
		for i := 0; i < v.Len(); i++ {
			w.generateScorecard(v.Index(i).Interface())
		}
	case reflect.Map:
		for _, key := range v.MapKeys() {
			w.generateScorecard(key.Interface())
			w.generateScorecard(v.MapIndex(key).Interface())
		}
	case reflect.String:
		w.Scorecard[v.String()] = false
	default:
		log.Warnf("tfs doesn't handle %s type", v.Kind())
	}
}

// ValidateScore check if attributes and keys exist
func (w *Worker) ValidateScore() {
	var err []string
	log.Infof("Score!: %+v\n", w.Scorecard)
	for key, value := range w.Scorecard {
		if !value {
			if key == w.Attribute {
				key = "attribute"
			}
			err = append(err, fmt.Sprintf("<%v %v> %v %v is missing ", w.Path,
				w.Resource, key, w.Attribute))
		}
	}

	w.Errors = err
}

// Verify goes through terraform file to look for attributes and keys
func (w *Worker) VerifyBody(body *hclsyntax.Body) {
	if len(body.Blocks) > 0 {
		for _, block := range body.Blocks {
			if _, ok := w.Scorecard[block.Type]; ok {
				log.Infof("> Found block %v\n", block.Type)
				w.Scorecard[block.Type] = true
			}
			w.VerifyBody(block.Body)
		}
	}

	if len(body.Attributes) > 0 {
		for _, attr := range body.Attributes {
			if _, ok := w.Scorecard[attr.Name]; ok {
				log.Infof("> Found attribue %v\n", attr.Name)
				w.Scorecard[attr.Name] = true
			}
			w.ExpressionWalk(attr.Expr)
		}
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
		//valueTypeWalk(t.Val)
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
		log.Warnf("Unknown expression type %v \n", reflect.TypeOf(t))
	}
}

func (w *Worker) traverseTypeWalk(v hcl.Traverser) {
	switch t := v.(type) {
	case hcl.TraverseRoot:
		if _, ok := w.Scorecard[t.Name]; ok {
			log.Infof("> Found Key %v", t.Name)
			w.Scorecard[t.Name] = true
		}
	case hcl.TraverseAttr:
		if _, ok := w.Scorecard[t.Name]; ok {
			log.Infof("> Found Key %v", t.Name)
			w.Scorecard[t.Name] = true
		}
	default:
		log.Warnf("Unknown trarverser type %v \n", reflect.TypeOf(t))
	}
}
