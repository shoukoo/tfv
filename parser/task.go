package parser

import (
	"fmt"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Task struct {
	Resource      string
	AttributeKeys map[string][]string
}

// PrepareTask creates new task struct
func PrepareTask(data []byte) ([]*Task, error) {
	var t map[string]map[string][]string
	var tasks []*Task
	err := yaml.Unmarshal([]byte(data), &t)
	if err != nil {
		return nil, err
	}
	for key, value := range t {
		for k, v := range value {
			var newTask Task
			newTask.Resource = key
			newTask.AttributeKeys = make(map[string][]string)
			newTask.AttributeKeys[k] = v
			tasks = append(tasks, &newTask)
			log.Infof("inside PrepareTask  %v \n", newTask)
		}

	}
	return tasks, nil

}

// GenerateTasks creates tasks based on the config
func GenerateTasks(b []byte) ([]*Task, error) {

	tasks, err := PrepareTask(b)
	if err != nil {
		log.Errorf("error preparing task %v", err)
		return nil, fmt.Errorf(
			`
Terraform Verifier only accepts this configuration format:
		
aws_resource:
	attributes:
	  - key1
	  - key2
	  - key3
		
example: if you want to check if see_algorithm exists in aws_s3_bucket resource

terraform.tf
server_side_encryption_configuration {
	rule {
	  apply_server_side_encryption_by_default {
		kms_master_key_id = "${aws_kms_key.mykey.arn}"
		sse_algorithm     = "aws:kms"
	  }
	}
}

configuration file:
aws_s3_bucket:
	server_side_encryption_configuration:
		- apply_server_side_encryption_by_default
		- kms_master_key_id
		- sse_algorithm
		`)
	}

	return tasks, nil
}
