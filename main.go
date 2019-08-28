package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/shoukoo/tf-verifier/walker"
	flags "github.com/simonleung8/flags"
	log "github.com/sirupsen/logrus"
)

var (
	debug  bool
	config string
	files  []string
)

func init() {
	f := flags.New()
	f.NewBoolFlag("debug", "d", "debug mode")
	f.NewStringFlagWithDefault("config", "c", "config file", "tf.yaml")
	f.Parse(os.Args...)

	debug = f.Bool("d")
	config = f.String("c")
	files = f.Args()[1:]

	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	if debug {
		log.SetLevel(log.DebugLevel)
	}
}

func main() {

	b, err := readConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	tasks, err := getTasks(b)
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		log.Fatal("List of terraform files not found")
	}

	var errs []string
	for _, f := range files {
		path := string(f)
		if _, err := os.Stat(path); err != nil {
			log.Fatalf("File does not exist: %s", path)
		}

		log.Infof("Examining: %s", path)
		p := hclparse.NewParser()
		file, d := p.ParseHCLFile(path)

		if d.HasErrors() {
			log.Fatalf("%v Error parsing %v", path, d.Error())
		}

		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			log.Fatalf("%v Error parsing %v", path, d.Error())
		}
		e := run(body, tasks, path)
		if len(e) > 0 {
			errs = append(errs, strings.Join(e, "\n"))
		}
	}

	if len(errs) > 0 {
		log.Fatalf("\n" + strings.Join(errs, "\n"))
	}
}

// readConfig to read config
func readConfig(config string) ([]byte, error) {
	b, err := ioutil.ReadFile(config)
	if err != nil {
		return nil, fmt.Errorf("Can't find config file %v", err)
	}

	return b, nil
}

// getTasks creates task based on the config
func getTasks(b []byte) ([]*walker.Task, error) {

	tasks, err := walker.PrepareTask(b)
	if err != nil {
		return nil, fmt.Errorf("Error preparing task %v", err)
	}

	return tasks, nil
}

// run to assign tasks to workers
func run(body *hclsyntax.Body, tasks []*walker.Task, path string) []string {
	var errStr []string
	var workers []*walker.Worker
	if len(body.Blocks) > 0 {
		for _, block := range body.Blocks {
			if block.Type == "resource" && len(block.Labels) > 0 {
				for _, w := range tasks {
					if block.Labels[0] == w.Resource {
						log.Infof("> Found %v %+v \n", w.Resource, strings.Join(block.Labels, " "))
						// Deploy worker
						worker := walker.NewWorker(
							strings.Join(block.Labels, " "),
							w.AttributeKeys,
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

// verify goes through terraform file to look for attributes and keys
func verify(b *hclsyntax.Body, w *walker.Worker) {
	log.Infof("*Worker* starts to verify %+v\n", w)
	if len(b.Blocks) > 0 {
		for _, block := range b.Blocks {
			if block.Type == w.Attribute {
				log.Infof("> Found block %v\n", block.Type)
				w.Scores[w.Attribute] = true
				verify(block.Body, w)
			}

		}

	}
	if len(b.Attributes) > 0 {
		for _, attr := range b.Attributes {
			if _, ok := w.Scores[attr.Name]; ok {
				log.Infof("> Found attribue %v\n", attr.Name)
				w.Scores[attr.Name] = true
				w.ExpressionWalk(attr.Expr)
			}
		}
	}
}
