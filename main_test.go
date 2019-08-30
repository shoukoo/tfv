package main

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/shoukoo/tf-verifier/parser"
)

//Helpeor

var (
	_testingFiles = []string{"test/terraform.tf", "test/terraform12.tf"}
)

// errorMsg common error message format
func errorMsg(t *testing.T, expected string, result interface{}) {
	t.Errorf("Expecting %v but got %v", expected, result)
}

func prepareTest(t *testing.T) ([]*parser.Task, []parser.Body) {

	b, err := readConfig("test/example.yaml")

	if err != nil {
		errorMsg(t, "to find test/example.yaml", err)
	}

	tasks, err := parser.GenerateTasks(b)

	if err != nil {
		errorMsg(t, "error preparing tasks", len(tasks))
	}

	files := _testingFiles
	bodies, err := parser.GetBodies(files)
	if err != nil {
		errorMsg(t, "error preparing body", len(tasks))
	}

	return tasks, bodies
}

// Test begins here
func TestReadConfig(t *testing.T) {

	_, err := readConfig("invalid")

	if err == nil {
		errorMsg(t, "an error", err)
	}

	_, err = readConfig("test/example.yaml")

	if err != nil {
		errorMsg(t, "no error", err)
	}

}

func TestGetTask(t *testing.T) {

	b, err := readConfig("test/example.yaml")

	if err != nil {
		errorMsg(t, "to find test/example.yaml", err)
	}

	tasks, err := parser.GenerateTasks(b)

	if len(tasks) != 2 {
		errorMsg(t, "2 tasks", len(tasks))
	}

}

func TestGetBodies(t *testing.T) {

	// Process Invalid Terraform file
	files := []string{"test/invalid/terraform.tf"}
	_, err := parser.GetBodies(files)
	if err == nil {
		errorMsg(t, "an error", "nil")
	}

	// Process valid Terraform file
	files = _testingFiles

	bodies, err := parser.GetBodies(files)

	if err != nil {
		errorMsg(t, "to parse terraform files sucessfully", err)
	}

	// Make sure all the files are being processed
	for i, b := range bodies {
		if files[i] != b.Path {
			errorMsg(t, b.Path, files[i])
		}
	}

}

func TestGenerateWorkers(t *testing.T) {
	tasks, bodies := prepareTest(t)
	var workers []*parser.Worker
	for _, b := range bodies {
		w := parser.GenerateWorkers(b.Body, tasks, b.Path)
		workers = append(workers, w...)
	}

	expect := &parser.Worker{
		Path:      "test/terraform12.tf",
		Resource:  "aws_instance main",
		Attribute: "tags",
		Scores: map[string]bool{
			"tags":      false,
			"Name":      false,
			"terraform": false,
		},
	}

	for _, w := range workers {
		if w.Path == expect.Path && w.Attribute == expect.Attribute && w.Resource == expect.Resource {
			for k := range expect.Scores {
				if _, ok := w.Scores[k]; !ok {
					errorMsg(t, fmt.Sprintf("can't find this key in the score %+v", k), expect)
				}
			}
			return
		}
	}

	errorMsg(t, fmt.Sprintf("can't find this worker %+v", expect), "null")

}

func TestValidateScore(t *testing.T) {
	tasks, bodies := prepareTest(t)
	var workers []*parser.Worker
	for _, b := range bodies {
		w := parser.GenerateWorkers(b.Body, tasks, b.Path)
		workers = append(workers, w...)
	}

	var errs []string
	for _, w := range workers {
		w.VerifyBody()
		w.ValidateScore()
		if len(w.Errors) > 0 {
			errs = append(errs, w.Errors...)
		}
	}

	expect := []string{
		"<test/terraform12.tf aws_instance main> create_before_destroy lifecycle is missing",
		"<test/terraform12.tf aws_instance main> attribute lifecycle is missing",
	}
	sort.Strings(expect)
	sort.Strings(errs)

	for i, v := range errs {
		v = strings.Trim(v, " ")
		if v != expect[i] {
			errorMsg(t, v, expect[i])
		}
	}
}
