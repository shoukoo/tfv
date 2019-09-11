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
	testingFiles = []string{"test/terraform.tf", "test/terraform12.tf"}
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

	files := testingFiles
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

	if err != nil {
		errorMsg(t, "to see some tasks", err)
	}

	if len(tasks) != 4 {
		errorMsg(t, "4 tasks", len(tasks))
	}

	b2, _ := readConfig("test/invalid/example.yaml")

	_, err = parser.GenerateTasks(b2)

	if err == nil {
		errorMsg(t, "to have tfv can't handle int in the configuration file error", err)
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
	files = testingFiles

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
		ws := parser.GenerateWorkers(b.Body, tasks, b.Path)
		workers = append(workers, ws...)
	}

	expect := &parser.Worker{
		Path:      "test/terraform12.tf",
		Resource:  "aws_instance main",
		Attribute: "tags",
		Scorecard: map[string]bool{
			"tags":      false,
			"Name":      false,
			"terraform": false,
		},
	}

	expect2 := &parser.Worker{
		Path:      "test/terraform12.tf",
		Resource:  "aws_s3_bucket main",
		Attribute: "server_side_encryption_configuration",
		Scorecard: map[string]bool{
			"rule": false,
			"apply_server_side_encryption_by_defaultapply_server_side_encryption_by_default": false,
			"kms_master_key_id": false,
		},
	}

	for _, w := range workers {
		if w.Path == expect.Path && w.Attribute == expect.Attribute && w.Resource == expect.Resource {
			for k := range expect.Scorecard {
				if _, ok := w.Scorecard[k]; !ok {
					errorMsg(t, fmt.Sprintf("can't find this key in the scorecard %+v", k), expect)
				}
			}
			return
		}
		if w.Path == expect2.Path && w.Attribute == expect2.Attribute && w.Resource == expect2.Resource {
			for k := range expect2.Scorecard {
				if _, ok := w.Scorecard[k]; !ok {
					errorMsg(t, fmt.Sprintf("can't find this key in the scorecard %+v", k), expect2)
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
		ws := parser.GenerateWorkers(b.Body, tasks, b.Path)
		for _, w := range ws {
			w.VerifyBody(b.Body)
		}
		workers = append(workers, ws...)
	}

	var errs []string
	for _, w := range workers {
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
