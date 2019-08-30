package parser

import (
	"fmt"
	"os"

	"github.com/hashicorp/hcl2/hcl/hclsyntax"
	"github.com/hashicorp/hcl2/hclparse"
	log "github.com/sirupsen/logrus"
)

type Body struct {
	Path string
	Body *hclsyntax.Body
}

func GetBodies(files []string) ([]Body, error) {
	var bodies []Body
	for _, f := range files {
		path := string(f)
		if _, err := os.Stat(path); err != nil {
			return nil, fmt.Errorf("File does not exist: %s", path)
		}

		log.Infof("Examining: %s", path)
		p := hclparse.NewParser()
		file, d := p.ParseHCLFile(path)

		if d.HasErrors() {
			return nil, fmt.Errorf("%v Error parsing %v", path, d.Error())
		}

		body, ok := file.Body.(*hclsyntax.Body)
		if !ok {
			return nil, fmt.Errorf("%v Error parsing %v", path, d.Error())
		}
		bodies = append(bodies, Body{Path: path, Body: body})
	}

	return bodies, nil
}
