package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//Helpeor

// errorMsg common error message format
func errorMsg(t *testing.T, expected string, result interface{}) {
	t.Errorf("Expecting %v but got %v", expected, result)
}

// getTerraformFiles to get all terraform files from test dir
func getTerraformFiles(t *testing.T) []string {
	var files []string
	err := filepath.Walk("test/", func(path string, info os.FileInfo, err error) error {

		// Ignore hidden files e.g. .terraform
		if info.IsDir() && strings.HasPrefix(info.Name(), ".") {
			return filepath.SkipDir
		}

		// Only includes terraform file
		if filepath.Ext(info.Name()) == ".tf" {
			files = append(files, path)
		}

		return nil
	})

	if err != nil {
		t.Errorf("Unable to get the test files, reason: %v", err)
	}

	return files
}

// TestReadConfig
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

// TestGetTask
func TestGetTask(t *testing.T) {

	b, err := readConfig("test/example.yaml")

	if err != nil {
		errorMsg(t, "to find test/example.yaml", err)
	}

	tasks, err := getTasks(b)

	if len(tasks) != 2 {
		errorMsg(t, "2 tasks", len(tasks))
	}

}

// TestGetHCLBodies
func TestGetHCLBodies(t *testing.T) {
	files := getTerraformFiles(t)

	bodies, err := getHCLBodies(files)

	if err != nil {
		errorMsg(t, "to parse terraform files sucessfully", err)
	}

	if len(bodies) != 2 {
		errorMsg(t, "2 hclBody", len(bodies))
	}
}
