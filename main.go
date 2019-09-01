package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/shoukoo/tf-verifier/parser"
	log "github.com/sirupsen/logrus"
	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	debug  = kingpin.Flag("debug", "Enable debug mode.").Bool()
	config = kingpin.Flag("config", "Configuration is a yaml file to tell tf-verfier what to check.").Default("tfv.yaml").String()
	files  = kingpin.Arg("files", "List of terraform files eg. t1.tf t2.tf").Strings()
)

func main() {

	// Don't parse kingpin in init func, it conflicts with go test flags
	kingpin.Version("0.0.2")
	kingpin.Parse()

	log.SetOutput(os.Stdout)
	log.SetLevel(log.WarnLevel)
	if *debug {
		log.SetLevel(log.DebugLevel)
	}

	// Read Config
	b, err := readConfig(*config)
	if err != nil {
		log.Fatal(err)
	}

	tasks, err := parser.GenerateTasks(b)
	if err != nil {
		log.Fatal(err)
	}

	if len(*files) == 0 {
		log.Fatal("List of terraform files not found")
	}

	bodies, err := parser.GetBodies(*files)
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
		return nil, fmt.Errorf("can't find config file %v", err)
	}

	return b, nil
}
