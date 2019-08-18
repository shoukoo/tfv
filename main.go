package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclparse"
	"github.com/shoukoo/tf-verifier/walker"
)

type terraform struct {
	file   string
	errors []string
}

func main() {

	var files []string

	pwd, err := os.Getwd()
	if err != nil {
		fmt.Errorf("Error getting current directory%e", err)
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

	if err != nil {
		fmt.Errorf("Error walking through current directory %e", err)
	}

	for _, path := range files {
		p := hclparse.NewParser()
		file, d := p.ParseHCLFile(path)

		if d.HasErrors() {
			log.Fatalf("%v Error hcl parsing %v", path, d.Error())
		}

		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			log.Fatalf("%v Error hcl parsing %v", path, d.Error())
		}

		walker.Walk(body)

	}

}
