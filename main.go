package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/shoukoo/tf-verifier/walker"

	log "github.com/sirupsen/logrus"
)

func init() {
	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.WarnLevel)
}

func main() {

	var files []string

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting current directory%e", err)
	}

	// Walk through all files and subdirectories
	err = filepath.Walk(pwd, func(path string, info os.FileInfo, err error) error {

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

	if err != nil && len(files) == 0 {
		log.Fatalf("Error walking through current directory or cannot find any terraform file %e", err)
	}

	// Prepare tasks
	tasks := []*walker.Task{}
	att := map[string][]string{
		"tags": []string{"Name", "terraform"},
	}
	tasks = append(tasks, walker.NewTask("aws_instance", att))

	var errStr []string
	for _, path := range files {
		p := hclparse.NewParser()
		file, d := p.ParseHCLFile(path)

		if d.HasErrors() {
			log.Fatalf("%v Error hcl parsing %v", path, d.Error())
		}

		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			log.Fatalf("%v Error parsing hck body %v", path, d.Error())
		}

		errStr = append(errStr, run(body, tasks, path)...)

	}

	if len(errStr) > 0 {
		for _, e := range errStr {
			log.Error(e)
		}
		log.Fatal("failed")
	}

	fmt.Println("succeed!")
}

func run(body *hclsyntax.Body, tasks []*walker.Task, path string) []string {
	var errStr []string
	var workers []*walker.Worker
	if len(body.Blocks) > 0 {
		for _, block := range body.Blocks {
			if block.Type == "resource" && len(block.Labels) > 0 {
				for _, w := range tasks {
					if block.Labels[0] == w.Resource {
						log.Infof("Found %v %+v \n", w.Resource, strings.Join(block.Labels, " "))
						// Deploy worker
						worker := walker.NewWorker(
							strings.Join(block.Labels, " "),
							w.AttibuteKeys,
							path,
						)
						workers = append(workers, worker)
						verify(block.Body, worker)
						worker.ValidateScore()
					}
				}
			}
		}
	}

	for _, w := range workers {
		for _, e := range w.Errors {
			errStr = append(errStr, e)
		}
	}

	return errStr
}

func verify(b *hclsyntax.Body, w *walker.Worker) {
	if len(b.Attributes) > 0 {
		for _, attr := range b.Attributes {
			if w.Attibute == attr.Name {
				log.Infof("Found attribue %v\n", attr.Name)
				w.Scores[w.Attibute] = true
				w.ExpressionWalk(attr.Expr)
			}
		}
	}
}
