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
		return nil, fmt.Errorf("Error preparing task %v", err)
	}

	return tasks, nil
}
