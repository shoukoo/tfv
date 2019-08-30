package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/shoukoo/tf-verifier/parser"
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

	tasks, err := parser.GenerateTasks(b)
	if err != nil {
		log.Fatal(err)
	}

	if len(files) == 0 {
		log.Fatal("List of terraform files not found")
	}

	bodies, err := parser.GetBodies(files)
	if err != nil {
		log.Fatal(err)
	}

	var workers []*parser.Worker
	for _, b := range bodies {
		ws := parser.GenerateWorkers(b.Body, tasks, b.Path)
		workers = append(workers, ws...)
	}

	var errs []string
	for _, w := range workers {
		w.VerifyBody()
		w.ValidateScore()
		if len(w.Errors) > 0 {
			errs = append(errs, strings.Join(w.Errors, "\n"))
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
